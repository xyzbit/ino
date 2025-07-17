/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package indexer

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/pkg/ctxwarp"
)

const typ = "neo4j"

type IndexerConfig struct {
	// Driver is the Neo4j driver to be called
	// Required
	Driver neo4j.DriverWithContext

	// Database is the database name in Neo4j
	// Optional, default is "neo4j"
	// Note: Community Edition does not support creating multiple databases.
	Database string

	// Extractor is the model used to extract entities and relations from documents
	// Required
	Extractor model.ToolCallingChatModel

	// EmbeddingModel is the model used to generate embeddings for entities
	// Required
	EmbeddingModel embedding.Embedder

	// Dimension is the dimension of the embedding
	// Optional, default is 2048
	Dimension int

	// DocumentConverter is the model used to convert documents to entities and relations
	// Required
	DocumentConverter func(ctx context.Context, docs []*schema.Document) (StorePairs, error)

	// MaxRetries is the maximum number of retries for failed operations
	// Optional, default is 3
	MaxRetries int

	// RetryDelay is the delay between retries
	// Optional, default is 1 second
	RetryDelay time.Duration

	// BatchSize is the number of documents to process in one batch
	// Optional, default is 100
	BatchSize int

	respSchema string // 响应格式
}

type Indexer struct {
	config IndexerConfig
}

type ImplOptions struct {
	// Database specifies the Neo4j database to use
	Database string

	// EntityTypes specifies the entity types to extract
	EntityTypes []string

	// RelationTypes specifies the relation types to extract
	RelationTypes []string

	// SimilarityThreshold specifies the threshold for determining similar nodes, default is 0.9
	SimilarityThreshold float64
}

// NewIndexer creates a new Neo4j indexer
func NewIndexer(ctx context.Context, conf *IndexerConfig) (*Indexer, error) {
	// conf check
	if err := conf.check(); err != nil {
		return nil, err
	}

	// Test Neo4j connection
	if err := conf.Driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("[NewIndexer] Neo4j connection failed: %w", err)
	}

	// Create database if not exists
	if err := conf.ensureDatabaseExists(ctx); err != nil {
		return nil, fmt.Errorf("[NewIndexer] failed to ensure database exists: %w", err)
	}

	// Create indexes for better performance
	if err := conf.createIndexes(ctx, conf.Dimension); err != nil {
		return nil, fmt.Errorf("[NewIndexer] failed to create indexes: %w", err)
	}

	toolInfo, err := utils.GoStruct2ToolInfo[ExtractEntityAndRelation](
		toolNameStoreEntityAndRelation,
		"保存提取的实体和关系, Store entities and relations based on the provided text.",
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 部分模型不支持工具调用或工具调用能力差，使用更通用的模式.
	respSchema, err := toolInfo.ToOpenAPIV3()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	respSchemaBytes, err := respSchema.MarshalJSON()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conf.respSchema = string(respSchemaBytes)

	return &Indexer{
		config: *conf,
	}, nil
}

// Store stores the documents into the Neo4j graph database
func (i *Indexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	// get impl specific options
	io := indexer.GetImplSpecificOptions(&ImplOptions{
		SimilarityThreshold: 0.9,
	}, opts...)

	ctx = callbacks.EnsureRunInfo(ctx, i.GetType(), components.ComponentOfIndexer)
	// callback info on start
	ctx = callbacks.OnStart(ctx, &indexer.CallbackInput{
		Docs: docs,
	})
	defer func() {
		if err != nil {
			callbacks.OnError(ctx, err)
		}
	}()

	if len(docs) == 0 {
		return []string{}, nil
	}

	// Convert documents to entities and relations
	addPairs, err := i.config.DocumentConverter(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("[Indexer.Store] failed to convert documents: %w", err)
	}

	// TODO: 解决冲突，获取相似节点，交给大模型判断是否需要删除。
	// search_output = self._search_graph_db(node_list=list(entity_type_map.keys()), filters=filters)
	// to_be_deleted = self._get_delete_entities_from_search_output(search_output, data, filters)

	// Process in batches
	var allIDs []string
	batchSize := i.config.BatchSize

	for idx := 0; idx < len(addPairs); idx += batchSize {
		end := idx + batchSize
		if end > len(addPairs) {
			end = len(addPairs)
		}

		batch := addPairs[idx:end]
		batchIDs, err := i.storeBatch(ctx, batch, io)
		if err != nil {
			return nil, fmt.Errorf("[Indexer.Store] failed to store entities batch: %w", err)
		}
		allIDs = append(allIDs, batchIDs...)
	}

	callbacks.OnEnd(ctx, &indexer.CallbackOutput{
		IDs: allIDs,
	})

	return allIDs, nil
}

func (i *Indexer) GetType() string {
	return typ
}

func (i *Indexer) IsCallbacksEnabled() bool {
	return true
}

// storeBatch stores a batch of entities and relations.
func (i *Indexer) storeBatch(ctx context.Context, storePairs StorePairs, opts *ImplOptions) ([]string, error) {
	database := opts.Database
	if database == "" {
		database = i.config.Database
	}

	session := i.config.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: database,
	})
	defer session.Close(ctx)

	var ids []string

	for _, pair := range storePairs {
		// Apply entity type filter
		if len(opts.EntityTypes) > 0 && (!contains(opts.EntityTypes, pair.From.Type) || !contains(opts.EntityTypes, pair.To.Type)) {
			continue
		}

		id, err := i.store(ctx, session, &pair, opts)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// store stores a single pair using the embedding-based similarity search logic
func (i *Indexer) store(ctx context.Context, session neo4j.SessionWithContext, pair *StorePair, opts *ImplOptions) (string, error) {
	threshold := opts.SimilarityThreshold

	embeddingModel := i.config.EmbeddingModel
	sourceEmbedding, err := embeddingModel.EmbedStrings(ctx, []string{pair.From.Name})
	if err != nil {
		return "", fmt.Errorf("failed to generate source embedding: %w", err)
	}

	destEmbedding, err := embeddingModel.EmbedStrings(ctx, []string{pair.To.Name})
	if err != nil {
		return "", fmt.Errorf("failed to generate destination embedding: %w", err)
	}

	// Search for similar nodes
	sourceNodeResult, err := i.searchSimilarNode(ctx, session, sourceEmbedding[0], pair.From.Type, threshold)
	if err != nil {
		return "", fmt.Errorf("failed to search source node: %w", err)
	}

	destNodeResult, err := i.searchSimilarNode(ctx, session, destEmbedding[0], pair.To.Type, threshold)
	if err != nil {
		return "", fmt.Errorf("failed to search destination node: %w", err)
	}

	var cypher string
	var params map[string]interface{}
	relationshipType := sanitizeRelationshipType(pair.Relation.Type)
	fromType := sanitizeEntityType(pair.From.Type)
	toType := sanitizeEntityType(pair.To.Type)

	// Build cypher query based on search results
	if destNodeResult == nil && sourceNodeResult != nil {
		// Only source node exists
		cypher = fmt.Sprintf(`
			MATCH (source:Entity)
			WHERE id(source) = $source_id
			MERGE (destination:%s:Entity {name: $destination_name})
			ON CREATE SET
				destination.created = timestamp(),
				destination.embedding = $destination_embedding,
				destination:Entity
			MERGE (source)-[r:%s]->(destination)
			ON CREATE SET 
				r.created = timestamp(),
			    r.hit = 1
            ON MATCH SET
                r.hit = coalesce(r.hit, 0) + 1
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, toType, relationshipType)

		params = map[string]interface{}{
			"source_id":             sourceNodeResult.ID,
			"destination_name":      pair.To.Name,
			"destination_embedding": destEmbedding[0],
		}

	} else if destNodeResult != nil && sourceNodeResult == nil {
		// Only destination node exists
		cypher = fmt.Sprintf(`
			MATCH (destination:Entity)
			WHERE id(destination) = $destination_id
			MERGE (source:%s:Entity {name: $source_name})
			ON CREATE SET
				source.created = timestamp(),
				source.embedding = $source_embedding,
				source:Entity
			MERGE (source)-[r:%s]->(destination)
			ON CREATE SET 
				r.created = timestamp(),
			    r.hit = 1
            ON MATCH SET
                r.hit = coalesce(r.hit, 0) + 1
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, fromType, relationshipType)

		params = map[string]interface{}{
			"destination_id":   destNodeResult.ID,
			"source_name":      pair.From.Name,
			"source_embedding": sourceEmbedding[0],
		}

	} else if sourceNodeResult != nil && destNodeResult != nil {
		// Both nodes exist
		cypher = fmt.Sprintf(`
			MATCH (source:Entity)
			WHERE id(source) = $source_id
			MATCH (destination:Entity)
			WHERE id(destination) = $destination_id
			MERGE (source)-[r:%s]->(destination)
			ON CREATE SET 
				r.created_at = timestamp(),
				r.updated_at = timestamp(),
				r.hit = 1
            ON MATCH SET
                r.hit = coalesce(r.hit, 0) + 1
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, relationshipType)

		params = map[string]interface{}{
			"source_id":      sourceNodeResult.ID,
			"destination_id": destNodeResult.ID,
		}

	} else {
		// Both nodes don't exist
		cypher = fmt.Sprintf(`
			MERGE (n:%s:Entity {name: $source_name})
			ON CREATE SET n.created = timestamp(), n.embedding = $source_embedding, n:Entity
			ON MATCH SET n.embedding = $source_embedding
			MERGE (m:%s:Entity {name: $dest_name})
			ON CREATE SET m.created = timestamp(), m.embedding = $dest_embedding, m:Entity
			ON MATCH SET m.embedding = $dest_embedding
			MERGE (n)-[rel:%s]->(m)
			ON CREATE SET 
				rel.created = timestamp(),
				rel.hit = 1
            ON MATCH SET
                rel.hit = coalesce(rel.hit, 0) + 1
			RETURN n.name AS source, type(rel) AS relationship, m.name AS target
		`, fromType, toType, relationshipType)

		params = map[string]interface{}{
			"source_name":      pair.From.Name,
			"dest_name":        pair.To.Name,
			"source_embedding": sourceEmbedding[0],
			"dest_embedding":   destEmbedding[0],
		}
	}

	log.Println("cypher", cypher)
	log.Println("params", params)

	// Execute the cypher query
	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			if source, found := record.Get("source"); found {
				return fmt.Sprintf("%s_%s", source, relationshipType), nil
			}
		}

		return nil, result.Err()
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute cypher query: %w", err)
	}

	return result.(string), nil
}

// searchSimilarNode searches for similar nodes using embedding similarity
func (i *Indexer) searchSimilarNode(ctx context.Context, session neo4j.SessionWithContext, embedding []float64, entityType string, threshold float64) (*SearchResult, error) {
	// Convert embedding to a format suitable for Neo4j
	embeddingList := make([]interface{}, len(embedding))
	for i, v := range embedding {
		embeddingList[i] = v
	}

	// Build cypher query to search for similar nodes
	cypher := `
		MATCH (n:Entity)
		WHERE n.type = $entity_type AND n.embedding IS NOT NULL
		CALL db.index.vector.queryNodes('entity_embedding_index', 5, $embedding)
		YIELD node, score
		RETURN id(node) as id, node.name as name, node.type as type, score
		ORDER BY score DESC
		LIMIT 1
	`

	params := map[string]interface{}{
		"entity_type": entityType,
		"embedding":   embeddingList,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			searchResult := &SearchResult{}

			if id, found := record.Get("id"); found {
				searchResult.ID = id.(int64)
			}
			if name, found := record.Get("name"); found {
				searchResult.Name = name.(string)
			}
			if nodeType, found := record.Get("type"); found {
				searchResult.Type = nodeType.(string)
			}
			if score, found := record.Get("score"); found {
				searchResult.Score = score.(float64)
			}

			return searchResult, nil
		}

		return nil, result.Err()
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	searchResult := result.(*SearchResult)
	if searchResult.Score < threshold {
		return nil, nil
	}

	return searchResult, nil
}

// getDefaultDocumentConverter returns the default document converter
func (i *IndexerConfig) getDefaultDocumentConverter() func(ctx context.Context, docs []*schema.Document) (StorePairs, error) {
	return func(ctx context.Context, docs []*schema.Document) (StorePairs, error) {
		results := make(StorePairs, 0)

		for _, doc := range docs {
			storePairs, err := i.extractEntitiesAndRelations(ctx, doc)
			if err != nil {
				return nil, err
			}

			results = append(results, storePairs...)
		}

		return results, nil
	}
}

// extractEntitiesAndRelations extracts entities and relations from a document using LLM
func (i *IndexerConfig) extractEntitiesAndRelations(ctx context.Context, doc *schema.Document) (StorePairs, error) {
	// Use the prompt from models to extract entities and relations
	msgs, err := PromptGraphExtractEntityAndRelation.Format(ctx, map[string]any{
		"origin_request": doc.Content,
		"user_key":       ctxwarp.GetHeaderContext(ctx).UserKey,
		"resp_schema":    i.respSchema,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, msg := range msgs {
		log.Println("msg", msg.Content)
	}

	output, err := i.Extractor.Generate(ctx, msgs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	log.Println("output", output.Content)

	extractEntityAndRelation := ExtractEntityAndRelation{}
	if err := sonic.Unmarshal([]byte(output.Content), &extractEntityAndRelation); err != nil {
		return nil, errors.WithStack(err)
	}

	pairs := make(StorePairs, 0)
	for _, relation := range extractEntityAndRelation.Relations {
		fromEntity := Entity{}
		toEntity := Entity{}
		for _, entity := range extractEntityAndRelation.Entities {
			if entity.Name == relation.FromEntityName {
				fromEntity = entity
			}
			if entity.Name == relation.ToEntityName {
				toEntity = entity
			}
		}
		if fromEntity.Name == "" || toEntity.Name == "" {
			continue
		}
		pairs = append(pairs, StorePair{
			From:     fromEntity,
			To:       toEntity,
			Relation: relation,
		})
	}

	return pairs, nil
}

// ensureDatabaseExists checks if the database exists and creates it if not
func (i *IndexerConfig) ensureDatabaseExists(ctx context.Context) error {
	// Connect to system database to manage other databases
	// Use write mode since we might need to create database
	systemSession := i.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "system",
	})
	defer systemSession.Close(ctx)

	// Check if database exists using efficient query
	databaseExists, err := systemSession.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, "SHOW DATABASES YIELD name WHERE name = $db_name", map[string]any{
			"db_name": i.Database,
		})
		if err != nil {
			return false, err
		}

		// If we get any result, the database exists
		exists := result.Next(ctx)
		if err := result.Err(); err != nil {
			return false, err
		}
		return exists, nil
	})

	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	// Create database if it doesn't exist
	if !databaseExists.(bool) {
		_, err := systemSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			query := fmt.Sprintf("CREATE DATABASE `%s`", i.Database)
			_, err := tx.Run(ctx, query, nil)
			return nil, err
		})
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	return nil
}

// createIndexes creates necessary indexes for better performance
func (i *IndexerConfig) createIndexes(ctx context.Context, dimension int) error {
	session := i.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: i.Database,
	})
	defer session.Close(ctx)

	indexes := []string{
		"CREATE INDEX entity_id_index IF NOT EXISTS FOR (e:Entity) ON (e.id)",
		"CREATE INDEX entity_type_index IF NOT EXISTS FOR (e:Entity) ON (e.type)",
		"CREATE INDEX entity_name_index IF NOT EXISTS FOR (e:Entity) ON (e.name)",
		fmt.Sprintf("CREATE VECTOR INDEX entity_embedding_index IF NOT EXISTS FOR (e:Entity) ON (e.embedding) OPTIONS { indexConfig: { `vector.dimensions`: %d, `vector.similarity_function`: 'cosine' }}", dimension),
		"CREATE INDEX relation_id_index IF NOT EXISTS FOR ()-[r:RELATION]-() ON (r.id)",
		"CREATE INDEX relation_type_index IF NOT EXISTS FOR ()-[r:RELATION]-() ON (r.type)",
	}

	for _, indexQuery := range indexes {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, indexQuery, nil)
			if err != nil {
				return nil, err
			}
			return result.Consume(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// check validates the indexer configuration
func (i *IndexerConfig) check() error {
	if i.Driver == nil {
		return fmt.Errorf("[NewIndexer] Neo4j driver not provided")
	}
	if i.Extractor == nil {
		return fmt.Errorf("[NewIndexer] entity extractor not provided")
	}
	if i.EmbeddingModel == nil {
		return fmt.Errorf("[NewIndexer] embedding model not provided")
	}
	if i.Dimension <= 0 {
		i.Dimension = 2048
	}
	if i.Database == "" {
		i.Database = "neo4j"
	}
	if i.MaxRetries <= 0 {
		i.MaxRetries = 3
	}
	if i.RetryDelay <= 0 {
		i.RetryDelay = time.Second
	}
	if i.BatchSize <= 0 {
		i.BatchSize = 100
	}
	if i.DocumentConverter == nil {
		i.DocumentConverter = i.getDefaultDocumentConverter()
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SearchResult represents the result of a similarity search
type SearchResult struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Score float64 `json:"score"`
}

func sanitizeEntityType(entityType string) string {
	entityType = strings.ReplaceAll(entityType, "（", "(")
	entityType = strings.ReplaceAll(entityType, "）", ")")
	entityType = strings.ReplaceAll(entityType, "(", "_")
	entityType = strings.ReplaceAll(entityType, ")", "_")
	entityType = strings.ReplaceAll(entityType, " ", "_")
	entityType = strings.ReplaceAll(entityType, "-", "_")
	entityType = strings.ReplaceAll(entityType, "，", "_")
	return entityType
}

// sanitizeRelationshipType cleans relationship type names to make them valid for Neo4j
func sanitizeRelationshipType(relType string) string {
	// Replace Chinese parentheses with English ones
	relType = strings.ReplaceAll(relType, "（", "(")
	relType = strings.ReplaceAll(relType, "）", ")")

	// Remove or replace invalid characters for Neo4j relationship types
	// Neo4j allows alphanumeric characters, underscores, and some unicode characters
	// but parentheses and some special characters need to be handled
	relType = strings.ReplaceAll(relType, "(", "_")
	relType = strings.ReplaceAll(relType, ")", "_")
	relType = strings.ReplaceAll(relType, " ", "_")
	relType = strings.ReplaceAll(relType, "-", "_")
	relType = strings.ReplaceAll(relType, "，", "_")
	relType = strings.ReplaceAll(relType, ",", "_")

	// Remove multiple consecutive underscores
	re := regexp.MustCompile(`_+`)
	relType = re.ReplaceAllString(relType, "_")

	// Remove leading and trailing underscores
	relType = strings.Trim(relType, "_")

	// Ensure it's not empty
	if relType == "" {
		relType = "RELATED_TO"
	}

	return relType
}
