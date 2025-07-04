package models

import (
	"time"
)

// Conversation 对话记录模型
type Conversation struct {
	ID             uint64              `json:"id" gorm:"primaryKey,autoIncrement"`
	ConversationID string              `json:"conversation_id" gorm:"uniqueIndex,type:varchar(64),not null"`
	DomainID       uint64              `json:"domain_id" gorm:"not null"`
	Domain         *Domain             `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
	UserID         uint64              `json:"user_id"`
	User           *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Content        ConversationContent `json:"content" gorm:"type:json"`
	Tags           []string            `json:"tags" gorm:"type:json"`
	ProcessedAt    *time.Time          `json:"processed_at"`
	CreatedAt      time.Time           `json:"created_at"`
}

// TableName 指定表名
func (Conversation) TableName() string {
	return "conversations"
}

// ConversationContent 对话内容
type ConversationContent struct {
	Messages []Message `json:"messages"`
	Context  Context   `json:"context"`
}

// Message 消息
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata"`
}

// Context 对话上下文
type Context struct {
	SessionID    string                 `json:"session_id"`
	Platform     string                 `json:"platform"` // web, mobile, api
	Source       string                 `json:"source"`   // 来源系统
	UserAgent    string                 `json:"user_agent"`
	IPAddress    string                 `json:"ip_address"`
	Location     string                 `json:"location"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

// Metadata 消息元数据
type Metadata struct {
	TokenCount     int                    `json:"token_count"`
	Model          string                 `json:"model"`
	Temperature    float64                `json:"temperature"`
	MaxTokens      int                    `json:"max_tokens"`
	ResponseTime   int                    `json:"response_time"` // 毫秒
	RetrievalInfo  RetrievalInfo          `json:"retrieval_info"`
	CustomMetadata map[string]interface{} `json:"custom_metadata"`
}

// RetrievalInfo 检索信息
type RetrievalInfo struct {
	QueryTime       int      `json:"query_time"`       // 查询时间（毫秒）
	DocumentsFound  int      `json:"documents_found"`  // 找到的文档数
	ChunksUsed      int      `json:"chunks_used"`      // 使用的分块数
	RetrievalMethod string   `json:"retrieval_method"` // 检索方法
	ScoreThreshold  float64  `json:"score_threshold"`  // 分数阈值
	Sources         []string `json:"sources"`          // 来源文档
}

// CollectConversationRequest 收集对话请求
type CollectConversationRequest struct {
	ConversationID string              `json:"conversation_id" binding:"required"`
	DomainID       uint64              `json:"domain_id" binding:"required"`
	UserID         uint64              `json:"user_id"`
	Content        ConversationContent `json:"content" binding:"required"`
	Tags           []string            `json:"tags"`
}

// UpdateConversationRequest 更新对话请求
type UpdateConversationRequest struct {
	Content ConversationContent `json:"content"`
	Tags    []string            `json:"tags"`
}

// ConversationResponse 对话响应
type ConversationResponse struct {
	ID             uint64              `json:"id"`
	ConversationID string              `json:"conversation_id"`
	DomainID       uint64              `json:"domain_id"`
	Domain         *DomainResponse     `json:"domain,omitempty"`
	UserID         uint64              `json:"user_id"`
	User           *UserResponse       `json:"user,omitempty"`
	Content        ConversationContent `json:"content"`
	Tags           []string            `json:"tags"`
	ProcessedAt    *time.Time          `json:"processed_at"`
	CreatedAt      time.Time           `json:"created_at"`
}

// ToResponse 转换为响应格式
func (c *Conversation) ToResponse() *ConversationResponse {
	resp := &ConversationResponse{
		ID:             c.ID,
		ConversationID: c.ConversationID,
		DomainID:       c.DomainID,
		UserID:         c.UserID,
		Content:        c.Content,
		Tags:           c.Tags,
		ProcessedAt:    c.ProcessedAt,
		CreatedAt:      c.CreatedAt,
	}

	if c.Domain != nil {
		resp.Domain = c.Domain.ToResponse()
	}

	if c.User != nil {
		resp.User = c.User.ToResponse()
	}

	return resp
}

// ConversationStats 对话统计
type ConversationStats struct {
	TotalConversations int     `json:"total_conversations"`
	TotalMessages      int     `json:"total_messages"`
	AvgMessagesPerConv float64 `json:"avg_messages_per_conversation"`
	TotalUsers         int     `json:"total_users"`
	ActiveUsers        int     `json:"active_users"`
	TopDomains         []struct {
		DomainName string `json:"domain_name"`
		Count      int    `json:"count"`
	} `json:"top_domains"`
}
