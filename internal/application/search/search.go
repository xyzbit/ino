package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchKnowledge(c *gin.Context) {
	// TODO: 实现知识搜索逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Search knowledge endpoint - TODO",
	})
}
