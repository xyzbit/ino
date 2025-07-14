package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/manager"
	"github.com/xyzbit/ino/internal/application/openapi"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, openAPI *openapi.OpenAPI, version string) {
	// 添加全局日志中间件
	r.Use(middlewareCheckError)

	// 健康检查接口
	r.GET("/health", healthCheck(version))

	// API版本1
	v1 := r.Group("/api/v1")
	{

		api := v1.Group("/openapi", middlewareSetHeaderContext)
		{
			// 知识收集接口
			api.POST("/collect", openAPI.CollectKnowledge)
			// 知识查询接口
			api.POST("/search", openAPI.SearchKnowledge)
		}

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
