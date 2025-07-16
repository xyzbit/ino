package services

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/pkg/constants"
	milvusClient "github.com/xyzbit/ino/pkg/infra/milvus"
)

// Retriever 检索器
type Retriever struct {
	milvus retriever.Retriever
}

func NewRetriever() (*Retriever, error) {
	milvusRetriever, err := newMilvusRetriever(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create milvus retriever")
	}
	return &Retriever{
		milvus: milvusRetriever,
	}, nil
}

func (r *Retriever) Exec(ctx context.Context, req *types.RetrieveRequest) (*types.RetrieveResponse, error) {
	documents, err := r.milvus.Retrieve(ctx, "milvus")
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve")
	}

	for i, doc := range documents {
		fmt.Printf("Document %d:\n", i)
		fmt.Printf("title: %s\n", doc.ID)
		fmt.Printf("content: %s\n", doc.Content)
		fmt.Printf("metadata: %v\n", doc.MetaData)
	}

	return &types.RetrieveResponse{}, nil
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
	retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:     milvusClient.Client,
		Collection: constants.VectorCollectionName,
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		MetricType:     constants.VectorMetricType,
		TopK:           5,
		ScoreThreshold: 0.5,
		Embedding:      emb,
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
