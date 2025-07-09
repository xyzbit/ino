package infra

import (
	"log"

	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/pkg/infra/milvus"
	"github.com/xyzbit/ino/pkg/infra/mysql"
	"github.com/xyzbit/ino/pkg/infra/neo4j"
	"github.com/xyzbit/ino/pkg/infra/redis"
)

// Init 初始化所有基础设施
func Init() {
	// 初始化数据库连接
	mysql.Init()
	redis.Init()
	milvus.Init()
	neo4j.Init()

	// 初始化种子数据
	if err := models.SeedData(mysql.DB); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}
}

// Close 关闭所有连接
func Close() error {
	mysql.Close()
	redis.Close()
	milvus.Close()
	neo4j.Close()
	return nil
}
