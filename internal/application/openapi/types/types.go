package types

type CollectKnowledgeRequest struct {
	// 必填，知识内容
	Content string `json:"content,omitempty"`
	// 必填，知识内容链接 （知识内容和知识内容链接二选一）
	ContentLink string `json:"content_link,omitempty"`
}

type CollectKnowledgeRequestHeader struct {
	RequestID    string `header:"request_id" json:"request_id,omitempty"`
	CollectionID string `header:"collection_id" json:"collection_id,omitempty"`
	User         string `header:"user" json:"user,omitempty"`
}

type CollectKnowledgeResponse struct {
	Message string `json:"message"`
}
