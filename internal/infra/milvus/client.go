package milvus

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/xyzbit/ino/config"
)

var Client client.Client

// Init 初始化Milvus连接
func Init() {
	cfg := config.AppConfig.Milvus

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = client.NewGrpcClient(ctx, addr)
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}

	// 测试连接
	version, err := Client.GetVersion(ctx)
	if err != nil {
		log.Fatalf("Failed to get Milvus version: %v", err)
	}

	log.Printf("Connected to Milvus successfully, version: %s", version)
}

// Close 关闭Milvus连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
