package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/pkg/errors"

	"github.com/xyzbit/ino/config"
	"github.com/xyzbit/ino/internal/application/openapi/types"
	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/pkg/constants"
)

// RequestOptimizer 请求优化器
type RequestOptimizer struct {
	optimizerModel *openai.ChatModel
}

// NewRequestOptimizer 创建请求优化器实例
func NewRequestOptimizer() (*RequestOptimizer, error) {
	optimizerConfig := config.AppConfig.Eino.Optimizer
	optimizerModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: optimizerConfig.BaseURL,
		Model:   optimizerConfig.Model,
		APIKey:  optimizerConfig.APIKey,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &RequestOptimizer{
		optimizerModel: optimizerModel,
	}, nil
}

// AnalyzeContent 分析内容类型并提取信息
func (ca *RequestOptimizer) Exec(ctx context.Context, req *types.CollectKnowledgeRequest) (*types.CollectKnowledgeRequest, error) {
	// classify auto content type.
	contentType, err := ca.reqContentClassification(ctx, req.ContentType, req)
	if err != nil {
		return req, err
	}

	// extract content.
	content, err := ca.reqExtraction(ctx, contentType, req)
	if err != nil {
		return req, err
	}

	return &types.CollectKnowledgeRequest{
		ContentType:  contentType,
		CollectionID: req.CollectionID,
		Tags:         req.Tags,
		Content:      content,
	}, nil
}

// reqContentClassification 请求内容分类.
func (ca *RequestOptimizer) reqContentClassification(ctx context.Context, contentType string, req *types.CollectKnowledgeRequest) (string, error) {
	if contentType != "" {
		return contentType, nil
	}
	result := models.ContentTypeClassificationResult{}

	params := map[string]interface{}{
		"origin_request": req.GetDescription(),
		"output_desc":    result.GetDescription(),
	}

	messages, err := models.PromptContentTypeClassification.Format(ctx, params)
	if err != nil {
		return "", errors.WithStack(err)
	}

	outMessage, err := ca.optimizerModel.Generate(ctx, messages)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if err := json.Unmarshal([]byte(outMessage.Content), &result); err != nil {
		return "", errors.WithStack(err)
	}

	if result.Confidence < 0.7 {
		return "", errors.Errorf("confidence is too low: %v", result)
	}

	return result.ContentType, nil
}

// reqExtraction 提取请求参数.
func (ca *RequestOptimizer) reqExtraction(ctx context.Context, contentType string, req *types.CollectKnowledgeRequest) (string, error) {
	result := models.OptimizedRequest{ContentType: contentType}
	params := map[string]interface{}{
		"origin_request": req.GetDescription(),
		"output_desc":    result.GetDescription(),
	}

	var (
		messages []*schema.Message
		err      error
	)

	switch contentType {
	case constants.ContentTypeConversation:
		messages, err = models.PromptConversationExtraction.Format(ctx, params)
	case constants.ContentTypeFeedback:
		messages, err = models.PromptFeedbackExtraction.Format(ctx, params)
	case constants.ContentTypeDocument:
		messages, err = models.PromptDocumentExtraction.Format(ctx, params)
	}
	if err != nil {
		return "", errors.WithStack(err)
	}

	outMessage, err := ca.optimizerModel.Generate(ctx, messages)
	if err != nil {
		return "", errors.WithStack(err)
	}
	// validate result.
	if err := json.Unmarshal([]byte(outMessage.Content), &result); err != nil {
		return "", errors.WithStack(err)
	}

	return outMessage.Content, nil
}

// optimizeConversation 优化对话请求
// func (ca *RequestOptimizer) optimizeConversation(ctx context.Context, req *models.ContentTypeAnalysisRequest) (*models.OptimizedRequest, error) {
// 	params := map[string]interface{}{
// 		"content": req.Content,
// 		"domain":  req.Domain,
// 	}

// 	result, err := ca.executePrompt(ctx, "conversation_extraction", params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to analyze conversation: %w", err)
// 	}

// 	var conversationExtract models.ConversationExtract
// 	if err := json.Unmarshal([]byte(result), &conversationExtract); err != nil {
// 		return nil, fmt.Errorf("failed to parse conversation result: %w", err)
// 	}

// 	return &models.OptimizedRequest{
// 		ContentType:  constants.ContentTypeConversation,
// 		Confidence:   0.9, // 专项分析后提高置信度
// 		Conversation: &conversationExtract,
// 		Reason:       "通过专项对话分析确认为对话类型",
// 	}, nil
// }

// optimizeFeedback 优化反馈请求
// func (ca *RequestOptimizer) optimizeFeedback(ctx context.Context, req *models.ContentTypeAnalysisRequest) (*models.OptimizedRequest, error) {
// 	params := map[string]interface{}{
// 		"content": req.Content,
// 	}

// 	result, err := ca.executePrompt(ctx, "feedback_extraction", params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to analyze feedback: %w", err)
// 	}

// 	var feedbackExtract models.FeedbackExtract
// 	if err := json.Unmarshal([]byte(result), &feedbackExtract); err != nil {
// 		return nil, fmt.Errorf("failed to parse feedback result: %w", err)
// 	}

// 	return &models.OptimizedRequest{
// 		ContentType: constants.ContentTypeFeedback,
// 		Confidence:  0.9,
// 		Feedback:    &feedbackExtract,
// 		Reason:      "通过专项反馈分析确认为反馈类型",
// 	}, nil
// }

// optimizeDocument 优化文档请求
// func (ca *RequestOptimizer) optimizeDocument(ctx context.Context, req *models.ContentTypeAnalysisRequest) (*models.OptimizedRequest, error) {
// 	params := map[string]interface{}{
// 		"content": req.Content,
// 	}

// 	result, err := ca.executePrompt(ctx, "document_extraction", params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to analyze document: %w", err)
// 	}

// 	var documentExtract models.DocumentExtract
// 	if err := json.Unmarshal([]byte(result), &documentExtract); err != nil {
// 		return nil, fmt.Errorf("failed to parse document result: %w", err)
// 	}

// 	return &models.OptimizedRequest{
// 		ContentType: constants.ContentTypeDocument,
// 		Confidence:  0.9,
// 		Document:    &documentExtract,
// 		Reason:      "通过专项文档分析确认为文档类型",
// 	}, nil
// }

// formatTags 格式化标签为字符串
func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "无标签"
	}

	var tagStrings []string
	for k, v := range tags {
		tagStrings = append(tagStrings, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(tagStrings, ", ")
}

func (ca *RequestOptimizer) getMockConversationResult(content string) string {
	return `{
		"content_type": "conversation",
		"confidence": 0.95,
		"conversation": {
			"messages": [
				{
					"speaker": "用户",
					"message": "` + content + `",
					"role": "user"
				}
			],
			"context": {
				"topic": "对话内容",
				"participants": ["用户"],
				"domain": "general"
			}
		},
		"reason": "检测到对话特征"
	}`
}

func (ca *RequestOptimizer) getMockFeedbackResult(content string) string {
	feedbackType := "neutral"
	if strings.Contains(strings.ToLower(content), "好") || strings.Contains(strings.ToLower(content), "赞") {
		feedbackType = "positive"
	} else if strings.Contains(strings.ToLower(content), "差") || strings.Contains(strings.ToLower(content), "不") {
		feedbackType = "negative"
	}

	return `{
		"content_type": "feedback",
		"confidence": 0.90,
		"feedback": {
			"type": "` + feedbackType + `",
			"rating": 3,
			"reason": "` + content + `"
		},
		"reason": "检测到反馈特征"
	}`
}

func (ca *RequestOptimizer) getMockDocumentResult(content string) string {
	return `{
		"content_type": "document",
		"confidence": 0.85,
		"document": {
			"title": "文档内容",
			"description": "` + content + `",
			"content_type": "text"
		},
		"reason": "检测到文档特征"
	}`
}

func (ca *RequestOptimizer) getMockGenericResult(content string) string {
	return `{
		"content_type": "conversation",
		"confidence": 0.60,
		"conversation": {
			"messages": [
				{
					"speaker": "用户",
					"message": "` + content + `",
					"role": "user"
				}
			],
			"context": {
				"topic": "一般内容",
				"participants": ["用户"],
				"domain": "general"
			}
		},
		"reason": "默认归类为对话类型"
	}`
}
