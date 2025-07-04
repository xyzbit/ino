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
	ID           uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	QueryID      string                 `json:"query_id" gorm:"type:varchar(64),not null"`
	UserID       uint64                 `json:"user_id" gorm:"not null"`
	User         *User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	FeedbackType FeedbackType           `json:"feedback_type" gorm:"not null"`
	Rating       int                    `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"`
	Comment      string                 `json:"comment" gorm:"type:text"`
	Context      map[string]interface{} `json:"context" gorm:"type:json"`
	CreatedAt    time.Time              `json:"created_at"`
}

// TableName 指定表名
func (Feedback) TableName() string {
	return "feedback"
}

// FeedbackContext 反馈上下文
type FeedbackContext struct {
	Query          string                 `json:"query"`
	QueryTime      int                    `json:"query_time"`      // 查询时间（毫秒）
	ResultsCount   int                    `json:"results_count"`   // 结果数量
	ResultsShown   int                    `json:"results_shown"`   // 展示的结果数
	ClickedResults []string               `json:"clicked_results"` // 点击的结果ID
	UserAgent      string                 `json:"user_agent"`
	IPAddress      string                 `json:"ip_address"`
	SessionID      string                 `json:"session_id"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

// CollectFeedbackRequest 收集反馈请求
type CollectFeedbackRequest struct {
	QueryID      string                 `json:"query_id" binding:"required"`
	UserID       uint64                 `json:"user_id" binding:"required"`
	FeedbackType FeedbackType           `json:"feedback_type" binding:"required"`
	Rating       int                    `json:"rating" binding:"min=1,max=5"`
	Comment      string                 `json:"comment"`
	Context      map[string]interface{} `json:"context"`
}

// UpdateFeedbackRequest 更新反馈请求
type UpdateFeedbackRequest struct {
	FeedbackType FeedbackType           `json:"feedback_type"`
	Rating       int                    `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment      string                 `json:"comment"`
	Context      map[string]interface{} `json:"context"`
}

// FeedbackResponse 反馈响应
type FeedbackResponse struct {
	ID           uint64                 `json:"id"`
	QueryID      string                 `json:"query_id"`
	UserID       uint64                 `json:"user_id"`
	User         *UserResponse          `json:"user,omitempty"`
	FeedbackType FeedbackType           `json:"feedback_type"`
	Rating       int                    `json:"rating"`
	Comment      string                 `json:"comment"`
	Context      map[string]interface{} `json:"context"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ToResponse 转换为响应格式
func (f *Feedback) ToResponse() *FeedbackResponse {
	resp := &FeedbackResponse{
		ID:           f.ID,
		QueryID:      f.QueryID,
		UserID:       f.UserID,
		FeedbackType: f.FeedbackType,
		Rating:       f.Rating,
		Comment:      f.Comment,
		Context:      f.Context,
		CreatedAt:    f.CreatedAt,
	}

	if f.User != nil {
		resp.User = f.User.ToResponse()
	}

	return resp
}

// FeedbackStats 反馈统计
type FeedbackStats struct {
	TotalFeedback int     `json:"total_feedback"`
	PositiveCount int     `json:"positive_count"`
	NegativeCount int     `json:"negative_count"`
	NeutralCount  int     `json:"neutral_count"`
	PositiveRate  float64 `json:"positive_rate"`
	NegativeRate  float64 `json:"negative_rate"`
	AverageRating float64 `json:"average_rating"`
	TopIssues     []struct {
		Issue string `json:"issue"`
		Count int    `json:"count"`
	} `json:"top_issues"`
	TrendData []struct {
		Date          time.Time `json:"date"`
		PositiveCount int       `json:"positive_count"`
		NegativeCount int       `json:"negative_count"`
		AverageRating float64   `json:"average_rating"`
	} `json:"trend_data"`
}
