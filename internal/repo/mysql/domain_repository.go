package mysql

import (
	"context"

	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/internal/domain/repository"
	"gorm.io/gorm"
)

type domainRepository struct {
	db *gorm.DB
}

// NewDomainRepository 创建知识域仓储实例
func NewDomainRepository(db *gorm.DB) repository.DomainRepository {
	return &domainRepository{db: db}
}

// Create 创建知识域
func (r *domainRepository) Create(ctx context.Context, domain *models.Domain) error {
	return r.db.WithContext(ctx).Create(domain).Error
}

// GetByID 根据ID获取知识域
func (r *domainRepository) GetByID(ctx context.Context, id uint64) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.WithContext(ctx).First(&domain, id).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetByName 根据名称获取知识域
func (r *domainRepository) GetByName(ctx context.Context, name string) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.WithContext(ctx).Where("domain_name = ?", name).First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// Update 更新知识域
func (r *domainRepository) Update(ctx context.Context, domain *models.Domain) error {
	return r.db.WithContext(ctx).Save(domain).Error
}

// Delete 删除知识域
func (r *domainRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Domain{}, id).Error
}

// List 获取知识域列表
func (r *domainRepository) List(ctx context.Context, offset, limit int) ([]*models.Domain, error) {
	var domains []*models.Domain
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&domains).Error
	return domains, err
}

// Count 获取知识域总数
func (r *domainRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Domain{}).Count(&count).Error
	return count, err
}
