package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"

	"github.com/xyzbit/ino/pkg/components/extractor/konwledge"
	neo4jIndexer "github.com/xyzbit/ino/pkg/components/indexer/neo4j"
	"github.com/xyzbit/ino/pkg/constants"
	milvusClient "github.com/xyzbit/ino/pkg/infra/milvus"
	neo4jClient "github.com/xyzbit/ino/pkg/infra/neo4j"
	redisClient "github.com/xyzbit/ino/pkg/infra/redis"
)

const (
	typ                           = "Milvus"
	defaultCollection             = "eino_collection"
	defaultDescription            = "the collection for eino"
	defaultCollectionID           = "id"
	defaultCollectionIDDesc       = "the unique id of the document"
	defaultCollectionVector       = "vector"
	defaultCollectionVectorDesc   = "the vector of the document"
	defaultCollectionContent      = "content"
	defaultCollectionContentDesc  = "the content of the document"
	defaultCollectionMetadata     = "metadata"
	defaultCollectionMetadataDesc = "the metadata of the document"

	defaultDim = 2560 // 注意要和向量模型维度一致
)

const (
	nodeRequestToDocs      = "requestToDocs"
	nodeAutoSpliter        = "autoSpliter"
	nodeMilvusIndexer      = "milvusIndexer"
	nodeRedisIndexer       = "redisIndexer"
	nodeNeo4jIndexer       = "neo4jIndexer"
	nodeKnowledgeExtractor = "knowledgeExtractor"
)

// Indexer 索引器
type Indexer struct {
	vectorIndexer indexer.Indexer
	redisIndexer  indexer.Indexer
	graphIndexer  indexer.Indexer
	autoSpliter   document.Transformer
	extractor     compose.Invoke[[]*schema.Document, []*schema.Document, any]
}

// NewIndexer 创建索引器实例
func NewIndexer() (*Indexer, error) {
	vectorIndexer, err := newVectorIndexer(context.Background())
	if err != nil {
		return nil, err
	}

	redisIndexer, err := newRedisIndexer(context.Background())
	if err != nil {
		return nil, err
	}

	graphIndexer, err := newGraphIndexer(context.Background())
	if err != nil {
		return nil, err
	}

	autoSpliter, err := newAutoSpliter(context.Background())
	if err != nil {
		return nil, err
	}

	extractor, err := newKnowledgeExtractor(context.Background())
	if err != nil {
		return nil, err
	}

	return &Indexer{
		vectorIndexer: vectorIndexer,
		redisIndexer:  redisIndexer,
		graphIndexer:  graphIndexer,
		autoSpliter:   autoSpliter,
		extractor:     extractor,
	}, nil
}

func (i *Indexer) Exec(ctx context.Context, req *types.CollectKnowledgeRequest) error {
	runner, err := i.buildKnowledgeIndexing(ctx)
	if err != nil {
		return err
	}
	reply, err := runner.Invoke(ctx, req)
	if err != nil {
		return err
	}
	fmt.Println(reply)
	return nil
}

func (i *Indexer) buildKnowledgeIndexing(ctx context.Context) (r compose.Runnable[*types.CollectKnowledgeRequest, map[string]any], err error) {

	g := compose.NewGraph[*types.CollectKnowledgeRequest, map[string]any]()

	// add node
	_ = g.AddLambdaNode(nodeRequestToDocs, compose.InvokableLambdaWithOption(requestToDocs))
	_ = g.AddDocumentTransformerNode(nodeAutoSpliter, i.autoSpliter)
	_ = g.AddIndexerNode(nodeMilvusIndexer, i.vectorIndexer, compose.WithOutputKey("milvus_ids"))
	_ = g.AddIndexerNode(nodeNeo4jIndexer, i.graphIndexer, compose.WithOutputKey("neo4j_ids"))
	_ = g.AddIndexerNode(nodeRedisIndexer, i.redisIndexer, compose.WithOutputKey("redis_ids"))
	_ = g.AddLambdaNode(nodeKnowledgeExtractor, compose.InvokableLambdaWithOption(i.extractor))
	// add edge
	_ = g.AddEdge(compose.START, nodeRequestToDocs)
	_ = g.AddEdge(nodeRequestToDocs, nodeAutoSpliter)
	_ = g.AddEdge(nodeAutoSpliter, nodeKnowledgeExtractor)
	_ = g.AddEdge(nodeKnowledgeExtractor, nodeMilvusIndexer)
	_ = g.AddEdge(nodeKnowledgeExtractor, nodeNeo4jIndexer)
	_ = g.AddEdge(nodeMilvusIndexer, compose.END)
	_ = g.AddEdge(nodeNeo4jIndexer, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}

	return r, err
}

func requestToDocs(ctx context.Context, input *types.CollectKnowledgeRequest, opts ...any) (output []*schema.Document, err error) {
	if input.Content != "" {
		return []*schema.Document{
			{
				ID:      uuid.New().String(),
				Content: input.Content,
			},
		}, nil
	}
	config := &file.FileLoaderConfig{UseNameAsID: true}
	ldr, err := file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}

	docs, err := ldr.Load(ctx, document.Source{
		URI: input.ContentLink,
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// TODO: 自动分段器,后续让大模型根据内容自动选择分段器（markdown、html、recursive、semantic）
type AutoSpliter struct {
	markdownSplitter  document.Transformer
	recursiveSplitter document.Transformer
}

func (a *AutoSpliter) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	if len(src) == 0 {
		return src, nil
	}

	doc := src[0]
	if doc.MetaData[file.MetaKeyExtension] == ".md" || doc.MetaData[file.MetaKeyExtension] == "md" {
		return a.markdownSplitter.Transform(ctx, src, opts...)
	}

	return a.recursiveSplitter.Transform(ctx, src, opts...)
}

// newAutoSpliter
func newAutoSpliter(ctx context.Context) (tfr document.Transformer, err error) {
	config := &markdown.HeaderConfig{
		Headers:     map[string]string{"##": "headerNameOfLevel2"},
		TrimHeaders: false,
	}
	mkdSplitter, err := markdown.NewHeaderSplitter(ctx, config)
	if err != nil {
		return nil, err
	}

	recursiveSplitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   1500,
		OverlapSize: 300,
	})
	if err != nil {
		return nil, err
	}

	return &AutoSpliter{
		markdownSplitter:  mkdSplitter,
		recursiveSplitter: recursiveSplitter,
	}, nil
}

func newKnowledgeExtractor(ctx context.Context) (compose.Invoke[[]*schema.Document, []*schema.Document, any], error) {
	extractorConfig := config.AppConfig.Indexer.Extractor

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: extractorConfig.BaseURL,
		Model:   extractorConfig.Model,
		APIKey:  extractorConfig.APIKey,
	})
	if err != nil {
		return nil, err
	}

	extractor, err := konwledge.NewExtractor(ctx, &konwledge.Config{
		Extractor: model,
	})
	if err != nil {
		return nil, err
	}

	return extractor.Extract, nil
}

func newRedisIndexer(ctx context.Context) (idr indexer.Indexer, err error) {
	indexerConfig := &redis.IndexerConfig{
		Client:    redisClient.Redis,
		KeyPrefix: "ino_collection",
		BatchSize: 1,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = uuid.New().String()
			}
			key := doc.ID

			metadataBytes, err := json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}

			return &redis.Hashes{
				Key: key,
				Field2Value: map[string]redis.FieldValue{
					redisClient.ContentField:  {Value: doc.Content, EmbedKey: redisClient.VectorField},
					redisClient.MetadataField: {Value: metadataBytes},
				},
			}, nil
		},
	}

	embeddingConfig := config.AppConfig.Indexer.Embedding
	embeddingIns11, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, err
	}
	indexerConfig.Embedding = embeddingIns11
	idr, err = redis.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, err
	}
	return idr, nil
}

func newVectorIndexer(ctx context.Context) (idx indexer.Indexer, err error) {
	embeddingConfig := config.AppConfig.Indexer.Embedding
	// Create an embedding model
	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, err
	}

	// Create an indexer
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:     milvusClient.Client,
		Embedding:  emb,
		Collection: constants.VectorCollectionName,
		MetricType: milvus.MetricType(constants.VectorMetricType),
		Fields: []*entity.Field{
			entity.NewField().
				WithName(defaultCollectionID).
				WithDescription(defaultCollectionIDDesc).
				WithIsPrimaryKey(true).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(255),
			entity.NewField().
				WithName(defaultCollectionVector).
				WithDescription(defaultCollectionVectorDesc).
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeFloatVector). //FieldTypeBinaryVector
				WithDim(defaultDim),
			entity.NewField().
				WithName(defaultCollectionContent).
				WithDescription(defaultCollectionContentDesc).
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(4096),
			entity.NewField().
				WithName(defaultCollectionMetadata).
				WithDescription(defaultCollectionMetadataDesc).
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeJSON),
		},
		// 自定义 DocumentConverter 用于 FloatVector 类型
		DocumentConverter: func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
			rows := make([]interface{}, 0, len(docs))

			for idx, doc := range docs {
				// 序列化 metadata
				metadataBytes, err := sonic.Marshal(doc.MetaData)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata: %w", err)
				}

				// 转换向量为 float32
				vector32 := make([]float32, len(vectors[idx]))
				for i, v := range vectors[idx] {
					vector32[i] = float32(v)
				}

				// 创建行数据
				row := map[string]interface{}{
					defaultCollectionID:       doc.ID,
					defaultCollectionContent:  doc.Content,
					defaultCollectionVector:   vector32,
					defaultCollectionMetadata: metadataBytes,
				}
				rows = append(rows, row)
			}
			return rows, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return indexer, nil
}

func newGraphIndexer(ctx context.Context) (idx indexer.Indexer, err error) {
	extractorConfig := config.AppConfig.Indexer.Extractor
	embeddingConfig := config.AppConfig.Indexer.Embedding

	// Create an extractor model
	extractorModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: extractorConfig.BaseURL,
		Model:   extractorConfig.Model,
		APIKey:  extractorConfig.APIKey,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create an embedding model
	embeddingModel, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create a Neo4j indexer
	graphIndexer, err := neo4jIndexer.NewIndexer(ctx, &neo4jIndexer.IndexerConfig{
		Driver:         neo4jClient.Driver,
		Extractor:      extractorModel,
		EmbeddingModel: embeddingModel,
		BatchSize:      50,
		Dimension:      defaultDim,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return graphIndexer, nil
}
