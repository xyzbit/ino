# 内容类型识别和提取功能使用示例

## 概述

本系统提供了智能的内容类型识别和信息提取功能，能够自动分析用户输入的内容，判断其类型（对话、反馈、文档），并提取相应的结构化信息。

## 支持的内容类型

1. **conversation**（对话）：多轮对话、讨论记录、会议纪要等
2. **feedback**（反馈）：用户评价、意见反馈、评分评论等  
3. **document**（文档）：文档信息、链接、文件描述等

## API 使用示例

### 1. 自动识别内容类型（推荐）

**请求示例 - 对话内容：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "code-review",
  "tags": {
    "project": "ino-system",
    "type": "discussion"
  },
  "content_type": "auto",
  "content": "DevOps助手: xx 行代码会导致 panic\n李韬: 不会导致；在 xxx 文件在这个函数的前面已经进行了nil判断\nDevOps助手: 你说得对，我检查了代码，确实有nil判断"
}
```

**响应示例：**

```json
{
  "message": "知识收集成功 - 自动识别类型",
  "processed_type": "conversation",
  "analysis_result": {
    "content_type": "conversation",
    "confidence": 0.95,
    "conversation": {
      "messages": [
        {
          "speaker": "DevOps助手",
          "message": "xx 行代码会导致 panic",
          "role": "assistant"
        },
        {
          "speaker": "李韬",
          "message": "不会导致；在 xxx 文件在这个函数的前面已经进行了nil判断",
          "role": "user"
        },
        {
          "speaker": "DevOps助手",
          "message": "你说得对，我检查了代码，确实有nil判断",
          "role": "assistant"
        }
      ],
      "context": {
        "topic": "代码评审讨论",
        "participants": ["DevOps助手", "李韬"],
        "domain": "code-review"
      }
    },
    "reason": "检测到对话特征"
  }
}
```

**请求示例 - 反馈内容：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "general",
  "content_type": "auto",
  "content": "这个功能很好用，但是响应速度有点慢，建议优化一下性能。给个4分吧。"
}
```

**响应示例：**

```json
{
  "message": "知识收集成功 - 自动识别类型",
  "processed_type": "feedback",
  "analysis_result": {
    "content_type": "feedback",
    "confidence": 0.90,
    "feedback": {
      "type": "positive",
      "rating": 4,
      "reason": "功能好用但响应速度慢，建议优化性能"
    },
    "reason": "检测到反馈特征"
  }
}
```

**请求示例 - 文档内容：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "documentation",
  "content_type": "auto",
  "content": "请查看最新的API文档：https://docs.example.com/api/v1.pdf，这份文档详细介绍了所有接口的使用方法和参数说明"
}
```

**响应示例：**

```json
{
  "message": "知识收集成功 - 自动识别类型",
  "processed_type": "document",
  "analysis_result": {
    "content_type": "document",
    "confidence": 0.85,
    "document": {
      "title": "API文档",
      "description": "详细介绍了所有接口的使用方法和参数说明",
      "url": "https://docs.example.com/api/v1.pdf",
      "content_type": "pdf",
      "summary": "API接口使用方法和参数说明文档"
    },
    "reason": "检测到文档特征"
  }
}
```

### 2. 手动指定内容类型

**对话类型：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "support",
  "content_type": "conversation",
  "conversation": [
    {
      "speaker": "客户",
      "message": "我的订单一直显示处理中，什么时候能发货？",
      "timestamp": "2024-01-01T10:00:00Z"
    },
    {
      "speaker": "客服",
      "message": "请提供您的订单号，我帮您查询一下",
      "timestamp": "2024-01-01T10:01:00Z"
    }
  ]
}
```

**反馈类型：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "product-feedback",
  "content_type": "feedback",
  "feedback": {
    "type": "negative",
    "reason": "页面加载太慢，用户体验很差"
  }
}
```

**文档类型：**

```json
POST /api/v1/knowledge/collect
{
  "domain": "documentation",
  "content_type": "document",
  "document": {
    "url": "https://example.com/manual.pdf",
    "title": "用户手册",
    "description": "产品使用说明书",
    "tags": ["manual", "guide", "help"]
  }
}
```

## Prompt 模版系统

### 主要识别模版 (ContentTypeAnalysisPrompt)

```go
// 系统提示词
你是一个专业的内容类型识别和信息提取专家。你的任务是分析用户输入的内容，判断其类型，并提取相应的结构化信息。

支持的内容类型：
1. conversation（对话）：包含多轮对话、讨论记录、会议纪要等
2. feedback（反馈）：包含用户评价、意见反馈、评分评论等  
3. document（文档）：包含文档信息、链接、文件描述等

分析要求：
- 准确识别内容类型（置信度 > 0.8）
- 提取关键结构化信息
- 提供识别理由
- 如果内容模糊，选择最可能的类型

// 用户输入模版
请分析以下内容并提取信息：

内容：{content}
领域：{domain}
标签：{tags}

请返回JSON格式的分析结果。
```

### 专项提取模版

1. **对话提取模版 (ConversationExtractionPrompt)**
   - 识别说话人、消息内容和时间戳
   - 推断消息角色类型（user/assistant/system）
   - 提取对话主题和上下文信息

2. **反馈提取模版 (FeedbackExtractionPrompt)**
   - 判断反馈类型（positive/negative/neutral）
   - 提取评分信息（1-5分）
   - 总结反馈具体原因

3. **文档提取模版 (DocumentExtractionPrompt)**
   - 识别文档标题或名称
   - 提取文档描述或摘要
   - 查找文档链接或路径
   - 判断文档类型

## 技术架构

### 核心组件

1. **ContentAnalyzer**: 内容分析器服务
   - 负责调用 AI 模型进行内容分析
   - 支持多层级分析（主分析 + 专项分析）
   - 置信度评估和二次确认

2. **Prompt Templates**: 提示词模版系统
   - 主要识别模版：ContentTypeAnalysisPrompt
   - 专项提取模版：ConversationExtractionPrompt, FeedbackExtractionPrompt, DocumentExtractionPrompt

3. **API Integration**: API 集成层
   - 统一的知识收集接口
   - 自动类型识别和处理
   - 结构化数据输出

### 工作流程

1. **用户输入** → 内容验证
2. **类型判断** → 自动识别 OR 手动指定
3. **内容分析** → 调用对应的 Prompt 模版
4. **结果处理** → 结构化数据提取
5. **响应返回** → 统一格式输出

## 配置和扩展

### 添加新的内容类型

1. 在 `constants.go` 中添加新的类型常量
2. 在 `content_type_analysis.go` 中添加对应的数据结构
3. 创建专用的 Prompt 模版
4. 在 `ContentAnalyzer` 中添加处理逻辑
5. 更新 API 接口支持新类型

### 优化识别准确率

1. **丰富训练数据**：收集更多样化的内容样本
2. **调整提示词**：根据实际使用情况优化 Prompt
3. **置信度阈值**：调整二次分析的触发条件
4. **上下文增强**：利用领域和标签信息提高准确性

## 注意事项

1. **模拟实现**：当前使用了模拟的 AI 调用，实际部署时需要集成真实的 Eino AI 服务
2. **错误处理**：建议添加更详细的错误处理和日志记录
3. **性能优化**：对于大量请求，考虑添加缓存和批处理机制
4. **安全性**：对输入内容进行安全验证，防止恶意输入

## 测试用例

### 对话测试样例
- 技术讨论：代码评审、问题解答
- 客服对话：问题咨询、售后服务
- 会议记录：讨论要点、决策过程

### 反馈测试样例
- 产品评价：功能体验、使用建议
- 服务反馈：满意度评价、改进意见
- Bug 报告：问题描述、严重程度

### 文档测试样例
- 文档分享：链接、标题、描述
- 文件上传：文档类型、元数据
- 知识条目：文档摘要、关键信息 