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
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pkg/errors"
	"github.com/xyzbit/ino/internal/domain/models"
)

const typ = "neo4j"

type IndexerConfig struct {
	// Driver is the Neo4j driver to be called
	// Required
	Driver neo4j.DriverWithContext

	// Database is the database name in Neo4j
	// Optional, default is "neo4j"
	Database string

	// EntityExtractor is the model used to extract entities and relations from documents
	// Required
	EntityExtractor model.ToolCallingChatModel

	// DocumentConverter is the model used to convert documents to entities and relations
	// Required
	DocumentConverter func(ctx context.Context, docs []*schema.Document) ([]*StoreRequest, error)

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

	// Source specifies the source of the documents
	Source string
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
	io := indexer.GetImplSpecificOptions(&ImplOptions{}, opts...)

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
	entities, relations, err := i.config.DocumentConverter(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("[Indexer.Store] failed to convert documents: %w", err)
	}

	// TODO: 查询当前数据库中相关的老数据，将老数据和新的数据传递给模型判断哪些数据需要删除。

	// Process in batches
	var allIDs []string
	batchSize := i.config.BatchSize

	// Process entities
	for idx := 0; idx < len(entities); idx += batchSize {
		end := idx + batchSize
		if end > len(entities) {
			end = len(entities)
		}

		batch := entities[idx:end]
		batchIDs, err := i.storeEntitiesBatch(ctx, batch, io)
		if err != nil {
			return nil, fmt.Errorf("[Indexer.Store] failed to store entities batch: %w", err)
		}
		allIDs = append(allIDs, batchIDs...)
	}

	// Process relations
	for idx := 0; idx < len(relations); idx += batchSize {
		end := idx + batchSize
		if end > len(relations) {
			end = len(relations)
		}

		batch := relations[idx:end]
		batchIDs, err := i.storeRelationsBatch(ctx, batch, io)
		if err != nil {
			return nil, fmt.Errorf("[Indexer.Store] failed to store relations batch: %w", err)
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

// storeEntitiesBatch stores a batch of entities
func (i *Indexer) storeEntitiesBatch(ctx context.Context, entities []*models.KnowledgeEntity, opts *ImplOptions) ([]string, error) {
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

	for _, entity := range entities {
		// Apply confidence filter
		if opts.MinConfidence > 0 && entity.Score < opts.MinConfidence {
			continue
		}

		// Apply entity type filter
		if len(opts.EntityTypes) > 0 && !contains(opts.EntityTypes, entity.Type) {
			continue
		}

		// Override source if specified
		if opts.Source != "" {
			entity.Source = opts.Source
		}

		id, err := i.storeEntity(ctx, session, entity)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// storeRelationsBatch stores a batch of relations
func (i *Indexer) storeRelationsBatch(ctx context.Context, relations []*models.KnowledgeRelation, opts *ImplOptions) ([]string, error) {
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

	for _, relation := range relations {
		// Apply confidence filter
		if opts.MinConfidence > 0 && relation.Score < opts.MinConfidence {
			continue
		}

		// Apply relation type filter
		if len(opts.RelationTypes) > 0 && !contains(opts.RelationTypes, relation.Type) {
			continue
		}

		// Override source if specified
		if opts.Source != "" {
			relation.Source = opts.Source
		}

		id, err := i.storeRelation(ctx, session, relation)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// storeEntity stores a single entity
func (i *Indexer) storeEntity(ctx context.Context, session neo4j.SessionWithContext, entity *models.KnowledgeEntity) (string, error) {
	// Build labels
	labels := "Entity"
	if entity.Type != "" {
		labels += ":" + entity.Type
	}
	for _, label := range entity.Labels {
		if label != "" {
			labels += ":" + label
		}
	}

	cypher := fmt.Sprintf(`
		MERGE (e:%s {id: $id})
		SET e += $props
		RETURN e.id as id
	`, labels)

	props := map[string]interface{}{
		"id":         entity.ID,
		"name":       entity.Name,
		"type":       entity.Type,
		"properties": entity.Properties,
		"source":     entity.Source,
		"score":      entity.Score,
		"created_at": entity.CreatedAt,
		"updated_at": time.Now(),
	}

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":    entity.ID,
			"props": props,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			if id, found := record.Get("id"); found {
				return id.(string), nil
			}
		}

		return nil, result.Err()
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// storeRelation stores a single relation
func (i *Indexer) storeRelation(ctx context.Context, session neo4j.SessionWithContext, relation *models.KnowledgeRelation) (string, error) {
	cypher := `
		MATCH (from:Entity {id: $from_id}), (to:Entity {id: $to_id})
		MERGE (from)-[r:RELATION {id: $id}]->(to)
		SET r += $props
		RETURN r.id as id
	`

	props := map[string]interface{}{
		"id":         relation.ID,
		"type":       relation.Type,
		"properties": relation.Properties,
		"source":     relation.Source,
		"score":      relation.Score,
		"created_at": relation.CreatedAt,
		"updated_at": time.Now(),
	}

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":      relation.ID,
			"from_id": relation.FromEntity,
			"to_id":   relation.ToEntity,
			"props":   props,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			if id, found := record.Get("id"); found {
				return id.(string), nil
			}
		}

		return nil, result.Err()
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
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
	output, err := i.EntityExtractor.Generate(ctx, msgs, model.WithTools([]*schema.ToolInfo{toolInfo}))
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
	if i.EntityExtractor == nil {
		return fmt.Errorf("[NewIndexer] entity extractor not provided")
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
