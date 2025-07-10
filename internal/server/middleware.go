package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/pkg/ctxwarp"
)

func MiddlewareSetHeaderContext(c *gin.Context) {
	header := &types.CollectKnowledgeRequestHeader{}
	if err := c.ShouldBindHeader(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if header.RequestID == "" {
		header.RequestID = uuid.New().String()
	}

	ctx := c.Request.Context()
	ctx = ctxwarp.SetHeaderContext(ctx, &ctxwarp.HeaderContext{
		RequestID:    header.RequestID,
		CollectionID: header.CollectionID,
		User:         header.User,
	})
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
