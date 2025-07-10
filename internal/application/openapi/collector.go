package openapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/internal/domain/services"
)

type OpenAPI struct {
	indexer *services.Indexer
}

func NewOpenAPI() *OpenAPI {
	indexer, err := services.NewIndexer()
	if err != nil {
		panic(err)
	}
	return &OpenAPI{
		indexer: indexer,
	}
}

// CollectKnowledge 收集知识.
func (o *OpenAPI) CollectKnowledge(c *gin.Context) {
	var (
		req *types.CollectKnowledgeRequest
		err error
	)
	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = validateCollectKnowledgeRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = o.indexer.Exec(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.CollectKnowledgeResponse{
		Message: "知识收集成功",
	})
}

func validateCollectKnowledgeRequest(req *types.CollectKnowledgeRequest) error {
	if req.Content == "" && req.ContentLink == "" {
		return errors.New("content or content_link is required")
	}
	return nil
}
