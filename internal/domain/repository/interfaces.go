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

// DomainRepository 知识域仓储接口
type DomainRepository interface {
	Create(ctx context.Context, domain *models.Domain) error
	GetByID(ctx context.Context, id uint64) (*models.Domain, error)
	GetByName(ctx context.Context, name string) (*models.Domain, error)
	Update(ctx context.Context, domain *models.Domain) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.Domain, error)
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

// ConversationRepository 对话仓储接口
type ConversationRepository interface {
	Create(ctx context.Context, conversation *models.Conversation) error
	GetByID(ctx context.Context, id uint64) (*models.Conversation, error)
	GetByConversationID(ctx context.Context, conversationID string) (*models.Conversation, error)
	Update(ctx context.Context, conversation *models.Conversation) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.Conversation, error)
	ListByDomain(ctx context.Context, domainID uint64, offset, limit int) ([]*models.Conversation, error)
	ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*models.Conversation, error)
	Count(ctx context.Context) (int64, error)
	CountByDomain(ctx context.Context, domainID uint64) (int64, error)
	CountByUser(ctx context.Context, userID uint64) (int64, error)
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
	GetStats(ctx context.Context) (*models.FeedbackStats, error)
}

// SearchLogRepository 搜索日志仓储接口
type SearchLogRepository interface {
	Create(ctx context.Context, log *models.SearchLog) error
	GetByID(ctx context.Context, id uint64) (*models.SearchLog, error)
	GetByQueryID(ctx context.Context, queryID string) (*models.SearchLog, error)
	Update(ctx context.Context, log *models.SearchLog) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*models.SearchLog, error)
	ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*models.SearchLog, error)
	ListByDomain(ctx context.Context, domainID uint64, offset, limit int) ([]*models.SearchLog, error)
	Count(ctx context.Context) (int64, error)
	GetStats(ctx context.Context) (*models.SearchStats, error)
}

// VectorRepository 向量数据库仓储接口
type VectorRepository interface {
	// 集合管理
	CreateCollection(ctx context.Context, collectionName string, dimension int) error
	DropCollection(ctx context.Context, collectionName string) error
	HasCollection(ctx context.Context, collectionName string) (bool, error)

	// 数据操作
	Insert(ctx context.Context, collectionName string, vectors []VectorData) error
	Delete(ctx context.Context, collectionName string, ids []string) error
	Update(ctx context.Context, collectionName string, vectors []VectorData) error

	// 搜索
	Search(ctx context.Context, collectionName string, vectors [][]float32, topK int, params map[string]interface{}) ([]VectorSearchResult, error)

	// 索引管理
	CreateIndex(ctx context.Context, collectionName string, params map[string]interface{}) error
	DropIndex(ctx context.Context, collectionName string) error

	// 统计
	GetCollectionStats(ctx context.Context, collectionName string) (*VectorCollectionStats, error)
}

// GraphRepository 图数据库仓储接口
type GraphRepository interface {
	// 实体操作
	CreateEntity(ctx context.Context, entity *models.KnowledgeEntity) error
	GetEntity(ctx context.Context, id string) (*models.KnowledgeEntity, error)
	UpdateEntity(ctx context.Context, entity *models.KnowledgeEntity) error
	DeleteEntity(ctx context.Context, id string) error
	ListEntities(ctx context.Context, entityType string, offset, limit int) ([]*models.KnowledgeEntity, error)

	// 关系操作
	CreateRelation(ctx context.Context, relation *models.KnowledgeRelation) error
	GetRelation(ctx context.Context, id string) (*models.KnowledgeRelation, error)
	UpdateRelation(ctx context.Context, relation *models.KnowledgeRelation) error
	DeleteRelation(ctx context.Context, id string) error
	ListRelations(ctx context.Context, fromEntity, toEntity string, relationType string) ([]*models.KnowledgeRelation, error)

	// 图遍历
	TraverseGraph(ctx context.Context, config *models.GraphTraversal) (*models.GraphTraversalResult, error)
	FindPath(ctx context.Context, fromEntity, toEntity string, maxDepth int) ([]*models.GraphPath, error)

	// 图搜索
	SearchEntities(ctx context.Context, query string, entityTypes []string, limit int) ([]*models.KnowledgeEntity, error)

	// 统计
	GetGraphStats(ctx context.Context) (*models.GraphStats, error)
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

// 向量数据结构
type VectorData struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

type VectorSearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type VectorCollectionStats struct {
	RowCount     int64 `json:"row_count"`
	IndexedCount int64 `json:"indexed_count"`
	MemorySize   int64 `json:"memory_size"`
	DiskSize     int64 `json:"disk_size"`
}

// Repository 仓储管理器
type Repository struct {
	User          UserRepository
	Domain        DomainRepository
	Document      DocumentRepository
	DocumentChunk DocumentChunkRepository
	Conversation  ConversationRepository
	Feedback      FeedbackRepository
	SearchLog     SearchLogRepository
	Vector        VectorRepository
	Graph         GraphRepository
	Cache         CacheRepository
}
