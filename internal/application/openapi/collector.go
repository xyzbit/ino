package openapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/internal/domain/services"
)

type OpenAPI struct {
	requestOptimizer *services.RequestOptimizer
}

func NewOpenAPI() *OpenAPI {
	optimizer, err := services.NewRequestOptimizer()
	if err != nil {
		panic(err)
	}

	return &OpenAPI{
		requestOptimizer: optimizer,
	}
}

// CollectKnowledge 收集知识.
func (o *OpenAPI) CollectKnowledge(c *gin.Context) {
	var (
		req *types.CollectKnowledgeRequest
		err error
	)
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = validateCollectKnowledgeRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// optimize request if enabled.
	req, err = o.requestOptimizer.Exec(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "内容优化失败: " + err.Error(),
		})
		return
	}

	jsonData, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("CollectKnowledge request: %s", string(jsonData))

	c.JSON(http.StatusOK, types.CollectKnowledgeResponse{
		Message: "知识收集成功",
	})
}

func validateCollectKnowledgeRequest(req *types.CollectKnowledgeRequest) error {
	if req.CollectionID == "" {
		return errors.New("collection_id is required")
	}
	if req.Content == "" {
		return errors.New("content is required")
	}
	return nil
}
