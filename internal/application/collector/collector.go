package collector

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// uploadDocument 上传文档
func UploadDocument(c *gin.Context) {
	// TODO: 实现文档上传逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Document upload endpoint - TODO",
	})
}

// collectConversation 收集对话
func CollectConversation(c *gin.Context) {
	// TODO: 实现对话收集逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Conversation collection endpoint - TODO",
	})
}

// collectFeedback 收集反馈
func CollectFeedback(c *gin.Context) {
	// TODO: 实现反馈收集逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Feedback collection endpoint - TODO",
	})
}
