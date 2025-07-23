package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/pkg/ctxwarp"
)

func middlewareSetHeaderContext(c *gin.Context) {
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
		RequestID:     header.RequestID,
		CollectionKey: header.CollectionKey,
		UserKey:       header.UserKey,
	})
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

// responseWriter 是一个包装器，用于捕获响应状态码和响应体
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// 将响应体写入缓冲区
	rw.body.Write(b)
	// 同时写入原始的ResponseWriter
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// 统一的日志中间件, 请求今日时打印请求和参数；请求返回时，检测response的错误码，如果有错则打印错误
func middlewareCheckError(c *gin.Context) {
	// 获取请求开始时间
	startTime := time.Now()

	// 获取请求ID
	var requestID string
	if headerCtx := ctxwarp.GetHeaderContext(c.Request.Context()); headerCtx != nil {
		requestID = headerCtx.RequestID
	}
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// 读取请求体（如果有的话）
	var requestBody string
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			requestBody = string(bodyBytes)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// 记录请求信息
	log.Printf("[REQUEST] [%s] %s %s - User: %s, Collection: %s, Body: %s",
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		c.GetHeader("user_key"),
		c.GetHeader("collection_key"),
		requestBody)

	// 创建响应包装器
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		statusCode:     200, // 默认状态码
		body:           bytes.NewBuffer([]byte{}),
	}
	c.Writer = rw

	// 继续处理请求
	c.Next()

	// 计算请求处理时间
	duration := time.Since(startTime)

	// 检查响应状态码
	if rw.statusCode >= 400 {
		// 记录错误信息
		log.Printf("[ERROR] [%s] %s %s - Status: %d, Duration: %v, Response: %s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			rw.statusCode,
			duration,
			rw.body.String())
	} else {
		// 记录成功请求信息
		log.Printf("[SUCCESS] [%s] %s %s - Status: %d, Duration: %v",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			rw.statusCode,
			duration)
	}
}
