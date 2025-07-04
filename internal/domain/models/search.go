package models

import (
	"time"
)

// SearchLog 搜索日志模型
type SearchLog struct {
	ID           uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	QueryID      string                 `json:"query_id" gorm:"uniqueIndex,type:varchar(64),not null"`
	UserID       uint64                 `json:"user_id"`
	User         *User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DomainID     uint64                 `json:"domain_id"`
	Domain       *Domain                `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
	QueryText    string                 `json:"query_text" gorm:"type:text,not null"`
	SearchConfig map[string]interface{} `json:"search_config" gorm:"type:json"`
	Results      SearchResults          `json:"results" gorm:"type:json"`
	ResponseTime int                    `json:"response_time_ms"`
	CreatedAt    time.Time              `json:"created_at"`
}

// TableName 指定表名
func (SearchLog) TableName() string {
	return "search_logs"
}

// SearchResults 搜索结果
type SearchResults struct {
	TotalHits    int                    `json:"total_hits"`
	ProcessingMS int                    `json:"processing_ms"`
	Results      []SearchResult         `json:"results"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

// SearchResult 单个搜索结果
type SearchResult struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // document, chunk, conversation
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Source     string                 `json:"source"`
	Score      float64                `json:"score"`
	Highlights []string               `json:"highlights"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query    string                 `json:"query" binding:"required"`
	DomainID uint64                 `json:"domain_id"`
	UserID   uint64                 `json:"user_id"`
	Filters  map[string]interface{} `json:"filters"`
	Options  SearchOptions          `json:"options"`
	Context  SearchContext          `json:"context"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Limit           int      `json:"limit"`            // 返回结果数量限制
	Offset          int      `json:"offset"`           // 偏移量
	ScoreThreshold  float64  `json:"score_threshold"`  // 分数阈值
	IncludeContent  bool     `json:"include_content"`  // 是否包含内容
	IncludeMetadata bool     `json:"include_metadata"` // 是否包含元数据
	ResultTypes     []string `json:"result_types"`     // 结果类型过滤
	SortBy          string   `json:"sort_by"`          // 排序字段
	SortOrder       string   `json:"sort_order"`       // 排序顺序
	Highlight       bool     `json:"highlight"`        // 是否高亮
	Rerank          bool     `json:"rerank"`           // 是否重排序
}

// SearchContext 搜索上下文
type SearchContext struct {
	SessionID      string                 `json:"session_id"`
	ConversationID string                 `json:"conversation_id"`
	PreviousQuery  string                 `json:"previous_query"`
	UserAgent      string                 `json:"user_agent"`
	IPAddress      string                 `json:"ip_address"`
	Location       string                 `json:"location"`
	Platform       string                 `json:"platform"`
	Source         string                 `json:"source"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	QueryID        string                 `json:"query_id"`
	Query          string                 `json:"query"`
	TotalHits      int                    `json:"total_hits"`
	ProcessingMS   int                    `json:"processing_ms"`
	Results        []SearchResult         `json:"results"`
	Aggregations   map[string]interface{} `json:"aggregations"`
	Suggestions    []string               `json:"suggestions"`
	RelatedQueries []string               `json:"related_queries"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ToResponse 转换为响应格式
func (sl *SearchLog) ToResponse() *SearchResponse {
	return &SearchResponse{
		QueryID:      sl.QueryID,
		Query:        sl.QueryText,
		TotalHits:    sl.Results.TotalHits,
		ProcessingMS: sl.ResponseTime,
		Results:      sl.Results.Results,
		Aggregations: sl.Results.Aggregations,
		Metadata:     sl.SearchConfig,
	}
}

// SearchStats 搜索统计
type SearchStats struct {
	TotalSearches   int     `json:"total_searches"`
	UniqueUsers     int     `json:"unique_users"`
	AvgResponseTime float64 `json:"avg_response_time"`
	TopQueries      []struct {
		Query string `json:"query"`
		Count int    `json:"count"`
	} `json:"top_queries"`
	PopularDomains []struct {
		DomainName string `json:"domain_name"`
		Count      int    `json:"count"`
	} `json:"popular_domains"`
	HourlyDistribution []struct {
		Hour  int `json:"hour"`
		Count int `json:"count"`
	} `json:"hourly_distribution"`
	QualityMetrics struct {
		ClickThroughRate float64 `json:"click_through_rate"`
		ZeroResultRate   float64 `json:"zero_result_rate"`
		AvgResultsCount  float64 `json:"avg_results_count"`
	} `json:"quality_metrics"`
}

// QuerySuggestion 查询建议
type QuerySuggestion struct {
	ID         uint64    `json:"id" gorm:"primaryKey,autoIncrement"`
	Query      string    `json:"query" gorm:"size:500,not null"`
	Suggestion string    `json:"suggestion" gorm:"size:500,not null"`
	Score      float64   `json:"score"`
	Source     string    `json:"source"` // auto, manual
	Status     string    `json:"status"` // active, inactive
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
