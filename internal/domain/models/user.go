package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID          uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	UserID      string                 `json:"user_id" gorm:"uniqueIndex,type:varchar(64),not null"`
	Username    string                 `json:"username" gorm:"size:100,not null"`
	Email       string                 `json:"email" gorm:"size:255,index"`
	AvatarURL   string                 `json:"avatar_url" gorm:"size:500"`
	Preferences map[string]interface{} `json:"preferences" gorm:"type:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserPreferences 用户偏好设置
type UserPreferences struct {
	Theme            string   `json:"theme"`             // 主题: light, dark
	Language         string   `json:"language"`          // 语言: zh, en
	NotificationTime string   `json:"notification_time"` // 通知时间
	FavoriteDomains  []string `json:"favorite_domains"`  // 收藏的知识域
	SearchHistory    []string `json:"search_history"`    // 搜索历史
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Username    string                 `json:"username" binding:"required"`
	Email       string                 `json:"email" binding:"email"`
	AvatarURL   string                 `json:"avatar_url"`
	Preferences map[string]interface{} `json:"preferences"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username    string                 `json:"username"`
	Email       string                 `json:"email" binding:"omitempty,email"`
	AvatarURL   string                 `json:"avatar_url"`
	Preferences map[string]interface{} `json:"preferences"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID          uint64                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	AvatarURL   string                 `json:"avatar_url"`
	Preferences map[string]interface{} `json:"preferences"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		UserID:      u.UserID,
		Username:    u.Username,
		Email:       u.Email,
		AvatarURL:   u.AvatarURL,
		Preferences: u.Preferences,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
