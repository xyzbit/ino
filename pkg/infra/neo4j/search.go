package neo4j

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pkg/errors"
)

type SearchOptions struct {
	TopK                int
	Limit               int
	SimilarityThreshold float64
	MaxDepth            int
	Filter              map[string]string
}

// graphSearch 执行图检索
func GraphSearch(ctx context.Context, session neo4j.SessionWithContext, embedding []float64, opts *SearchOptions) ([]*neo4j.Record, error) {
	// 构建Cypher查询
	cypher := buildSearchCypher(opts)

	// 准备参数
	embeddingList := make([]interface{}, len(embedding))
	for i, v := range embedding {
		embeddingList[i] = v
	}

	params := map[string]interface{}{
		"query_embedding":      embeddingList,
		"similarity_threshold": opts.SimilarityThreshold,
		"top_k":                opts.TopK,
		"limit":                opts.Limit,
	}

	log.Printf("Neo4j Search - TopK: %d, Threshold: %f", opts.TopK, opts.SimilarityThreshold)

	// 执行查询
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var documents []*neo4j.Record
		for result.Next(ctx) {
			record := result.Record()
			documents = append(documents, record)
		}

		return documents, result.Err()
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result.([]*neo4j.Record), nil
}

// searchRecentNode 查询最近加入的节点
func SearchRecentNode(ctx context.Context, session neo4j.SessionWithContext, collectionKey string) ([]*neo4j.Record, error) {
	attr := ""
	if collectionKey != "" {
		attr = "{collection_key: $collection_key}"
	}
	cypher := fmt.Sprintf(`
		MATCH (n:Entity%s)
		RETURN n.name as entity_name, n.type as entity_type
		ORDER BY n.created DESC
		LIMIT $limit
	`, attr)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]interface{}{
			"limit":          20,
			"collection_key": collectionKey,
		})
		if err != nil {
			return nil, err
		}

		var documents []*neo4j.Record
		for result.Next(ctx) {
			record := result.Record()
			documents = append(documents, record)
		}

		return documents, result.Err()
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result.([]*neo4j.Record), nil
}

// buildSearchCypher 构建搜索的Cypher查询
func buildSearchCypher(opts *SearchOptions) string {
	return fmt.Sprintf(`
			CALL db.index.vector.queryNodes('entity_embedding_index', $top_k, $query_embedding)
			YIELD node, score
			WHERE score >= $similarity_threshold %s
			
			// 获取相关实体和关系
			OPTIONAL MATCH (node)-[rels*1..%d]-(related:Entity)
			WHERE all(r IN rels WHERE r.hit >= 1)  // 只获取有足够支撑的关系
			
			WITH node, rels, related, score
			
			RETURN 
			    node.name as entity_name,
			    node.type as entity_type,
				related.name as related_name,
				related.type as related_type,
				[rel in rels | type(rel)] as relations,
                size(rels) as layer,
				score
			ORDER BY score DESC, layer ASC

			LIMIT $limit
		`, opts.GetFilter(), opts.MaxDepth)
}

func (r *SearchOptions) GetFilter() string {
	if len(r.Filter) == 0 {
		return ""
	}
	filter := ""
	for k, v := range r.Filter {
		filter += fmt.Sprintf(" AND node.%s = \"%s\"", k, v)
	}
	return filter
}
