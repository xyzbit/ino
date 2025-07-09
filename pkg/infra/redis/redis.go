package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	Redis *redis.Client
)

// InitRedis 初始化Redis连接
func Init() {
	host := viper.GetString("redis.host")
	port := viper.GetInt("redis.port")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")

	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// 测试Redis连接
	_, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		// 在开发阶段先不Fatal，允许不连接Redis启动
	} else {
		log.Printf("Connected to Redis successfully")
	}
}

// Close 关闭数据库连接
func Close() {
	if Redis != nil {
		Redis.Close()
	}
}
