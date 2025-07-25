package neo4j

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const retrieveTyp = "neo4j_retriever"

// RetrieverConfig Neo4j检索器配置
type RetrieverConfig struct {
	// Driver is the Neo4j driver to be called
	// Required
	Driver neo4j.DriverWithContext

	// Database is the database name in Neo4j
	// Optional, default is "neo4j"
	Database string

	// EmbeddingModel is the model used to generate embeddings for queries
	// Required
	EmbeddingModel embedding.Embedder

	// Dimension is the dimension of the embedding
	// Optional, default is 2048
	Dimension int

	// SimilarityThreshold is the threshold for similarity search
	// Optional, default is 0.7
	SimilarityThreshold float64

	// MaxDepth is the maximum depth for graph traversal
	// Optional, default is 2
	MaxDepth int

	// TopK is the default number of results to return
	// Optional, default is 10
	TopK int

	// Limit is the maximum number of results to return
	// Optional, default is 50
	Limit int
}

// Retriever Neo4j图检索器
type Retriever struct {
	config RetrieverConfig
}

// NewRetriever 创建新的Neo4j检索器
func NewRetriever(ctx context.Context, conf *RetrieverConfig) (*Retriever, error) {
	if err := conf.check(); err != nil {
		return nil, err
	}

	// Test Neo4j connection
	if err := conf.Driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("[NewRetriever] Neo4j connection failed: %w", err)
	}

	return &Retriever{
		config: *conf,
	}, nil
}

// Retrieve 实现eino Retriever接口的检索方法
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) (docs []*schema.Document, err error) {
	// 获取检索选项
	options := retriever.GetImplSpecificOptions(&RetrieverImplOptions{
		TopK:                r.config.TopK,
		SimilarityThreshold: r.config.SimilarityThreshold,
		MaxDepth:            r.config.MaxDepth,
		Limit:               r.config.Limit,
	}, opts...)

	// callback
	ctx = callbacks.EnsureRunInfo(ctx, r.GetType(), components.ComponentOfRetriever)
	ctx = callbacks.OnStart(ctx, &retriever.CallbackInput{
		Query:          query,
		TopK:           options.TopK,
		Filter:         options.GetFilter(),
		ScoreThreshold: &options.SimilarityThreshold,
		Extra: map[string]any{
			"limit": options.Limit,
		},
	})
	defer func() {
		if err != nil {
			callbacks.OnError(ctx, err)
		}
	}()

	// 生成查询的embedding
	queryEmbedding, err := r.config.EmbeddingModel.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// 创建Neo4j会话
	session := r.config.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	// 执行图检索
	documents, err := r.graphSearch(ctx, session, query, queryEmbedding[0], options)
	if err != nil {
		return nil, fmt.Errorf("failed to perform graph search: %w", err)
	}

	callbacks.OnEnd(ctx, &retriever.CallbackOutput{Docs: documents})

	return documents, nil
}

// graphSearch 执行图检索
func (r *Retriever) graphSearch(ctx context.Context, session neo4j.SessionWithContext, query string, embedding []float64, opts *RetrieverImplOptions) ([]*schema.Document, error) {
	// 构建Cypher查询
	cypher := r.buildSearchCypher(opts)

	// 准备参数
	embeddingList := make([]interface{}, len(embedding))
	for i, v := range embedding {
		embeddingList[i] = v
	}

	params := map[string]interface{}{
		"query_embedding":      embeddingList,
		"similarity_threshold": opts.SimilarityThreshold,
		"top_k":                opts.TopK,
		"limit":                opts.Limit,
	}

	log.Printf("Neo4j Search - Query: %s, TopK: %d, Threshold: %f", query, opts.TopK, opts.SimilarityThreshold)

	// 执行查询
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var documents []*schema.Document
		for result.Next(ctx) {
			record := result.Record()
			doc, err := r.recordToDocument(record, query)
			if err != nil {
				log.Printf("Failed to convert record to document: %v", err)
				continue
			}
			documents = append(documents, doc)
		}

		return documents, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*schema.Document), nil
}

// buildSearchCypher 构建搜索的Cypher查询
func (r *Retriever) buildSearchCypher(opts *RetrieverImplOptions) string {
	return fmt.Sprintf(`
			CALL db.index.vector.queryNodes('entity_embedding_index', $top_k, $query_embedding)
			YIELD node, score
			WHERE score >= $similarity_threshold %s
			
			// 获取相关实体和关系
			OPTIONAL MATCH (node)-[rels*1..%d]-(related:Entity)
			WHERE all(r IN rels WHERE r.hit >= 1)  // 只获取有足够支撑的关系
			
			WITH node, rels, related, score
			
			RETURN 
			    node.name as entity_name,
			    node.type as entity_type,
				related.name as related_name,
				related.type as related_type,
				[rel in rels | type(rel)] as relations,
                size(rels) as layer,
				score
			ORDER BY score DESC, layer ASC

			LIMIT $limit
		`, opts.GetFilter(), opts.MaxDepth)
}

// recordToDocument 将Neo4j记录转换为schema.Document
func (r *Retriever) recordToDocument(record *neo4j.Record, originalQuery string) (*schema.Document, error) {
	entityName, _ := record.Get("entity_name")
	entityType, _ := record.Get("entity_type")
	relatedName, _ := record.Get("related_name")
	relatedType, _ := record.Get("related_type")
	score, _ := record.Get("score")
	layer, _ := record.Get("layer")
	relations, _ := record.Get("relations")

	if entityName == nil || relatedName == nil {
		return nil, fmt.Errorf("entity name or related name is nil")
	}

	// 构建文档ID
	docID := fmt.Sprintf("neo4j_entity_%v_%v", entityType, entityName)

	// 构建元数据
	metadata := map[string]interface{}{
		"source":                       "neo4j_graph",
		"entity_name":                  entityName,
		"entity_type":                  entityType,
		"related_name":                 relatedName,
		"related_type":                 relatedType,
		"query_entity_relevance_score": score, // 查询和实体相似度得分
		"retrieved_at":                 time.Now().Unix(),
	}

	// 添加实体属性
	if layer != nil {
		metadata["relation_layer"] = layer
	}

	// 添加关系信息
	if relations != nil {
		metadata["relations"] = relations
	}

	// 构建更丰富的文档内容
	docContent := r.buildDocumentContent(entityName, entityType, relatedName, relatedType, relations)

	return &schema.Document{
		ID:       docID,
		Content:  docContent,
		MetaData: metadata,
	}, nil
}

// buildDocumentContent 构建文档内容
func (r *Retriever) buildDocumentContent(entityName, entityType interface{}, relatedName, relatedType interface{}, relations interface{}) string {
	content := fmt.Sprintf("实体:(%v (%v))", entityName, entityType)

	// 添加关系信息
	if relations != nil {
		content += fmt.Sprintf("-> 关系: %v ->", relations)
	} else {
		content += "->"
	}

	content += fmt.Sprintf("相关实体:(%v (%v))\n", relatedName, relatedType)

	return content
}

// check 验证检索器配置
func (c *RetrieverConfig) check() error {
	if c.Driver == nil {
		return fmt.Errorf("[NewRetriever] Neo4j driver not provided")
	}
	if c.EmbeddingModel == nil {
		return fmt.Errorf("[NewRetriever] embedding model not provided")
	}
	if c.Dimension <= 0 {
		c.Dimension = 2048
	}
	if c.Database == "" {
		c.Database = "neo4j"
	}
	if c.SimilarityThreshold <= 0 {
		c.SimilarityThreshold = 0.7
	}
	if c.MaxDepth <= 0 {
		c.MaxDepth = 2
	}
	if c.TopK <= 0 {
		c.TopK = 10
	}
	if c.Limit <= 0 {
		c.Limit = 50
	}
	return nil
}

// GetType 返回检索器类型
func (r *Retriever) GetType() string {
	return retrieveTyp
}

// IsCallbacksEnabled 返回是否启用回调
func (r *Retriever) IsCallbacksEnabled() bool {
	return false
}
