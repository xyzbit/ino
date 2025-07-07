package models

import (
	"time"
)

// FeedbackType 反馈类型
type FeedbackType string

const (
	FeedbackTypePositive FeedbackType = "positive"
	FeedbackTypeNegative FeedbackType = "negative"
	FeedbackTypeNeutral  FeedbackType = "neutral"
)

// Feedback 反馈模型
type Feedback struct {
	ID           uint64       `json:"id"`
	UserID       string       `json:"user_id"`
	FeedbackType FeedbackType `json:"feedback_type"`
	Rating       int          `json:"rating"`
	Comment      string       `json:"comment"`
	Context      string       `json:"context"`
	CreatedAt    time.Time    `json:"created_at"`
}

func (f *Feedback) GetDescription() string {
	return `{
		"user_id": 1234567 // 用户ID
		"feedback_type": "positive", // 反馈类型, 可选值: positive, negative, neutral
		"rating": 1, // 评分, 1-5
		"comment": "反馈内容", // 反馈内容
		"context": "反馈上下文" // 反馈上下文
	}`
}
