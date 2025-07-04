package mysql

import (
	"context"

	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/internal/domain/repository"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档仓储实例
func NewDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

// Create 创建文档
func (r *documentRepository) Create(ctx context.Context, document *models.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

// GetByID 根据ID获取文档
func (r *documentRepository) GetByID(ctx context.Context, id uint64) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).Preload("Domain").First(&document, id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// GetByDocumentID 根据文档ID获取文档
func (r *documentRepository) GetByDocumentID(ctx context.Context, documentID string) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).
		Preload("Domain").
		Where("document_id = ?", documentID).
		First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// Update 更新文档
func (r *documentRepository) Update(ctx context.Context, document *models.Document) error {
	return r.db.WithContext(ctx).Save(document).Error
}

// Delete 删除文档
func (r *documentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Document{}, id).Error
}

// List 获取文档列表
func (r *documentRepository) List(ctx context.Context, offset, limit int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Preload("Domain").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&documents).Error
	return documents, err
}

// ListByDomain 根据知识域获取文档列表
func (r *documentRepository) ListByDomain(ctx context.Context, domainID uint64, offset, limit int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Preload("Domain").
		Where("domain_id = ?", domainID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&documents).Error
	return documents, err
}

// ListByStatus 根据状态获取文档列表
func (r *documentRepository) ListByStatus(ctx context.Context, status models.DocumentStatus, offset, limit int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Preload("Domain").
		Where("status = ?", status).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&documents).Error
	return documents, err
}

// Count 获取文档总数
func (r *documentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Document{}).Count(&count).Error
	return count, err
}

// CountByDomain 根据知识域获取文档总数
func (r *documentRepository) CountByDomain(ctx context.Context, domainID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("domain_id = ?", domainID).
		Count(&count).Error
	return count, err
}

// CountByStatus 根据状态获取文档总数
func (r *documentRepository) CountByStatus(ctx context.Context, status models.DocumentStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}
