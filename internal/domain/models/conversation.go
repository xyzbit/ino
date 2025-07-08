package models

import (
	"time"
)

// Conversation 对话记录模型
type Conversation struct {
	ID             uint64 `json:"id" `
	ConversationID string `json:"conversation_id"`
	// AgentID        string    `json:"agent_id"`
	Domain    string    `json:"domain"`
	Tags      []string  `json:"tags"`
	Context   string    `json:"context"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Conversation) GetDescription() string {
	return `
	{
		"domain": "coding", // 对话的领域或业务, 如：电商、金融、教育、医疗、编程等
		"messages": [
			{
				"role": "user", // user, assistant, system
				"content": "您好，我想咨询一下订单退款的问题", // 对话内容
				"time": "2025-07-07T09:15:30Z" // 对话时间 格式：2025-07-07T09:15:30Z
			}
		],
		"context": "用户咨询ORD-20250701-12345订单的退款事宜，该订单于2025年7月1日创建，商品未发货", // 对话上下文,场景
		"tags": ["refund", "unshipped", "customer_service"],
		"created_at": "2025-07-07T09:15:00Z" // 对话创建时间 格式：2025-07-07T09:15:00Z
	}
	`
}
