package models

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/xyzbit/ino/pkg/constants"
)

// OptimizedRequest 内容类型分析结果
type OptimizedRequest struct {
	ContentType  string        `json:"content_type"`
	Conversation *Conversation `json:"conversation,omitempty"`
	Feedback     *Feedback     `json:"feedback,omitempty"`
	Document     *Document     `json:"document,omitempty"`
}

func (o *OptimizedRequest) GetDescription() string {
	desc := strings.Builder{}
	desc.WriteString(`{`)
	switch o.ContentType {
	case constants.ContentTypeConversation:
		desc.WriteString(fmt.Sprintf(`"conversation": %s,`, o.Conversation.GetDescription()))
	case constants.ContentTypeFeedback:
		desc.WriteString(fmt.Sprintf(`"feedback": %s,`, o.Feedback.GetDescription()))
	case constants.ContentTypeDocument:
		desc.WriteString(fmt.Sprintf(`"document": %s,`, o.Document.GetDescription()))
	}
	desc.WriteString(`}`)
	return desc.String()
}

type ContentTypeClassificationResult struct {
	ContentType string  `json:"content_type"`
	Confidence  float64 `json:"confidence"`
	Reason      string  `json:"reason"`
}

func (o *ContentTypeClassificationResult) GetDescription() string {
	return `{
		"content_type": "conversation|feedback|document",
		"confidence": 0.0-1.0,
		"reason": "识别理由"
	  }`
}

// PromptContentTypeClassification 内容类型识别模版
var PromptContentTypeClassification = prompt.FromMessages(schema.FString,
	// 系统消息模板
	schema.SystemMessage(`你是一个专业的内容类型识别和信息提取专家。你的任务是分析用户输入的内容，判断其类型，并提取相应的结构化信息。

支持的内容类型：
1. conversation(对话):包含多轮对话、讨论记录、会议纪要等
2. feedback(反馈):包含用户评价、意见反馈、评分评论等  
3. document(文档):包含文档信息、链接、文件描述等

分析要求：
- 准确识别内容类型（置信度 > 0.7）
- 提取关键结构化信息
- 提供识别理由
- 如果内容模糊，选择最可能的类型

输出格式要求：
{output_desc}
严格按照 JSON 格式返回，包含以下字段：
`),

	// 用户消息模板
	schema.UserMessage(`请分析以下内容并提取信息：{origin_request}

请返回JSON格式的分析结果。`),
)

// ConversationExtractionPrompt 对话提取专用模版
var PromptConversationExtraction = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是对话内容提取专家。请从给定文本中提取对话信息，识别说话人、消息内容和时间戳。

提取规则：
1. 识别所有参与对话的人员
2. 按时间顺序提取每条消息
3. 尝试推断消息的角色类型(user/assistant/system)
4. 提取对话主题和上下文信息

避免事项：
1. 无法提取或分析的部分，不要瞎编，直接返回空

输出格式：
{output_desc}
严格按照 JSON 格式返回.
`),

	schema.UserMessage(`{origin_request}

请从上述内容中提取对话信息：`),
)

// FeedbackExtractionPrompt 反馈提取专用模版
var PromptFeedbackExtraction = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是用户反馈分析专家。请从给定文本中提取反馈信息，包括情感倾向、评分和具体原因。

分析规则：
1. 判断反馈类型: positive(正面)、negative(负面)、neutral(中性)
2. 如果有明确评分，提取数字评分(1-5分)
3. 总结反馈的具体原因
4. 识别相关的查询或上下文

避免事项：
1. 无法提取或分析的部分，不要瞎编，直接返回空

输出格式：
{output_desc}
严格按照 JSON 格式返回.`),

	schema.UserMessage(`{origin_request}

请从上述内容中提取反馈信息：`),
)

// DocumentExtractionPrompt 文档提取专用模版
var PromptDocumentExtraction = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是文档信息提取专家。请从给定文本中提取文档相关信息，包括标题、描述、链接等。

提取规则：
1. 识别文档标题或名称
2. 提取文档描述或摘要
3. 查找文档链接或路径(可能原始内容也可能是链接)
4. 判断文档类型(无法判断时则为 markdown)
5. 根据文档类型进行格式优化
5. 提取相关标签

避免事项：
1. 无法提取或分析的部分，不要瞎编，直接返回空

输出格式：
{output_desc}
严格按照 JSON 格式返回.`),

	schema.UserMessage(`{origin_request}

请从上述内容中提取文档信息：`),
)
