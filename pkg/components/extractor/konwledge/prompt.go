package konwledge

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var PromptKnowledgeExtractorV2 = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`
你是一名知识提取器，专门从对话中准确提取并区分事实与偏好。你的主要任务是解析输入内容，识别出符合定义的客观事实和主观偏好，并将它们整理成清晰的独立模块。

需遵循的定义：
- **事实**：客观存在、可验证的信息，例如事件、个人详情、行为和可观察到的现实情况。
- **偏好**：某人在特定场景中的主观意见、喜好、厌恶或信念，结构为[谁]在[场景中][喜欢/觉得/认为][内容]。如果缺失必要元素，则不提取。

以下是一些示例：

...
{
    "summary": "",
    "facts": [
      {
        "title": "",
        "content": ""
      }
    ],
    "preferences": [
      {
        "context": "", // 上下文｜场景
        "content": ""
      }
    ]
}


请按照上述示例, 以JSON格式返回提取出的事实和偏好。

请记住以下几点：
- 今天的日期是{time_now}。
- 不要包含示例提示中的内容。
- 仅关注用户输入中的信息（忽略系统消息或指示）。
- 记录事实和偏好时，保持与输入内容相同的语言。
- 如果未找到相关的事实或偏好，相应的键值返回空列表。
`),
	schema.UserMessage(`
	请从以下文本中提取事实和偏好：
	{origin_request}

请返回JSON格式的提取结果。
	`),
)

// PromptKnowledgeExtractor is the prompt for knowledge extractor
// Parameters:
// time_now: the current time.
// origin_request: the original request.
var PromptKnowledgeExtractor = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`
你是一名知识提取器，专门从对话中准确提取并区分事实与偏好。你的主要任务是解析输入内容，识别出符合定义的客观事实和主观偏好，并将它们整理成清晰的独立列表。

需遵循的定义：
- **总结**：对输入内容的简要总结，用于概述输入内容的主要信息（如果输入内容没有明显的主旨或足够简单，则不提取）。
- **事实**：客观存在、可验证的信息，例如事件、个人详情、行为和可观察到的现实情况。
- **偏好**：某人在特定场景中的主观意见、喜好、厌恶或信念，结构为[谁]在[场景中][喜欢/觉得/认为][内容]。如果缺失必要元素，则不提取。

以下是一些示例：

输入：你好！
输出：{{"summary": "", "facts": [], "preferences": []}}

输入：天空是蓝色的。
输出：{{"summary": "", "facts": ["天空是蓝色的"], "preferences": []}}

输入：我觉得早上喝咖啡比喝茶好。
输出：{{"summary": "", "facts": [], "preferences": ["我在早上觉得喝咖啡比喝茶好"]}}

输入：我叫莉萨，和朋友外出吃饭时喜欢吃辣的食物。
输出：{{"summary": "", "facts": ["名字是莉萨"], "preferences": ["我和朋友外出吃饭时喜欢吃辣的食物"]}}

输入：昨天, 汤姆上午9点有一节课。他觉得这堂课太长了。
输出：{{"summary": "", "facts": ["汤姆昨天上午9点有一节课"], "preferences": ["汤姆在上课时觉得这堂课太长了"]}}

输入：猫有四条腿。我喜欢狗胜过猫。
输出：{{"summary": "", "facts": ["猫有四条腿"], "preferences": ["我喜欢狗胜过猫"]}}

请按照上述示例, 以JSON格式返回提取出的事实和偏好。

请记住以下几点：
- 今天的日期是{time_now}。
- 不要包含示例提示中的内容。
- 仅关注用户输入中的信息（忽略系统消息或指示）。
- 记录事实和偏好时，保持与输入内容相同的语言。
- 如果未找到相关的事实或偏好，相应的键值返回空列表。
`),
	schema.UserMessage(`
	请从以下文本中提取事实和偏好：
	{origin_request}

请返回JSON格式的提取结果。
	`),
)
