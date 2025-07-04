package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/collector"
	"github.com/xyzbit/ino/internal/application/manager"
	"github.com/xyzbit/ino/internal/application/search"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, version string) {
	// 健康检查接口
	r.GET("/health", healthCheck(version))

	// API版本1
	v1 := r.Group("/api/v1")
	{
		// 知识收集接口
		knowledge := v1.Group("/collect")
		{
			knowledge.POST("/document", collector.UploadDocument)
			knowledge.POST("/conversation", collector.CollectConversation)
			knowledge.POST("/feedback", collector.CollectFeedback)
		}

		// 知识查询接口
		knowledge.POST("/search", search.SearchKnowledge)

		// 管理接口
		admin := v1.Group("/admin")
		{
			admin.GET("/stats", manager.GetStats)
			admin.GET("/users", manager.GetUsers)
			admin.POST("/users", manager.CreateUser)
		}
	}
}

// healthCheck 健康检查
func healthCheck(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "INO Knowledge System",
			"version": version,
		})
	}
}
