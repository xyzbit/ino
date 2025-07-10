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
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
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
	Database string

	// Extractor is the model used to extract entities and relations from documents
	// Required
	Extractor model.ToolCallingChatModel

	// EmbeddingModel is the model used to generate embeddings for entities
	// Required
	EmbeddingModel embedding.Embedder

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

	// MinConfidence specifies the minimum confidence score for entities/relations
	MinConfidence float64

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

	// Create indexes for better performance
	if err := conf.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("[NewIndexer] failed to create indexes: %w", err)
	}

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
		// Apply confidence filter
		if opts.MinConfidence > 0 && pair.Confidence < opts.MinConfidence {
			continue
		}

		// Apply entity type filter
		if len(opts.EntityTypes) > 0 && (!contains(opts.EntityTypes, pair.From.Type) || !contains(opts.EntityTypes, pair.To.Type)) {
			continue
		}

		id, err := i.store(ctx, session, &pair, opts.SimilarityThreshold)
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
	// if opts

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

	// Prepare agent_id clause for node creation
	agentIDClause := ""
	if agentID != "" {
		agentIDClause = ", agent_id: $agent_id"
	}

	var cypher string
	var params map[string]interface{}
	relationshipType := pair.Relation.Type

	// Build cypher query based on search results
	if destNodeResult == nil && sourceNodeResult != nil {
		// Only source node exists
		cypher = fmt.Sprintf(`
			MATCH (source:Entity)
			WHERE id(source) = $source_id
			MERGE (destination:%s:Entity {name: $destination_name, user_id: $user_id%s})
			ON CREATE SET
				destination.created = timestamp(),
				destination.embedding = $destination_embedding,
				destination:Entity
			MERGE (source)-[r:%s]->(destination)
			ON CREATE SET 
				r.created = timestamp()
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, pair.To.Type, agentIDClause, relationshipType)

		params = map[string]interface{}{
			"source_id":             sourceNodeResult.ID,
			"destination_name":      pair.To.Name,
			"destination_embedding": destEmbedding[0],
			"user_id":               userID,
		}
		if agentID != "" {
			params["agent_id"] = agentID
		}

	} else if destNodeResult != nil && sourceNodeResult == nil {
		// Only destination node exists
		cypher = fmt.Sprintf(`
			MATCH (destination:Entity)
			WHERE id(destination) = $destination_id
			MERGE (source:%s:Entity {name: $source_name, user_id: $user_id%s})
			ON CREATE SET
				source.created = timestamp(),
				source.embedding = $source_embedding,
				source:Entity
			MERGE (source)-[r:%s]->(destination)
			ON CREATE SET 
				r.created = timestamp()
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, pair.From.Type, agentIDClause, relationshipType)

		params = map[string]interface{}{
			"destination_id":   destNodeResult.ID,
			"source_name":      pair.From.Name,
			"source_embedding": sourceEmbedding[0],
			"user_id":          userID,
		}
		if agentID != "" {
			params["agent_id"] = agentID
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
				r.updated_at = timestamp()
			RETURN source.name AS source, type(r) AS relationship, destination.name AS target
		`, relationshipType)

		params = map[string]interface{}{
			"source_id":      sourceNodeResult.ID,
			"destination_id": destNodeResult.ID,
			"user_id":        userID,
		}
		if agentID != "" {
			params["agent_id"] = agentID
		}

	} else {
		// Both nodes don't exist
		cypher = fmt.Sprintf(`
			MERGE (n:%s:Entity {name: $source_name, user_id: $user_id%s})
			ON CREATE SET n.created = timestamp(), n.embedding = $source_embedding, n:Entity
			ON MATCH SET n.embedding = $source_embedding
			MERGE (m:%s:Entity {name: $dest_name, user_id: $user_id%s})
			ON CREATE SET m.created = timestamp(), m.embedding = $dest_embedding, m:Entity
			ON MATCH SET m.embedding = $dest_embedding
			MERGE (n)-[rel:%s]->(m)
			ON CREATE SET rel.created = timestamp()
			RETURN n.name AS source, type(rel) AS relationship, m.name AS target
		`, pair.From.Type, agentIDClause, pair.To.Type, agentIDClause, relationshipType)

		params = map[string]interface{}{
			"source_name":      pair.From.Name,
			"dest_name":        pair.To.Name,
			"source_embedding": sourceEmbedding[0],
			"dest_embedding":   destEmbedding[0],
			"user_id":          userID,
		}
		if agentID != "" {
			params["agent_id"] = agentID
		}
	}

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
		WHERE n.type = $entity_type AND exists(n.embedding)
		WITH n, gds.similarity.cosine(n.embedding, $embedding) AS similarity
		WHERE similarity >= $threshold
		RETURN id(n) as id, n.name as name, n.type as type, similarity
		ORDER BY similarity DESC
		LIMIT 1
	`

	params := map[string]interface{}{
		"entity_type": entityType,
		"embedding":   embeddingList,
		"threshold":   threshold,
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
			if score, found := record.Get("similarity"); found {
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

	return result.(*SearchResult), nil
}

// getDefaultDocumentConverter returns the default document converter
// TODO: 默认添加一个来源于哪个文档id或摘要的关系.
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
	if doc.MetaData["is_preference"] != nil { // is user preference doc.
		if isPreference := doc.MetaData["is_preference"].(bool); isPreference {
			header := ctxwarp.GetHeaderContext(ctx)
			if header == nil || header.User == "" {
				return StorePairs{}, nil
			}
		}
	}
	// Use the prompt from models to extract entities and relations
	msgs, err := PromptGraphExtractEntityAndRelation.Format(ctx, map[string]any{
		"origin_request": doc.Content,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	toolInfo, err := utils.GoStruct2ToolInfo[StorePairs](
		"store_entity_and_relation",
		"保存提取的实体和关系, Store entities and relations based on the provided text.",
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// utils.InferTool("add_user", "add user", AddUser)
	output, err := i.Extractor.Generate(ctx, msgs, model.WithTools([]*schema.ToolInfo{toolInfo}))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pairs := make(StorePairs, 0)
	for _, toolCall := range output.ToolCalls {
		if toolCall.Function.Name != "store_entity_and_relation" {
			continue
		}

		if err := sonic.Unmarshal([]byte(toolCall.Function.Arguments), pairs); err != nil {
			return nil, errors.WithStack(err)
		}
		pairs = append(pairs, pairs...)
	}

	return pairs, nil
}

// createIndexes creates necessary indexes for better performance
func (i *IndexerConfig) createIndexes(ctx context.Context) error {
	session := i.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: i.Database,
	})
	defer session.Close(ctx)

	indexes := []string{
		"CREATE INDEX entity_id_index IF NOT EXISTS FOR (e:Entity) ON (e.id)",
		"CREATE INDEX entity_type_index IF NOT EXISTS FOR (e:Entity) ON (e.type)",
		"CREATE INDEX entity_name_index IF NOT EXISTS FOR (e:Entity) ON (e.name)",
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
	if i.SimilarityThreshold <= 0 {
		i.SimilarityThreshold = 0.9
	}
	if i.DocumentConverter == nil {
		i.DocumentConverter = i.getDefaultDocumentConverter()
	}
	return nil
}

// Utility functions

func generateEntityID(name, entityType string) string {
	if name == "" {
		return uuid.New().String()
	}
	return fmt.Sprintf("%s_%s_%s", entityType, name, uuid.New().String()[:8])
}

func generateRelationID(from, to, relationType string) string {
	return fmt.Sprintf("%s_%s_%s_%s", from, relationType, to, uuid.New().String()[:8])
}

func getDocumentSource(doc *schema.Document) string {
	if source, ok := doc.MetaData["source"]; ok {
		if str, ok := source.(string); ok {
			return str
		}
	}
	return "unknown"
}

func getDocumentTitle(doc *schema.Document) string {
	if title, ok := doc.MetaData["title"]; ok {
		if str, ok := title.(string); ok {
			return str
		}
	}
	return fmt.Sprintf("Document_%s", doc.ID)
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
