package types

type CollectKnowledgeRequest struct {
	// 必填，知识内容
	Content string `json:"content,omitempty"`
	// 必填，知识内容链接 （知识内容和知识内容链接二选一）
	ContentLink string `json:"content_link,omitempty"`
}

type CollectKnowledgeRequestHeader struct {
	RequestID string `header:"request_id" json:"request_id,omitempty"`
	// CollectionKey 暂时不支持，用于数据隔离.
	CollectionKey string `header:"collection_key" json:"collection_key,omitempty"`
	UserKey       string `header:"user_key" json:"user_key,omitempty"`
}

type CollectKnowledgeResponse struct {
	Message string `json:"message"`
}
