package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/openapi/types"
)

func (o *OpenAPI) SearchKnowledge(c *gin.Context) {
	var (
		req *types.RetrieveRequest
		err error
	)
	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := o.retriever.Exec(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
