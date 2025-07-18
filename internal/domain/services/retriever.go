package services

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/pkg/components/retrieve/neo4j"
	"github.com/xyzbit/ino/pkg/constants"
	milvusClient "github.com/xyzbit/ino/pkg/infra/milvus"
	neo4jClient "github.com/xyzbit/ino/pkg/infra/neo4j"
)

const (
	nodeMilvusRetriever = "milvus_retriever"
	nodeNeo4jRetriever  = "neo4j_retriever"
)

// Retriever 检索器
type Retriever struct {
	milvus retriever.Retriever
	neo4j  retriever.Retriever
}

func NewRetriever() (*Retriever, error) {
	milvusRetriever, err := newMilvusRetriever(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create milvus retriever")
	}
	neo4jRetriever, err := newNeo4jRetriever(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create neo4j retriever")
	}
	return &Retriever{
		milvus: milvusRetriever,
		neo4j:  neo4jRetriever,
	}, nil
}

func (r *Retriever) Exec(ctx context.Context, req *types.RetrieveRequest) (*types.RetrieveResponse, error) {
	re, err := r.buildKnowledgeRetriever(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := re.Invoke(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	var (
		milvusItems []types.RetrieveItem
		neo4jItems  []types.RetrieveItem
	)

	if r := reply["milvus_documents"]; r != nil {
		if docs, ok := r.([]*schema.Document); ok {
			milvusItems = lo.Map(docs, func(doc *schema.Document, _ int) types.RetrieveItem {
				return types.RetrieveItem{
					Source:   "milvus",
					ID:       doc.ID,
					Content:  doc.Content,
					Metadata: doc.MetaData,
				}
			})
		}
	}
	if r := reply["neo4j_documents"]; r != nil {
		if docs, ok := r.([]*schema.Document); ok {
			neo4jItems = lo.Map(docs, func(doc *schema.Document, _ int) types.RetrieveItem {
				return types.RetrieveItem{
					Source:   "neo4j",
					ID:       doc.ID,
					Content:  doc.Content,
					Metadata: doc.MetaData,
				}
			})
		}
	}

	return &types.RetrieveResponse{
		RetrieveItems: append(milvusItems, neo4jItems...),
	}, nil
}

func (r *Retriever) buildKnowledgeRetriever(ctx context.Context) (retriever compose.Runnable[string, map[string]any], err error) {

	g := compose.NewGraph[string, map[string]any]()

	// add node
	g.AddRetrieverNode(nodeMilvusRetriever, r.milvus, compose.WithOutputKey("milvus_documents"))
	g.AddRetrieverNode(nodeNeo4jRetriever, r.neo4j, compose.WithOutputKey("neo4j_documents"))

	// add edge
	_ = g.AddEdge(compose.START, nodeMilvusRetriever)
	_ = g.AddEdge(compose.START, nodeNeo4jRetriever)
	_ = g.AddEdge(nodeMilvusRetriever, compose.END)
	_ = g.AddEdge(nodeNeo4jRetriever, compose.END)

	retriever, err = g.Compile(ctx, compose.WithGraphName("KnowledgeRetriever"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return retriever, nil
}

func newMilvusRetriever(ctx context.Context) (retriever.Retriever, error) {
	embeddingConfig := config.AppConfig.Indexer.Embedding
	// Create an embedding model
	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create embedding model")
	}

	// Create a retriever
	sp, _ := entity.NewIndexAUTOINDEXSearchParam(1)
	retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:     milvusClient.Client,
		Collection: constants.VectorCollectionName,
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		MetricType: constants.VectorMetricType,
		TopK:       5,
		Sp:         sp,
		Embedding:  emb,
		VectorConverter: func(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
			vec := make([]entity.Vector, 0, len(vectors))
			for _, vector := range vectors {
				vector32 := make([]float32, len(vector))
				for i, v := range vector {
					vector32[i] = float32(v)
				}
				vec = append(vec, entity.FloatVector(vector32))
			}
			return vec, nil
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create retriever")
	}
	return retriever, nil
}

// NewNeo4jRetriever 创建Neo4j检索器的便捷函数
func newNeo4jRetriever(ctx context.Context) (retriever.Retriever, error) {
	embeddingConfig := config.AppConfig.Indexer.Embedding

	// 创建embedding模型
	embeddingModel, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding model: %w", err)
	}

	// 创建Neo4j检索器
	r, err := neo4j.NewRetriever(ctx, &neo4j.RetrieverConfig{
		Driver:              neo4jClient.Driver,
		Database:            "neo4j",
		EmbeddingModel:      embeddingModel,
		Dimension:           defaultDim,
		SimilarityThreshold: 0.7,
		MaxDepth:            2,
		TopK:                10,
		Limit:               50,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j retriever: %w", err)
	}

	return r, nil
}
