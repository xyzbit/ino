package repository

import (
	"context"

	"github.com/xyzbit/ino/internal/domain/models"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	GetByUserID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
}

// DocumentRepository 文档仓储接口
type DocumentRepository interface {
	Create(ctx context.Context, document *models.Document) error
	GetByID(ctx context.Context, id uint64) (*models.Document, error)
	GetByDocumentID(ctx context.Context, documentID string) (*models.Document, error)
	Update(ctx context.Context, document *models.Document) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.Document, error)
	ListByDomain(ctx context.Context, domainID uint64, offset, limit int) ([]*models.Document, error)
	ListByStatus(ctx context.Context, status models.DocumentStatus, offset, limit int) ([]*models.Document, error)
	Count(ctx context.Context) (int64, error)
	CountByDomain(ctx context.Context, domainID uint64) (int64, error)
	CountByStatus(ctx context.Context, status models.DocumentStatus) (int64, error)
}

// DocumentChunkRepository 文档分块仓储接口
type DocumentChunkRepository interface {
	Create(ctx context.Context, chunk *models.DocumentChunk) error
	GetByID(ctx context.Context, id uint64) (*models.DocumentChunk, error)
	GetByChunkID(ctx context.Context, chunkID string) (*models.DocumentChunk, error)
	Update(ctx context.Context, chunk *models.DocumentChunk) error
	Delete(ctx context.Context, id uint64) error
	ListByDocument(ctx context.Context, documentID string, offset, limit int) ([]*models.DocumentChunk, error)
	CountByDocument(ctx context.Context, documentID string) (int64, error)
	BatchCreate(ctx context.Context, chunks []*models.DocumentChunk) error
	BatchDelete(ctx context.Context, documentID string) error
}

// FeedbackRepository 反馈仓储接口
type FeedbackRepository interface {
	Create(ctx context.Context, feedback *models.Feedback) error
	GetByID(ctx context.Context, id uint64) (*models.Feedback, error)
	GetByQueryID(ctx context.Context, queryID string) ([]*models.Feedback, error)
	Update(ctx context.Context, feedback *models.Feedback) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.Feedback, error)
	ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*models.Feedback, error)
	ListByType(ctx context.Context, feedbackType models.FeedbackType, offset, limit int) ([]*models.Feedback, error)
	Count(ctx context.Context) (int64, error)
	CountByType(ctx context.Context, feedbackType models.FeedbackType) (int64, error)
}

// CacheRepository 缓存仓储接口
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl int) error
	Keys(ctx context.Context, pattern string) ([]string, error)

	// 哈希操作
	HSet(ctx context.Context, key, field string, value interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// 列表操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// 集合操作
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...interface{}) error
	SCard(ctx context.Context, key string) (int64, error)
}

// Repository 仓储管理器
type Repository struct {
	User          UserRepository
	Document      DocumentRepository
	DocumentChunk DocumentChunkRepository
	Feedback      FeedbackRepository
	Cache         CacheRepository
}
