package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/xyzbit/ino/internal/domain/repository"
)

type vectorRepository struct {
	client client.Client
}

// NewVectorRepository 创建向量仓储实例
func NewVectorRepository(client client.Client) (repository.VectorRepository, func() error) {
	return &vectorRepository{client: client}, func() error {
		return client.Close()
	}
}

// CreateCollection 创建集合
func (r *vectorRepository) CreateCollection(ctx context.Context, collectionName string, dimension int) error {
	// 检查集合是否已存在
	exists, err := r.client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if exists {
		return fmt.Errorf("collection %s already exists", collectionName)
	}

	// 定义集合schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    fmt.Sprintf("Vector collection for %s", collectionName),
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{
					"max_length": "64",
				},
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dimension),
				},
			},
			{
				Name:     "metadata",
				DataType: entity.FieldTypeJSON,
			},
		},
	}

	// 创建集合
	err = r.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	log.Printf("Collection %s created successfully", collectionName)
	return nil
}

// DropCollection 删除集合
func (r *vectorRepository) DropCollection(ctx context.Context, collectionName string) error {
	return r.client.DropCollection(ctx, collectionName)
}

// HasCollection 检查集合是否存在
func (r *vectorRepository) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	return r.client.HasCollection(ctx, collectionName)
}

// Insert 插入向量数据
func (r *vectorRepository) Insert(ctx context.Context, collectionName string, vectors []repository.VectorData) error {
	if len(vectors) == 0 {
		return nil
	}

	// 准备数据
	ids := make([]string, len(vectors))
	vectorData := make([][]float32, len(vectors))
	metadata := make([][]byte, len(vectors))

	for i, v := range vectors {
		ids[i] = v.ID
		vectorData[i] = v.Vector

		// 序列化metadata为JSON
		metadataBytes, err := json.Marshal(v.Metadata)
		if err != nil {
			return fmt.Errorf("failed to serialize metadata: %w", err)
		}
		metadata[i] = metadataBytes
	}

	// 构建列数据
	columns := []entity.Column{
		entity.NewColumnVarChar("id", ids),
		entity.NewColumnFloatVector("vector", len(vectorData[0]), vectorData),
		entity.NewColumnJSONBytes("metadata", metadata),
	}

	// 插入数据
	_, err := r.client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	// 刷新数据
	err = r.client.Flush(ctx, collectionName, false)
	if err != nil {
		log.Printf("Warning: failed to flush collection %s: %v", collectionName, err)
	}

	return nil
}

// Delete 删除向量数据
func (r *vectorRepository) Delete(ctx context.Context, collectionName string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建删除表达式
	expr := fmt.Sprintf("id in [%s]", buildStringList(ids))

	err := r.client.Delete(ctx, collectionName, "", expr)
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	return nil
}

// Update 更新向量数据 (Milvus通过删除+插入实现)
func (r *vectorRepository) Update(ctx context.Context, collectionName string, vectors []repository.VectorData) error {
	if len(vectors) == 0 {
		return nil
	}

	// 先删除现有数据
	ids := make([]string, len(vectors))
	for i, v := range vectors {
		ids[i] = v.ID
	}

	err := r.Delete(ctx, collectionName, ids)
	if err != nil {
		return fmt.Errorf("failed to delete existing vectors: %w", err)
	}

	// 插入新数据
	return r.Insert(ctx, collectionName, vectors)
}

// Search 搜索向量
func (r *vectorRepository) Search(ctx context.Context, collectionName string, vectors [][]float32, topK int, params map[string]interface{}) ([]repository.VectorSearchResult, error) {
	// 加载集合
	err := r.client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	// 构建搜索参数
	searchParams, err := entity.NewIndexHNSWSearchParam(40) // ef = 40
	if err != nil {
		return nil, fmt.Errorf("failed to create search params: %w", err)
	}
	if params != nil {
		if ef, ok := params["ef"].(int); ok {
			searchParams, err = entity.NewIndexHNSWSearchParam(ef)
			if err != nil {
				return nil, fmt.Errorf("failed to create search params: %w", err)
			}
		}
	}

	// 转换向量数据
	vectorEntities := make([]entity.Vector, len(vectors))
	for i, v := range vectors {
		vectorEntities[i] = entity.FloatVector(v)
	}

	// 执行搜索
	results, err := r.client.Search(
		ctx,
		collectionName,
		[]string{},
		"",
		[]string{"id", "metadata"},
		vectorEntities,
		"vector",
		entity.IP, // 内积距离
		topK,
		searchParams,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 转换结果
	var searchResults []repository.VectorSearchResult
	for _, result := range results {
		for i := 0; i < result.ResultCount; i++ {
			id, _ := result.IDs.Get(i)
			score := result.Scores[i]

			// 获取元数据
			var metadata map[string]interface{}
			if result.Fields.GetColumn("metadata") != nil {
				metadataCol := result.Fields.GetColumn("metadata").(*entity.ColumnJSONBytes)
				if i < metadataCol.Len() {
					rawData, _ := metadataCol.Get(i)
					if rawBytes, ok := rawData.([]byte); ok {
						json.Unmarshal(rawBytes, &metadata)
					}
				}
			}

			searchResults = append(searchResults, repository.VectorSearchResult{
				ID:       fmt.Sprintf("%v", id),
				Score:    float64(score),
				Metadata: metadata,
			})
		}
	}

	return searchResults, nil
}

// CreateIndex 创建索引
func (r *vectorRepository) CreateIndex(ctx context.Context, collectionName string, params map[string]interface{}) error {
	// 默认使用HNSW索引
	indexType := entity.HNSW

	if params != nil {
		if it, ok := params["index_type"].(string); ok {
			switch it {
			case "HNSW":
				indexType = entity.HNSW
			case "IVF_FLAT":
				indexType = entity.IvfFlat
			case "IVF_PQ":
				indexType = entity.IvfPQ
			case "FLAT":
				indexType = entity.Flat
			}
		}
	}

	// 构建索引参数
	var indexParams map[string]string
	switch indexType {
	case entity.HNSW:
		indexParams = map[string]string{
			"M":              "16",
			"efConstruction": "200",
		}
	case entity.IvfFlat:
		indexParams = map[string]string{
			"nlist": "1024",
		}
	case entity.IvfPQ:
		indexParams = map[string]string{
			"nlist": "1024",
			"m":     "8",
		}
	default:
		indexParams = map[string]string{}
	}

	// 覆盖默认参数
	if params != nil {
		for k, v := range params {
			if k != "index_type" && k != "metric_type" {
				if strVal, ok := v.(string); ok {
					indexParams[k] = strVal
				}
			}
		}
	}

	index := entity.NewGenericIndex("vector", indexType, indexParams)

	err := r.client.CreateIndex(ctx, collectionName, "vector", index, false)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	log.Printf("Index created for collection %s", collectionName)
	return nil
}

// DropIndex 删除索引
func (r *vectorRepository) DropIndex(ctx context.Context, collectionName string) error {
	return r.client.DropIndex(ctx, collectionName, "vector")
}

// GetCollectionStats 获取集合统计信息
func (r *vectorRepository) GetCollectionStats(ctx context.Context, collectionName string) (*repository.VectorCollectionStats, error) {
	stats, err := r.client.GetCollectionStatistics(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	// 解析统计信息
	rowCount := int64(0)
	if countStr, ok := stats["row_count"]; ok {
		if parsed, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			rowCount = parsed
		}
	}

	return &repository.VectorCollectionStats{
		RowCount:     rowCount,
		IndexedCount: rowCount, // Milvus没有单独的indexed_count
		MemorySize:   0,        // 需要通过其他API获取
		DiskSize:     0,        // 需要通过其他API获取
	}, nil
}

// 辅助函数
func buildStringList(ids []string) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("\"%s\"", id)
	}
	return result
}
