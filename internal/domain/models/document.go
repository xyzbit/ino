package models

import (
	"time"
)

// DocumentStatus 文档状态
type DocumentStatus string

const (
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusCompleted  DocumentStatus = "completed"
	DocumentStatusFailed     DocumentStatus = "failed"
)

// Document 文档模型
type Document struct {
	ID          uint64                 `json:"id"`
	DocumentID  string                 `json:"document_id"`
	Domain      string                 `json:"domain"`
	Title       string                 `json:"title"`
	Summary     string                 `json:"summary"`
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	FilePath    string                 `json:"file_path"`
	FileSize    int64                  `json:"file_size"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	Status      DocumentStatus         `json:"status"`
	ChunksCount int                    `json:"chunks_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TODO: 优化描述格式(1. GetPromptDesc() 2. openapi 格式给文档，给出可能的值等等)
func (d *Document) GetDescription() string {
	return `
	{
		"domain": "coding", // 文档领域, 如：电商、金融、教育、医疗、编程等
		"title": "文档标题", // 文档标题
		"summary": "文档摘要", // 文档摘要
		"content_type": "markdown", // 文档内容类型, markdown|pdf|word|excel|ppt|txt|image|video|audio
		"content": "文档内容", // 文档内容
		"file_path": "https://example.com/document.txt", // 文档文件路径
		"file_size": 1024, // 文档文件大小, 单位: 字节
		"metadata": { // 文档元数据, 可选
			"author": "张三", // 文档作者
			"subject": "文档主题", // 文档主题
			"keywords": ["关键词1", "关键词2"], // 文档关键词
		},
		"tags": ["tag1", "tag2"], // 文档标签, 可选
	}
	`
}

// DocumentMetadata 文档元数据
type DocumentMetadata struct {
	Author       string    `json:"author"`
	Subject      string    `json:"subject"`
	Keywords     []string  `json:"keywords"`
	CreateDate   time.Time `json:"create_date"`
	ModifyDate   time.Time `json:"modify_date"`
	Language     string    `json:"language"`
	PageCount    int       `json:"page_count"`
	WordCount    int       `json:"word_count"`
	Format       string    `json:"format"`
	Encoding     string    `json:"encoding"`
	OriginalName string    `json:"original_name"`
}

// DocumentChunk 文档分块
type DocumentChunk struct {
	ID         uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	DocumentID string                 `json:"document_id" gorm:"index,type:varchar(64),not null"`
	ChunkID    string                 `json:"chunk_id" gorm:"uniqueIndex,type:varchar(64),not null"`
	Content    string                 `json:"content" gorm:"type:text"`
	StartPos   int                    `json:"start_pos"`
	EndPos     int                    `json:"end_pos"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:json"`
	Vector     []float32              `json:"-" gorm:"-"` // 向量数据存储在Milvus中
	CreatedAt  time.Time              `json:"created_at"`
}

// TableName 指定表名
func (DocumentChunk) TableName() string {
	return "document_chunks"
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	DomainID    uint64                 `json:"domain_id" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	ContentType string                 `json:"content_type"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateDocumentRequest 更新文档请求
type UpdateDocumentRequest struct {
	Title    string                 `json:"title"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DocumentResponse 文档响应
type DocumentResponse struct {
	ID          uint64                 `json:"id"`
	DocumentID  string                 `json:"document_id"`
	Domain      string                 `json:"domain"`
	Title       string                 `json:"title"`
	ContentType string                 `json:"content_type"`
	FilePath    string                 `json:"file_path"`
	FileSize    int64                  `json:"file_size"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	Status      DocumentStatus         `json:"status"`
	ChunksCount int                    `json:"chunks_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ToResponse 转换为响应格式
func (d *Document) ToResponse() *DocumentResponse {
	resp := &DocumentResponse{
		ID:          d.ID,
		DocumentID:  d.DocumentID,
		Domain:      d.Domain,
		Title:       d.Title,
		ContentType: d.ContentType,
		FilePath:    d.FilePath,
		FileSize:    d.FileSize,
		Metadata:    d.Metadata,
		Tags:        d.Tags,
		Status:      d.Status,
		ChunksCount: d.ChunksCount,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}

	return resp
}

// DocumentChunkResponse 文档分块响应
type DocumentChunkResponse struct {
	ID         uint64                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	ChunkID    string                 `json:"chunk_id"`
	Content    string                 `json:"content"`
	StartPos   int                    `json:"start_pos"`
	EndPos     int                    `json:"end_pos"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ToResponse 转换为响应格式
func (dc *DocumentChunk) ToResponse() *DocumentChunkResponse {
	return &DocumentChunkResponse{
		ID:         dc.ID,
		DocumentID: dc.DocumentID,
		ChunkID:    dc.ChunkID,
		Content:    dc.Content,
		StartPos:   dc.StartPos,
		EndPos:     dc.EndPos,
		Metadata:   dc.Metadata,
		CreatedAt:  dc.CreatedAt,
	}
}
