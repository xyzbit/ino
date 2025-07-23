package services

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
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

type QueryStrategyHandler func(ctx context.Context, query string) (*types.RetrieveResponse, error)

// queryStrategiesHandlers 查询策略处理器集合.
// note: not thread safe
var queryStrategiesHandlers = make(map[string]QueryStrategyHandler)

// Retriever 检索器
type Retriever struct {
	milvus      retriever.Retriever
	neo4j       retriever.Retriever
	agentConfig *react.AgentConfig
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

	agentConfig, err := newAgentConfig(context.Background(), milvusRetriever, neo4jRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create agent")
	}

	r := &Retriever{
		milvus:      milvusRetriever,
		neo4j:       neo4jRetriever,
		agentConfig: agentConfig,
	}
	queryStrategiesHandlers[types.QueryStrategyQuick] = r.RetrieveQuick
	queryStrategiesHandlers[types.QueryStrategyAgent] = r.RetrieveAgent

	return r, nil
}

func (r *Retriever) Exec(ctx context.Context, req *types.RetrieveRequest) (*types.RetrieveResponse, error) {
	handler, ok := queryStrategiesHandlers[req.QueryStrategy]
	if !ok {
		return nil, errors.New("query strategy not found")
	}
	return handler(ctx, req.Query)
}

func (r *Retriever) RetrieveQuick(ctx context.Context, query string) (*types.RetrieveResponse, error) {
	re, err := r.buildKnowledgeRetriever(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := re.Invoke(ctx, query)
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

type VectorDBSearchParams struct {
	Query          string  `json:"query" jsonschema:"required,description=the query to search"`
	TopK           int     `json:"top_k" jsonschema:"description=the top k results to return, default is no limit"`
	ScoreThreshold float64 `json:"score_threshold" jsonschema:"description=the score threshold to filter results, default is no filter"`
}

type GraphDBSearchParams struct {
	Query          string  `json:"query" jsonschema:"required,description=the query to search"`
	TopK           int     `json:"top_k" jsonschema:"description=the top k results to return, default is no limit"`
	ScoreThreshold float64 `json:"score_threshold" jsonschema:"description=the score threshold to filter results, default is no filter"`
	MaxDepth       int     `json:"max_depth" jsonschema:"description=the max depth to search, default is 2"`
	Limit          int     `json:"limit" jsonschema:"description=the limit to search, default is 50"`
}

func (r *Retriever) RetrieveAgent(ctx context.Context, query string) (*types.RetrieveResponse, error) {
	ragent, err := react.NewAgent(ctx, r.agentConfig)
	if err != nil {
		return nil, err
	}

	reply, err := ragent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(`
		你是一位具备深度检索与精准整合能力的知识库专家。当接收用户问题时，请遵循以下要求执行任务：

深度检索启动：以问题核心为锚点，全面遍历知识库相关领域内容，包括但不限于核心概念、关联背景、细分维度、典型案例及权威解释。若初始检索信息不足，需进一步拓展检索范围，挖掘次级关联内容，确保覆盖问题涉及的关键细节与潜在延伸点。
信息筛选与验证：对检索到的内容进行真实性、相关性校验，优先保留权威来源、逻辑严谨的信息，剔除冗余或冲突内容，确保回答的准确性与可靠性。
深度与完整性保障：回答需覆盖用户问题的显性需求与合理隐性需求，避免浅尝辄止。若问题存在多视角解读，需客观呈现不同观点并说明适用场景，确保回答的全面性与深度。
多轮检索：如果检索到的内容不足以回答用户问题，请继续检索、分析，直到找到足够的信息为止。

返回内容格式：
{
	"content": "返回内容",
	"reasoning_content": "推理内容"
}
`),
		schema.UserMessage(query),
	})
	if err != nil {
		return nil, err
	}

	log.Printf("reasoningContent: %s", reply.ReasoningContent)

	return &types.RetrieveResponse{
		Content: reply.Content,
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

func newAgentConfig(
	ctx context.Context,
	milvusRetriever retriever.Retriever,
	neo4jRetriever retriever.Retriever,
) (agentConfig *react.AgentConfig, err error) {
	llmConfig := config.AppConfig.Retriever.LLM
	maxRounds := config.AppConfig.Retriever.MaxRounds

	// init chat model
	toolableChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   llmConfig.Model,
		BaseURL: llmConfig.BaseURL,
		APIKey:  llmConfig.APIKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chat model")
	}

	// init tools
	vectorDBSearchToolInfo, err := utils.GoStruct2ToolInfo[VectorDBSearchParams](
		"search_by_vector_db",
		"在向量数据库 milvus 中搜索相关内容",
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vector db search tool")
	}

	vectorDBSearchTool := utils.NewTool(vectorDBSearchToolInfo,
		func(ctx context.Context, input *VectorDBSearchParams) (output []*schema.Document, err error) {
			if input == nil {
				return nil, errors.New("input is nil")
			}

			opts := make([]retriever.Option, 0)
			if input.ScoreThreshold != 0 {
				opts = append(opts, retriever.WithScoreThreshold(input.ScoreThreshold))
			}
			if input.TopK != 0 {
				opts = append(opts, retriever.WithTopK(input.TopK))
			}

			docs, err := milvusRetriever.Retrieve(ctx, input.Query, opts...)
			if err != nil {
				return nil, err
			}
			return docs, nil
		},
	)

	graphDBSearchToolInfo, err := utils.GoStruct2ToolInfo[GraphDBSearchParams](
		"search_entity_and_relation_by_graph_db",
		"在图数据库 neo4j 中搜索实体和关系",
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph db search tool")
	}
	graphDBSearchTool := utils.NewTool(graphDBSearchToolInfo,
		func(ctx context.Context, input *GraphDBSearchParams) (output []*schema.Document, err error) {
			if input == nil {
				return nil, errors.New("input is nil")
			}

			opts := make([]retriever.Option, 0)
			if input.ScoreThreshold != 0 {
				opts = append(opts, neo4j.WithSimilarityThreshold(input.ScoreThreshold))
			}
			if input.TopK != 0 {
				opts = append(opts, neo4j.WithTopK(input.TopK))
			}
			if input.MaxDepth != 0 {
				opts = append(opts, neo4j.WithMaxDepth(input.MaxDepth))
			}
			if input.Limit != 0 {
				opts = append(opts, neo4j.WithLimit(input.Limit))
			}

			docs, err := neo4jRetriever.Retrieve(ctx, input.Query, opts...)
			if err != nil {
				return nil, err
			}
			return docs, nil
		},
	)

	// create agent
	return &react.AgentConfig{
		ToolCallingModel: toolableChatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				vectorDBSearchTool,
				graphDBSearchTool,
			},
		},
		MaxStep: maxRounds,
	}, nil
}
