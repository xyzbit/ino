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
你是一个先进的算法，旨在从文本中提取结构化信息以构建知识图谱。你的目标是捕捉全面且准确的信息。请遵循以下关键原则：
 
关系要求：
- 使用一致、通用且不受时间限制的关系类型。
- 例如：优先使用“教授”，而非“成为教授”。
- 关系仅应在用户消息中明确提及的实体之间建立。

实体要求：
- 确保关系连贯，且在消息语境下逻辑一致。
- 在提取的数据中，保持实体命名的一致性。
- 在用户消息中, 对于任何自我指代 (如"我"、"我的"等), 提取实体时使用 "{user_key}" 关键字统一替换(果"{user_key}"=="", 则不要提取"我"的偏好)

整体要求：
1. 仅从文本中提取明确陈述的信息。
2. 严格按照下面的OpenAPIV3参数规范去提取请求内容。
 {resp_schema}

努力通过在实体之间建立所有关系并遵循用户的语境，构建一个连贯且易于理解的知识图谱。注意：不要回答问题本身，如果给定的文本是一个问题。

严格遵守这些准则，以确保高质量的知识图谱提取。
`),
	schema.UserMessage(`
	{origin_request}
	
	请从上述文本中，提取实体和关系。
	`),
)
