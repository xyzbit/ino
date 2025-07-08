package services

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/internal/domain/models"
	"golang.org/x/sync/errgroup"
)

// Indexer 索引器
type Indexer struct {
	embeddingModel *openai.ChatModel
	extractorModel *openai.ChatModel
}

// NewRequestOptimizer 创建请求优化器实例
func NewIndexer() (*Indexer, error) {
	embeddingConfig := config.AppConfig.Indexer.Embedding
	extractorConfig := config.AppConfig.Indexer.Extractor
	embeddingModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: embeddingConfig.BaseURL,
		Model:   embeddingConfig.Model,
		APIKey:  embeddingConfig.APIKey,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	extractorModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: extractorConfig.BaseURL,
		Model:   extractorConfig.Model,
		APIKey:  extractorConfig.APIKey,
	})
	extractorModel.WithTools([]*schema.ToolInfo{
		{
			Name:        "extract_entity_and_relation",
			Description: "提取实体和关系",
			Parameters:  schema.FString,
		},
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Indexer{
		embeddingModel: embeddingModel,
		extractorModel: extractorModel,
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
	// 1. 大模型提取知识中的实体和关系内容
	i.extractEntityAndRelation(ctx, req)
	// 2. 查询知识图谱，获取相关知识
	// i.searchRelatedKnowledge(ctx, req)
	// 3. 大模型判断相关知识和当前请求是否存在冲突，获取需要删除内容
	// i.getConflictKnowledge(ctx, req)
	// 4. 更新知识图谱

	return nil
}

func (i *Indexer) extractEntityAndRelation(ctx context.Context, req *types.CollectKnowledgeRequest) error {
	msgs, err := models.PromptGraphExtractEntityAndRelation.Format(ctx, map[string]any{
		"origin_request": req.Content,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	output, err := i.extractorModel.Generate(ctx, msgs)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, toolCall := range output.ToolCalls {
		// if toolCall.ToolName == "extract_entity_and_relation" {
		// 	entityAndRelation := toolCall.ToolCallID
		// 	fmt.Println(entityAndRelation)
		// }
	}

	return nil
}

func BuildKnowledgeIndexing(ctx context.Context) (r compose.Runnable[*types.CollectKnowledgeRequest, []string], err error) {
	const (
		FileLoader       = "FileLoader"
		MarkdownSplitter = "MarkdownSplitter"
		RedisIndexer     = "RedisIndexer"
	)
	g := compose.NewGraph[document.Source, []string]()
	fileLoaderKeyOfLoader, err := newLoader(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)
	markdownSplitterKeyOfDocumentTransformer, err := newDocumentTransformer(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddDocumentTransformerNode(MarkdownSplitter, markdownSplitterKeyOfDocumentTransformer)
	redisIndexerKeyOfIndexer, err := newIndexer(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddIndexerNode(RedisIndexer, redisIndexerKeyOfIndexer)
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, MarkdownSplitter)
	_ = g.AddEdge(MarkdownSplitter, RedisIndexer)
	_ = g.AddEdge(RedisIndexer, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}

// newLoader component initialization function of node 'FileLoader' in graph 'KnowledgeIndexing'
func newLoader(ctx context.Context) (ldr document.Loader, err error) {
	// TODO Modify component configuration here.
	config := &file.FileLoaderConfig{}
	ldr, err = file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}
	return ldr, nil
}
