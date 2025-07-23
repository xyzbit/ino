# Neo4j 图检索器

这个包实现了基于 Neo4j 图数据库的智能检索器，支持实体相似度搜索和图关系遍历。

## 功能特性

### 🔍 **向量相似度搜索**
- 基于 embedding 的语义相似度匹配
- 使用 Neo4j 向量索引进行高效检索
- 支持 COSINE 相似度计算

### 🕸️ **图关系遍历**
- 自动发现实体间的关联关系
- 支持多跳图遍历
- 智能过滤低质量关系（基于 hit 计数）

### 📄 **丰富的文档构建**
- 自动生成结构化文档内容
- 包含实体属性和关系信息
- 保留检索上下文和相似度分数

## 使用方法

### 1. 基本用法

```go
package main

import (
    "context"
    "log"
    
    neo4jIndexer "github.com/xyzbit/ino/pkg/components/indexer/neo4j"
    "github.com/cloudwego/eino/components/retriever"
)

func main() {
    ctx := context.Background()
    
    // 创建检索器
    neo4jRetriever, err := neo4jIndexer.NewNeo4jRetriever(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 执行检索
    documents, err := neo4jRetriever.Retrieve(ctx, "代码评审流程")
    if err != nil {
        log.Fatal(err)
    }
    
    // 处理结果
    for _, doc := range documents {
        log.Printf("实体: %s", doc.MetaData["entity_name"])
        log.Printf("内容: %s", doc.Content)
    }
}
```

### 2. 高级配置

```go
// 自定义配置
retriever, err := neo4jIndexer.NewRetriever(ctx, &neo4jIndexer.RetrieverConfig{
    Driver:              neo4jClient.Driver,
    Database:            "neo4j",
    EmbeddingModel:      embeddingModel,
    Dimension:           2560,
    SimilarityThreshold: 0.8,  // 提高相似度阈值
    MaxDepth:           3,     // 增加图遍历深度
    TopK:               15,    // 增加返回结果数量
})

// 使用选项进行检索
documents, err := retriever.Retrieve(ctx, query,
    retriever.WithTopK(10),
    retriever.WithScoreThreshold(0.9),
)
```

### 3. 检索选项

```go
// 支持的检索选项
type RetrieverImplOptions struct {
    TopK                int      // 返回结果数量
    SimilarityThreshold float64  // 相似度阈值
    MaxDepth           int      // 图遍历最大深度
    EntityTypes        []string // 限制实体类型
    IncludeRelations   bool     // 是否包含关系信息
}
```

## 检索流程

### 1. **查询处理**
```
用户查询 → Embedding生成 → 向量相似度搜索
```

### 2. **图遍历**
```
相似实体 → 关系发现 → 相关实体获取 → 结果聚合
```

### 3. **文档构建**
```
实体信息 + 属性信息 + 关系信息 → 结构化文档
```

## Cypher 查询示例

### 向量相似度搜索 + 图遍历
```cypher
CALL db.index.vector.queryNodes('entity_embedding_index', $top_k, $query_embedding)
YIELD node, score
WHERE score >= $similarity_threshold

// 获取相关实体和关系
OPTIONAL MATCH (node)-[r]-(related:Entity)
WHERE r.hit >= 1

WITH node, score, 
     collect(DISTINCT {
         entity: related.name, 
         entity_type: related.type,
         relation: type(r),
         relation_props: r.properties
     })[..5] as relations

RETURN 
    node.name as entity_name,
    node.type as entity_type,
    score,
    node.properties as properties,
    relations
ORDER BY score DESC
```

## 返回结果结构

### Document 结构
```go
type Document struct {
    ID       string                 // "neo4j_entity_{type}_{name}"
    Content  string                 // 结构化的实体描述
    MetaData map[string]interface{} // 丰富的元数据
}
```

### MetaData 字段
```go
{
    "source":             "neo4j_graph",
    "entity_name":        "实体名称",
    "entity_type":        "实体类型", 
    "score":              0.85,
    "query":              "原始查询",
    "retrieved_at":       1640995200,
    "entity_properties":  {...},      // 实体属性
    "relations":          [...]       // 关系信息
}
```

### Content 格式
```
实体: 代码评审 (流程)
属性:
  - 阶段: 开发完成后
  - 目的: 质量控制
相关关系:
  - 包含: 静态分析 (工具)
  - 需要: 同行评审 (活动)
  - 产生: 评审报告 (文档)

[检索查询: 代码评审流程]
```

## 配置参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `SimilarityThreshold` | float64 | 0.7 | 相似度阈值 (0-1) |
| `TopK` | int | 10 | 返回结果数量 |
| `MaxDepth` | int | 2 | 图遍历最大深度 |
| `Dimension` | int | 2048 | Embedding 维度 |

## 性能优化建议

### 1. **索引优化**
- 确保向量索引已正确创建
- 监控查询性能和内存使用

### 2. **查询优化**
- 合理设置 `TopK` 值，避免过大
- 调整 `SimilarityThreshold` 平衡召回率和精度
- 限制 `MaxDepth` 避免深度遍历性能问题

### 3. **数据质量**
- 定期清理低质量关系 (`hit` 计数低)
- 优化实体命名和类型分类
- 保持图数据的结构化和一致性

## 错误处理

```go
documents, err := retriever.Retrieve(ctx, query)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "connection"):
        // Neo4j 连接问题
        log.Printf("Neo4j connection error: %v", err)
    case strings.Contains(err.Error(), "embedding"):
        // Embedding 生成问题
        log.Printf("Embedding generation error: %v", err)
    default:
        // 其他错误
        log.Printf("Retrieval error: %v", err)
    }
}
```

## 与其他检索器集成

```go
// 多路检索融合
type MultiRetriever struct {
    vectorRetriever *milvus.Retriever
    graphRetriever  *neo4j.Retriever
    keywordRetriever *redis.Retriever
}

func (mr *MultiRetriever) Retrieve(ctx context.Context, query string) ([]*schema.Document, error) {
    var allDocs []*schema.Document
    
    // 并行检索
    vectorDocs, _ := mr.vectorRetriever.Retrieve(ctx, query)
    graphDocs, _ := mr.graphRetriever.Retrieve(ctx, query)
    keywordDocs, _ := mr.keywordRetriever.Retrieve(ctx, query)
    
    allDocs = append(allDocs, vectorDocs...)
    allDocs = append(allDocs, graphDocs...)
    allDocs = append(allDocs, keywordDocs...)
    
    // 去重和重排序
    return deduplicateAndRerank(allDocs, query), nil
}
```


## 完整案例

```go
package indexer

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/xyzbit/ino/config"
	neo4jClient "github.com/xyzbit/ino/pkg/infra/neo4j"
)

// NewNeo4jRetriever 创建Neo4j检索器的便捷函数
func NewNeo4jRetriever(ctx context.Context) (*Retriever, error) {
	embeddingConfig := config.AppConfig.Indexer.Embedding

	// 创建embedding模型
	embeddingModel, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		BaseURL: embeddingConfig.BaseURL,
		APIKey:  embeddingConfig.APIKey,
		Model:   embeddingConfig.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding model: %w", err)
	}

	// 创建Neo4j检索器
	retriever, err := NewRetriever(ctx, &RetrieverConfig{
		Driver:              neo4jClient.Driver,
		Database:            "neo4j",
		EmbeddingModel:      embeddingModel,
		Dimension:           2560, // 与索引器保持一致
		SimilarityThreshold: 0.7,
		MaxDepth:            2,
		TopK:                10,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j retriever: %w", err)
	}

	return retriever, nil
}

// ExampleUsage 展示如何使用Neo4j检索器
func ExampleUsage() error {
	ctx := context.Background()

	// 创建Neo4j检索器
	neo4jRetriever, err := NewNeo4jRetriever(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Neo4j retriever: %w", err)
	}

	// 执行检索
	query := "代码评审流程"
	documents, err := neo4jRetriever.Retrieve(ctx, query,
		retriever.WithTopK(5),             // 限制返回5个结果
		retriever.WithScoreThreshold(0.8), // 设置相似度阈值
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// 输出结果
	log.Printf("Found %d documents for query: %s", len(documents), query)
	for i, doc := range documents {
		log.Printf("Document %d:", i+1)
		log.Printf("  ID: %s", doc.ID)
		log.Printf("  Content: %s", doc.Content[:min(200, len(doc.Content))])
		log.Printf("  Metadata: %+v", doc.MetaData)
		log.Println("---")
	}

	return nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


```

## 注意事项

1. **数据一致性**：确保 Neo4j 中的数据与其他数据源同步
2. **内存管理**：大规模图遍历可能消耗大量内存
3. **查询复杂度**：避免过于复杂的 Cypher 查询影响性能
4. **并发安全**：检索器是线程安全的，可以并发使用 