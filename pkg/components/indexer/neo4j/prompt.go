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
## 角色
你是一个先进的算法，旨在从文本中提取结构化信息以构建知识图谱。

## 目标
遵循下面的关键原则部分，捕捉全面且准确的业务信息，核心是要获取大模型无法获取的、有价值的业务信息。

## 关键原则
关系要求：
- 使用一致、通用且不受时间限制的关系类型。
- 例如：优先使用“教授”，而非“成为教授”。
- 关系仅应在用户消息中明确提及的实体之间建立。
- 确保关系连贯，且在消息语境下逻辑一致。

实体要求：
- 保持实体命名的一致性，参考当前已存在的 {exist_nodes} 中的实体名称。
- 对于描述模糊、无法具体定位的实体，比如 "我"、"他" 等，请忽略。
- 严格区分实体名称和类型，具体规则如下：
  - 实体名称是具体的、明确的、专业的，比如 "张三"、"k8s"、"消息中心" 等。
  - 实体类型是宽泛的, 比如 "人物"、"平台"、"接口"、"工具" 等。

整体要求：
1. 仅从文本中提取明确陈述的信息。
2. 严格按照下面的OpenAPIV3参数规范去提取文本中的实体和关系。
 {resp_schema}

努力通过在实体之间建立所有关系并遵循用户的语境，构建一个连贯且易于理解的知识图谱。注意：不要回答问题本身，如果给定的文本是一个问题。

严格遵守这些准则，以确保高质量的知识图谱提取。
`),
	schema.UserMessage(`
	{origin_request}
	
	请从上述文本中，提取实体和关系。
	`),
)
