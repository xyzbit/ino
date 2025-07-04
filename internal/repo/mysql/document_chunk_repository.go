package mysql

import (
	"context"

	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/internal/domain/repository"
	"gorm.io/gorm"
)

type documentChunkRepository struct {
	db *gorm.DB
}

// NewDocumentChunkRepository 创建文档分块仓储实例
func NewDocumentChunkRepository(db *gorm.DB) repository.DocumentChunkRepository {
	return &documentChunkRepository{db: db}
}

// Create 创建文档分块
func (r *documentChunkRepository) Create(ctx context.Context, chunk *models.DocumentChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

// GetByID 根据ID获取文档分块
func (r *documentChunkRepository) GetByID(ctx context.Context, id uint64) (*models.DocumentChunk, error) {
	var chunk models.DocumentChunk
	err := r.db.WithContext(ctx).First(&chunk, id).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// GetByChunkID 根据分块ID获取文档分块
func (r *documentChunkRepository) GetByChunkID(ctx context.Context, chunkID string) (*models.DocumentChunk, error) {
	var chunk models.DocumentChunk
	err := r.db.WithContext(ctx).Where("chunk_id = ?", chunkID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// Update 更新文档分块
func (r *documentChunkRepository) Update(ctx context.Context, chunk *models.DocumentChunk) error {
	return r.db.WithContext(ctx).Save(chunk).Error
}

// Delete 删除文档分块
func (r *documentChunkRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.DocumentChunk{}, id).Error
}

// ListByDocument 根据文档ID获取分块列表
func (r *documentChunkRepository) ListByDocument(ctx context.Context, documentID string, offset, limit int) ([]*models.DocumentChunk, error) {
	var chunks []*models.DocumentChunk
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("start_pos ASC").
		Offset(offset).
		Limit(limit).
		Find(&chunks).Error
	return chunks, err
}

// CountByDocument 根据文档ID获取分块总数
func (r *documentChunkRepository) CountByDocument(ctx context.Context, documentID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.DocumentChunk{}).
		Where("document_id = ?", documentID).
		Count(&count).Error
	return count, err
}

// BatchCreate 批量创建文档分块
func (r *documentChunkRepository) BatchCreate(ctx context.Context, chunks []*models.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error
}

// BatchDelete 批量删除文档分块
func (r *documentChunkRepository) BatchDelete(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&models.DocumentChunk{}).Error
}
