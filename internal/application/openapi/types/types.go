package types

import "encoding/json"

type CollectKnowledgeRequest struct {
	Domain      string            `json:"domain,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Content     string            `json:"content,omitempty"`
}

func (c *CollectKnowledgeRequest) GetDescription() string {
	desc := []map[string]interface{}{
		{
			"field_name":  "domain",
			"field_type":  "string",
			"description": "可选，知识内容涉及的领域或业务, 如：电商、金融、教育、医疗、编程等",
			"value":       c.Domain,
		},
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
