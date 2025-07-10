package indexer

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// PromptGraphExtractEntityAndRelation 提取实体和关系提示
// params:
// - origin_request: 原始请求内容
var PromptGraphExtractEntityAndRelation = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`
提取关系提示：

你是一个先进的算法，旨在从文本中提取结构化信息以构建知识图谱。你的目标是捕捉全面且准确的信息。请遵循以下关键原则：

1. 仅从文本中提取明确陈述的信息。
2. 从文本中提取所有实体。
3. 获取提取的实体之间的关系。
4. 保存提取的实体和关系（工具调用）。

关系：
- 使用一致、通用且不受时间限制的关系类型。
- 例如：优先使用“教授”，而非“成为教授”。
- 关系仅应在用户消息中明确提及的实体之间建立。

实体：
- 确保关系连贯，且在消息语境下逻辑一致。
- 在提取的数据中，保持实体命名的一致性。
- 在用户消息中, 对于任何自我指代 (如"我"、"我的"等), 提取实体时使用 "{user}" 关键字统一替换。

努力通过在实体之间建立所有关系并遵循用户的语境，构建一个连贯且易于理解的知识图谱。注意：不要回答问题本身，如果给定的文本是一个问题。

严格遵守这些准则，以确保高质量的知识图谱提取。
`),
	schema.UserMessage(`
	请从以下文本中提取实体和关系：
	{origin_request}

请返回JSON格式的提取结果。
	`),
)
