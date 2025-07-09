package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xyzbit/ino/internal/application/openapi/types"
)

type headerContextKey struct{}

func setHeaderContext(c *gin.Context) {
	header := &types.CollectKnowledgeRequestHeader{}
	if err := c.ShouldBindHeader(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if header.RequestID == "" {
		header.RequestID = uuid.New().String()
	}
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, headerContextKey{}, header)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

func GetHeaderContext(ctx context.Context) *types.CollectKnowledgeRequestHeader {
	header := ctx.Value(headerContextKey{})
	if header == nil {
		return nil
	}
	return header.(*types.CollectKnowledgeRequestHeader)
}
