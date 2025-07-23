## 简介
大模型应用当前有着 没有长期记忆、缺乏业务知识、缺乏实时性等问题，这些问题的本质是大模型自主获取｜记录部分知识，比如：
- 企业私有知识
- 记忆数据，如用户行为、反馈、偏好...
这些可以统称为知识，当前有着 RAG、Memeroy System等等解决方案，去补充各种类型的知识，但是接入复杂，有很多重复工作。

为了解决这些问题 ino 诞生了，所以有了 KAG（Knowledge-Augmented Generation）或者叫 统一检索框架（Unified Retrieval Framework），这个系统的核心思想是：将所有可能对LLM有帮助的信息源（外部文档、对话历史、用户画像等）视为可检索的“知识”，并建立一个统一的框架来智能地检索、筛选、整合这些知识，最后以最优化的方式注入到Prompt中。

## 如何开发
> ！注意需要先安装 Docker

### 启动依赖组件
make services-up

### 运行 ino
make dev

## 如何使用

### 写入

通过调用 `POST /api/v1/openapi/collect` 接口来写入知识。

**请求示例**
```bash
curl --location 'http://localhost:8080/api/v1/openapi/collect' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "content": "The quick brown fox jumps over the lazy dog."
}'
```
或者通过链接写入
```bash
curl --location 'http://localhost:8080/api/v1/openapi/collect' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "content_link": "https://en.wikipedia.org/wiki/Fox"
}'
```

**参数说明**
- `user-key`: (Header) 可选，用于标识用户。
- `collection-key` (Header) 可选，用于标识集合。
- `content`: (Body) 知识内容，`content` 和 `content_link` 至少需要一个。
- `content_link`: (Body) 知识内容链接。


### 读取

通过调用 `POST /api/v1/openapi/search` 接口来读取知识。

**请求示例**
```bash
curl --location 'http://localhost:8080/api/v1/openapi/search' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "query": "What does the fox say?"
}'
```

**参数说明**
- `user-key`: (Header) 可选，用于筛选用户。
- `collection-key` (Header) 可选，用于筛选集合。
- `query`: (Body) 必需，查询的问题。
- `query_strategy`: (Body) 可选，查询策略，`quick`（默认）或 `agent`。

> 更多文档请在 ./docs 中查看