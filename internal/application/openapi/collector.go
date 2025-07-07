package openapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/internal/domain/services"
	"github.com/xyzbit/ino/pkg/constants"
)

type OpenAPI struct {
	requestOptimizer *services.RequestOptimizer
}

func NewOpenAPI() *OpenAPI {
	requestOptimizer, err := services.NewRequestOptimizer()
	if err != nil {
		log.Fatalf("failed to create request optimizer: %v", err)
	}
	return &OpenAPI{
		requestOptimizer: requestOptimizer,
	}
}

/*
CollectKnowledge 收集知识.

example:

	{
	  "domain": "code-review",
	  "tags": {
	    "project": "ino-system",
	    "type": "code-review"
	  },
	  "content_type": "auto | conversation | feedback | document", // 内容类型 auto 自动识别，conversation 对话，feedback 反馈，document 文档; 默认 auto
	  "content": "xx 行代码会导致 panic", // 如果 content_type = auto，则 content 必填.
	  "conversation": [ // 如果 content_type = conversation，则 conversation 必填.
	    {
	      "speaker": "DevOps助手",
	      "message": "xx 行代码会导致 panic",
	      "timestamp": "2024-01-01T10:00:00Z"
	    },
	    {
	      "speaker": "李韬",
	      "message": "不会导致；在 xxx 文件在这个函数的前面已经进行了nil判断",
	      "timestamp": "2024-01-01T10:01:00Z",
	    }
	  ],
	  "feedback": { // 如果 content_type = feedback，则 feedback 必填.
	    "type": "downvote",
	    "reason": "考虑不周全"
	  },
	  "document": { // 如果 content_type = document，则 document 必填.
	    "url": "https://xxx.com/xxx.pdf",
	    "title": "xxx.pdf",
	    "description": "xxx.pdf 的描述"
	  }
	}
*/
func (o *OpenAPI) CollectKnowledge(c *gin.Context) {
	var req types.CollectKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// default auto
	if req.ContentType == "" {
		req.ContentType = constants.ContentTypeAuto
	}

	if err := validateCollectKnowledgeRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	optimizedReq, err := o.requestOptimizer.Exec(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "内容优化失败: " + err.Error(),
		})
		return
	}

	log.Println("optimizedReq", optimizedReq)

	// TODO: 调用 indexer 收集知识

	c.JSON(http.StatusOK, types.CollectKnowledgeResponse{
		Message: "知识收集成功 - 自动识别类型",
	})
}

func validateCollectKnowledgeRequest(req *types.CollectKnowledgeRequest) error {
	if req.ContentType == "" {
		req.ContentType = constants.ContentTypeAuto
	}

	if req.Content == "" {
		return errors.New("content is required for auto type")
	}
	return nil
}
