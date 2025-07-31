package types

type CollectKnowledgeRequest struct {
	// 必填，知识内容
	Content string `json:"content,omitempty"`
	// 必填，知识内容链接 （知识内容和知识内容链接二选一）
	ContentLink string `json:"content_link,omitempty"`
}

type CollectKnowledgeRequestHeader struct {
	RequestID string `header:"request-id" json:"request_id,omitempty"`
	// CollectionKey 用于数据隔离.
	CollectionKey string `header:"collection-key" json:"collection_key,omitempty"`
	// AccessToken   string `header:"access-token" json:"access_token,omitempty"`
}

type CollectKnowledgeResponse struct {
	Message string `json:"message"`
}

const (
	QueryStrategyQuick = "quick"
	QueryStrategyAgent = "agent"
)

type RetrieveRequest struct {
	Query string `json:"query"`
	// 查询策略，默认是 "quick"，可选 "quick" 和 "agent"
	// quick: 快速查询，agent: 智能查询
	QueryStrategy string `json:"query_strategy"`
}

type RetrieveResponse struct {
	Content       string         `json:"content"`
	RetrieveItems []RetrieveItem `json:"retrieve_items"`
}

type RetrieveItem struct {
	ID       string         `json:"id"`
	Source   string         `json:"source"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}
