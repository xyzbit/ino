package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/infra"
	"github.com/xyzbit/ino/internal/server"
)

func main() {
	// 加载配置
	config.Init()

	// 设置Gin模式
	if config.AppConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化基础设施
	infra.Init()
	defer infra.Close()

	// 创建路由
	r := gin.Default()

	// 注册路由
	server.RegisterRoutes(r, "v1.0.0")

	// 创建服务器
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Server.Port,
		Handler: r,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
