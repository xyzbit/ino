package neo4j

import (
	"context"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/xyzbit/ino/config"
)

var (
	Driver  neo4j.DriverWithContext
	Session neo4j.SessionWithContext
)

// Init 初始化Neo4j连接
func Init() {
	cfg := config.AppConfig.Neo4j

	// 创建驱动
	var err error
	Driver, err = neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionLifetime = 5 * time.Minute
			config.MaxConnectionPoolSize = 50
			config.ConnectionAcquisitionTimeout = 2 * time.Minute
		},
	)
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = Driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	log.Println("Connected to Neo4j successfully")

	// 创建会话
	Session = Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

// Close 关闭Neo4j连接
func Close() error {
	if Session != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Session.Close(ctx)
	}

	if Driver != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return Driver.Close(ctx)
	}

	return nil
}

// GetSession 获取新的Neo4j会话
func GetSession(ctx context.Context) neo4j.SessionWithContext {
	return Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

// GetReadSession 获取只读会话
func GetReadSession(ctx context.Context) neo4j.SessionWithContext {
	return Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
}
