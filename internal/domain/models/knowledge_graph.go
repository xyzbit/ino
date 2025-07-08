package models

import (
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// PromptGraphExtractEntityAndRelation 提取实体和关系提示
// params:
// - origin_request: 原始请求内容
var PromptGraphExtractEntityAndRelation = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`
提取关系提示：

你是一个先进的算法，旨在从文本中提取结构化信息以构建知识图谱。你的目标是捕捉全面且准确的信息。请遵循以下关键原则：

1. 仅从文本中提取明确陈述的信息。
2. 在提供的实体之间建立关系。
3. 在用户消息中, 对于任何自我指代 (如“我”“我的”等), 如果无法通过上下文判断究竟是谁, 则过滤掉。

关系：
- 使用一致、通用且不受时间限制的关系类型。
- 例如：优先使用“教授”，而非“成为教授”。
- 关系仅应在用户消息中明确提及的实体之间建立。

实体一致性：
- 确保关系连贯，且在消息语境下逻辑一致。
- 在提取的数据中，保持实体命名的一致性。

努力通过在实体之间建立所有关系并遵循用户的语境，构建一个连贯且易于理解的知识图谱。

严格遵守这些准则，以确保高质量的知识图谱提取。
`),
	schema.UserMessage(`
	请从以下文本中提取实体和关系：
	{origin_request}

请返回JSON格式的提取结果。
	`),
)

// KnowledgeEntity 知识实体
type KnowledgeEntity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // person, organization, concept, etc.
	Name       string                 `json:"name"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
	Source     string                 `json:"source"` // 来源文档或对话
	Score      float64                `json:"score"`  // 置信度分数
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// KnowledgeRelation 知识关系
type KnowledgeRelation struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`        // 关系类型
	FromEntity string                 `json:"from_entity"` // 源实体ID
	ToEntity   string                 `json:"to_entity"`   // 目标实体ID
	Properties map[string]interface{} `json:"properties"`
	Source     string                 `json:"source"` // 来源文档或对话
	Score      float64                `json:"score"`  // 置信度分数
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// KnowledgeGraph 知识图谱
type KnowledgeGraph struct {
	Entities  []KnowledgeEntity   `json:"entities"`
	Relations []KnowledgeRelation `json:"relations"`
	Metadata  GraphMetadata       `json:"metadata"`
}

// GraphMetadata 图谱元数据
type GraphMetadata struct {
	DomainID       uint64    `json:"domain_id"`
	TotalEntities  int       `json:"total_entities"`
	TotalRelations int       `json:"total_relations"`
	EntityTypes    []string  `json:"entity_types"`
	RelationTypes  []string  `json:"relation_types"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GraphTraversal 图遍历配置
type GraphTraversal struct {
	StartEntity   string   `json:"start_entity"`
	MaxDepth      int      `json:"max_depth"`
	Direction     string   `json:"direction"`      // IN, OUT, BOTH
	RelationTypes []string `json:"relation_types"` // 关系类型过滤
	EntityTypes   []string `json:"entity_types"`   // 实体类型过滤
	MinScore      float64  `json:"min_score"`      // 最小置信度
	Limit         int      `json:"limit"`          // 结果数量限制
}

// GraphTraversalResult 图遍历结果
type GraphTraversalResult struct {
	Paths []GraphPath `json:"paths"`
	Stats struct {
		TotalPaths    int     `json:"total_paths"`
		AvgPathLength float64 `json:"avg_path_length"`
		MaxDepth      int     `json:"max_depth"`
		ExecutionTime int     `json:"execution_time_ms"`
	} `json:"stats"`
}

// GraphPath 图路径
type GraphPath struct {
	Entities  []KnowledgeEntity   `json:"entities"`
	Relations []KnowledgeRelation `json:"relations"`
	Score     float64             `json:"score"`
	Length    int                 `json:"length"`
}

// EntityExtractionRequest 实体抽取请求
type EntityExtractionRequest struct {
	Text     string `json:"text" binding:"required"`
	DomainID uint64 `json:"domain_id" binding:"required"`
	Source   string `json:"source"`
	Language string `json:"language"`
	Options  struct {
		ExtractTypes     []string `json:"extract_types"`     // 要抽取的实体类型
		MinConfidence    float64  `json:"min_confidence"`    // 最小置信度
		MergeEntities    bool     `json:"merge_entities"`    // 是否合并相似实体
		ExtractRelations bool     `json:"extract_relations"` // 是否抽取关系
	} `json:"options"`
}

// EntityExtractionResponse 实体抽取响应
type EntityExtractionResponse struct {
	Entities  []KnowledgeEntity   `json:"entities"`
	Relations []KnowledgeRelation `json:"relations"`
	Metadata  struct {
		ProcessingTime int     `json:"processing_time_ms"`
		Confidence     float64 `json:"avg_confidence"`
		Language       string  `json:"language"`
		Model          string  `json:"model"`
	} `json:"metadata"`
}

// GraphSearchRequest 图搜索请求
type GraphSearchRequest struct {
	Query     string         `json:"query" binding:"required"`
	DomainID  uint64         `json:"domain_id"`
	Traversal GraphTraversal `json:"traversal"`
	Options   struct {
		IncludeSubgraph bool    `json:"include_subgraph"`
		ScoreThreshold  float64 `json:"score_threshold"`
		Limit           int     `json:"limit"`
	} `json:"options"`
}

// GraphSearchResponse 图搜索响应
type GraphSearchResponse struct {
	Results []GraphSearchResult `json:"results"`
	Stats   struct {
		TotalResults  int     `json:"total_results"`
		AvgScore      float64 `json:"avg_score"`
		ExecutionTime int     `json:"execution_time_ms"`
	} `json:"stats"`
}

// GraphSearchResult 图搜索结果
type GraphSearchResult struct {
	Entity    KnowledgeEntity `json:"entity"`
	Subgraph  *KnowledgeGraph `json:"subgraph,omitempty"`
	Score     float64         `json:"score"`
	Relevance string          `json:"relevance"` // 相关性描述
	Path      []string        `json:"path"`      // 到达路径
}

// GraphStats 图统计信息
type GraphStats struct {
	TotalEntities  int `json:"total_entities"`
	TotalRelations int `json:"total_relations"`
	EntityTypes    []struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	} `json:"entity_types"`
	RelationTypes []struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	} `json:"relation_types"`
	Density             float64 `json:"density"`
	AvgDegree           float64 `json:"avg_degree"`
	MaxDegree           int     `json:"max_degree"`
	ConnectedComponents int     `json:"connected_components"`
}
