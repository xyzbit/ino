package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/xyzbit/ino/internal/domain/repository"
)

// Repository Neo4j仓储管理器
type Repository struct {
	Graph repository.GraphRepository
}

// NewRepository 创建Neo4j仓储管理器
func NewRepository(driver neo4j.DriverWithContext) *Repository {
	return &Repository{
		Graph: NewGraphRepository(driver),
	}
}
