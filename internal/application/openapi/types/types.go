package types

import "encoding/json"

type CollectKnowledgeRequest struct {
	// 必填，知识库ID
	CollectionID string `json:"collection_id,omitempty"`
	// 可选，知识内容涉及的标签, 不同维度的标识内容
	Tags map[string]string `json:"tags,omitempty"`
	// 可选，知识内容类型, 可选值: conversation, feedback, document
	ContentType string `json:"content_type,omitempty"`
	// 必填，知识内容
	Content string `json:"content,omitempty"`
}

func (c *CollectKnowledgeRequest) GetDescription() string {
	desc := []map[string]interface{}{
		{
			"field_name":  "tags",
			"field_type":  "map[string]string",
			"description": "可选，知识内容涉及的标签, 不同维度的标识内容",
			"value":       c.Tags,
		},
		{
			"field_name":  "content_type",
			"field_type":  "string",
			"description": "可选，知识内容类型, 可选值: auto, conversation, feedback, document",
			"value":       c.ContentType,
		},
		{
			"field_name":  "content",
			"field_type":  "string",
			"description": "必填，知识内容",
			"value":       c.Content,
		},
	}
	str, _ := json.Marshal(desc)
	return string(str)
}

type CollectKnowledgeResponse struct {
	Message string `json:"message"`
}
