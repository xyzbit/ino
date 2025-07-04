package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/infra"
	"github.com/xyzbit/ino/internal/server"
)

var version string

func main() {
	log.Printf("Server starting version: %s", version)
	// 初始化配置
	config.Init()

	// 初始化数据库
	infra.Init()

	// 初始化路由
	r := gin.Default()

	// 注册路由
	server.RegisterRoutes(r, version)

	// 启动服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
