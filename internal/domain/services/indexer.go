package services

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"golang.org/x/sync/errgroup"

	"github.com/xyzbit/ino/pkg/components/extractor/konwledge"
	neo4jIndexer "github.com/xyzbit/ino/pkg/components/indexer/neo4j"
	milvusClient "github.com/xyzbit/ino/pkg/infra/milvus"
	neo4jClient "github.com/xyzbit/ino/pkg/infra/neo4j"
)

// Indexer 索引器
type Indexer struct {
	vectorIndexer indexer.Indexer
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
		graphIndexer:  graphIndexer,
		autoSpliter:   autoSpliter,
		extractor:     extractor,
	}, nil
}

func (i *Indexer) Exec(ctx context.Context, req *types.CollectKnowledgeRequest) error {
	var eg errgroup.Group

	gctx := context.WithoutCancel(ctx)
	eg.Go(func() error {
		return i.addtoVectorStore(gctx, req)
	})
	eg.Go(func() error {
		return i.addtoGraphStore(gctx, req)
	})

	return eg.Wait()
}

func (i *Indexer) addtoVectorStore(ctx context.Context, req *types.CollectKnowledgeRequest) error {
	return nil
}

func (i *Indexer) addtoGraphStore(ctx context.Context, req *types.CollectKnowledgeRequest) error {
	runner, err := i.buildKnowledgeIndexing(ctx)
	if err != nil {
		return err
	}
	ids, err := runner.Invoke(ctx, req)
	if err != nil {
		return err
	}
	fmt.Println(ids)

	return nil
}

func (i *Indexer) buildKnowledgeIndexing(ctx context.Context) (r compose.Runnable[*types.CollectKnowledgeRequest, []string], err error) {
	const (
		RequestToDocs      = "RequestToDocs"
		AutoSpliter        = "AutoSpliter"
		MilvusIndexer      = "MilvusIndexer"
		Neo4jIndexer       = "Neo4jIndexer"
		KnowledgeExtractor = "KnowledgeExtractor"
	)
	g := compose.NewGraph[*types.CollectKnowledgeRequest, []string]()

	// add node
	_ = g.AddLambdaNode(RequestToDocs, compose.InvokableLambdaWithOption(requestToDocs))
	_ = g.AddDocumentTransformerNode(AutoSpliter, i.autoSpliter)
	_ = g.AddIndexerNode(MilvusIndexer, i.vectorIndexer)
	_ = g.AddIndexerNode(Neo4jIndexer, i.graphIndexer)
	_ = g.AddLambdaNode(KnowledgeExtractor, compose.InvokableLambdaWithOption(i.extractor))
	// add edge
	_ = g.AddEdge(compose.START, RequestToDocs)
	_ = g.AddEdge(RequestToDocs, AutoSpliter)
	_ = g.AddEdge(AutoSpliter, KnowledgeExtractor)
	_ = g.AddEdge(KnowledgeExtractor, MilvusIndexer)
	_ = g.AddEdge(KnowledgeExtractor, Neo4jIndexer)
	_ = g.AddEdge(MilvusIndexer, compose.END)
	_ = g.AddEdge(Neo4jIndexer, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
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
		Client:    milvusClient.Client,
		Embedding: emb,
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
		Database:       "neo4j",
		Extractor:      extractorModel,
		EmbeddingModel: embeddingModel,
		BatchSize:      50,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return graphIndexer, nil
}
