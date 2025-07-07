package infra

import (
	"log"

	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/internal/domain/repository"
	"github.com/xyzbit/ino/internal/infra/milvus"
	"github.com/xyzbit/ino/internal/infra/mysql"
	"github.com/xyzbit/ino/internal/infra/neo4j"
	"github.com/xyzbit/ino/internal/infra/redis"
	milvusRepo "github.com/xyzbit/ino/internal/repo/milvus"
	mysqlRepo "github.com/xyzbit/ino/internal/repo/mysql"
	neo4jRepo "github.com/xyzbit/ino/internal/repo/neo4j"
)

var (
	// Repository 全局仓储实例
	Repository *repository.Repository
)

// Init 初始化所有基础设施
func Init() {
	// 初始化数据库连接
	mysql.Init()
	redis.Init()
	milvus.Init()
	neo4j.Init()

	// 初始化仓储层
	initRepositories()

	// 初始化种子数据
	if err := models.SeedData(mysql.DB); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}
}

// initRepositories 初始化仓储层
func initRepositories() {
	// 创建各个仓储实例
	mysqlRepos := mysqlRepo.NewRepository(mysql.DB)
	vectorRepo, _ := milvusRepo.NewVectorRepository(milvus.Client)
	neo4jRepos := neo4jRepo.NewRepository(neo4j.Driver)

	// 组装总仓储
	Repository = &repository.Repository{
		User:          mysqlRepos.User,
		Domain:        mysqlRepos.Domain,
		Document:      mysqlRepos.Document,
		DocumentChunk: mysqlRepos.DocumentChunk,
		Vector:        vectorRepo,
		Graph:         neo4jRepos.Graph,
		// Cache:         redisRepos.Cache, // TODO: 需要实现Redis仓储
	}

	log.Println("Repositories initialized successfully")
}

// Close 关闭所有连接
func Close() error {
	mysql.Close()
	redis.Close()
	milvus.Close()
	neo4j.Close()
	return nil
}
