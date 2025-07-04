package models

import (
	"time"
)

// Domain 知识域模型
type Domain struct {
	ID          uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	DomainName  string                 `json:"domain_name" gorm:"uniqueIndex,size:100,not null"`
	Description string                 `json:"description" gorm:"type:text"`
	Config      map[string]interface{} `json:"config" gorm:"type:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TableName 指定表名
func (Domain) TableName() string {
	return "domains"
}

// DomainConfig 知识域配置
type DomainConfig struct {
	VectorDimension int               `json:"vector_dimension"` // 向量维度
	IndexType       string            `json:"index_type"`       // 索引类型
	MetricType      string            `json:"metric_type"`      // 相似度计算类型
	SearchParams    map[string]string `json:"search_params"`    // 搜索参数
	GraphConfig     GraphConfig       `json:"graph_config"`     // 图数据库配置
}

// GraphConfig 图数据库配置
type GraphConfig struct {
	NodeTypes     []string `json:"node_types"`     // 节点类型
	RelationTypes []string `json:"relation_types"` // 关系类型
	MaxDepth      int      `json:"max_depth"`      // 最大搜索深度
	MinScore      float64  `json:"min_score"`      // 最小相关度分数
}

// CreateDomainRequest 创建知识域请求
type CreateDomainRequest struct {
	DomainName  string                 `json:"domain_name" binding:"required"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// UpdateDomainRequest 更新知识域请求
type UpdateDomainRequest struct {
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// DomainResponse 知识域响应
type DomainResponse struct {
	ID          uint64                 `json:"id"`
	DomainName  string                 `json:"domain_name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ToResponse 转换为响应格式
func (d *Domain) ToResponse() *DomainResponse {
	return &DomainResponse{
		ID:          d.ID,
		DomainName:  d.DomainName,
		Description: d.Description,
		Config:      d.Config,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
