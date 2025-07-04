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
	ID          uint64                 `json:"id" gorm:"primaryKey,autoIncrement"`
	DocumentID  string                 `json:"document_id" gorm:"uniqueIndex,type:varchar(64),not null"`
	DomainID    uint64                 `json:"domain_id" gorm:"not null"`
	Domain      *Domain                `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
	Title       string                 `json:"title" gorm:"size:500,not null"`
	ContentType string                 `json:"content_type" gorm:"size:50"`
	FilePath    string                 `json:"file_path" gorm:"size:1000"`
	FileSize    int64                  `json:"file_size"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:json"`
	Tags        []string               `json:"tags" gorm:"type:json"`
	Status      DocumentStatus         `json:"status" gorm:"default:processing"`
	ChunksCount int                    `json:"chunks_count" gorm:"default:0"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
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
	DomainID    uint64                 `json:"domain_id"`
	Domain      *DomainResponse        `json:"domain,omitempty"`
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
		DomainID:    d.DomainID,
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

	if d.Domain != nil {
		resp.Domain = d.Domain.ToResponse()
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
