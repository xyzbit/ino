package mysql

import (
	"github.com/xyzbit/ino/internal/domain/repository"
	"gorm.io/gorm"
)

// NewRepository 创建MySQL仓储管理器
func NewRepository(db *gorm.DB) *repository.Repository {
	return &repository.Repository{
		User:          NewUserRepository(db),
		Domain:        NewDomainRepository(db),
		Document:      NewDocumentRepository(db),
		DocumentChunk: NewDocumentChunkRepository(db),
		// 其他仓储将在后续添加
	}
}
