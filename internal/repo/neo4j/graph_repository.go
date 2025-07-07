package neo4j

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/xyzbit/ino/internal/domain/models"
	"github.com/xyzbit/ino/internal/domain/repository"
	neo4jClient "github.com/xyzbit/ino/internal/infra/neo4j"
)

type graphRepository struct {
	driver neo4j.DriverWithContext
}

// NewGraphRepository 创建图数据库仓储实例
func NewGraphRepository(driver neo4j.DriverWithContext) repository.GraphRepository {
	return &graphRepository{driver: driver}
}

// CreateEntity 创建实体
func (r *graphRepository) CreateEntity(ctx context.Context, entity *models.KnowledgeEntity) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	// 构建标签字符串
	labels := "Entity"
	if entity.Type != "" {
		labels += ":" + entity.Type
	}
	for _, label := range entity.Labels {
		if label != "" {
			labels += ":" + label
		}
	}

	cypher := fmt.Sprintf("CREATE (e:%s) SET e = $props RETURN e", labels)

	props := map[string]interface{}{
		"id":         entity.ID,
		"name":       entity.Name,
		"type":       entity.Type,
		"properties": entity.Properties,
		"source":     entity.Source,
		"score":      entity.Score,
		"created_at": entity.CreatedAt,
		"updated_at": entity.UpdatedAt,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"props": props})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	return err
}

// GetEntity 获取实体
func (r *graphRepository) GetEntity(ctx context.Context, id string) (*models.KnowledgeEntity, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH (e:Entity {id: $id}) RETURN e"

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			entityNode, found := record.Get("e")
			if !found {
				return nil, fmt.Errorf("entity not found")
			}

			return entityNode, nil
		}

		return nil, result.Err()
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("entity not found")
	}

	// 转换Neo4j节点为实体
	node := result.(neo4j.Node)

	entity := &models.KnowledgeEntity{
		ID:         getStringProp(node.Props, "id"),
		Type:       getStringProp(node.Props, "type"),
		Name:       getStringProp(node.Props, "name"),
		Labels:     node.Labels[1:], // 去掉第一个"Entity"标签
		Properties: getMapProp(node.Props, "properties"),
		Source:     getStringProp(node.Props, "source"),
		Score:      getFloatProp(node.Props, "score"),
		CreatedAt:  getTimeProp(node.Props, "created_at"),
		UpdatedAt:  getTimeProp(node.Props, "updated_at"),
	}

	return entity, nil
}

// UpdateEntity 更新实体
func (r *graphRepository) UpdateEntity(ctx context.Context, entity *models.KnowledgeEntity) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH (e:Entity {id: $id}) SET e += $props, e.updated_at = $updated_at RETURN e"

	props := map[string]interface{}{
		"name":       entity.Name,
		"type":       entity.Type,
		"properties": entity.Properties,
		"source":     entity.Source,
		"score":      entity.Score,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":         entity.ID,
			"props":      props,
			"updated_at": time.Now(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	return err
}

// DeleteEntity 删除实体
func (r *graphRepository) DeleteEntity(ctx context.Context, id string) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH (e:Entity {id: $id}) DETACH DELETE e"

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		return summary.Counters().NodesDeleted(), nil
	})

	return err
}

// ListEntities 列出实体
func (r *graphRepository) ListEntities(ctx context.Context, entityType string, offset, limit int) ([]*models.KnowledgeEntity, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	var cypher string
	var params map[string]any

	if entityType != "" {
		cypher = "MATCH (e:Entity {type: $type}) RETURN e ORDER BY e.created_at DESC SKIP $offset LIMIT $limit"
		params = map[string]any{
			"type":   entityType,
			"offset": offset,
			"limit":  limit,
		}
	} else {
		cypher = "MATCH (e:Entity) RETURN e ORDER BY e.created_at DESC SKIP $offset LIMIT $limit"
		params = map[string]any{
			"offset": offset,
			"limit":  limit,
		}
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var entities []*models.KnowledgeEntity
		for result.Next(ctx) {
			record := result.Record()

			entityNode, found := record.Get("e")
			if !found {
				continue
			}

			node := entityNode.(neo4j.Node)

			entity := &models.KnowledgeEntity{
				ID:         getStringProp(node.Props, "id"),
				Type:       getStringProp(node.Props, "type"),
				Name:       getStringProp(node.Props, "name"),
				Labels:     node.Labels[1:], // 去掉第一个"Entity"标签
				Properties: getMapProp(node.Props, "properties"),
				Source:     getStringProp(node.Props, "source"),
				Score:      getFloatProp(node.Props, "score"),
				CreatedAt:  getTimeProp(node.Props, "created_at"),
				UpdatedAt:  getTimeProp(node.Props, "updated_at"),
			}

			entities = append(entities, entity)
		}

		return entities, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.KnowledgeEntity), nil
}

// CreateRelation 创建关系
func (r *graphRepository) CreateRelation(ctx context.Context, relation *models.KnowledgeRelation) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH (a:Entity {id: $from_id}), (b:Entity {id: $to_id}) CREATE (a)-[r:RELATION]->(b) SET r = $props RETURN r"

	props := map[string]interface{}{
		"id":         relation.ID,
		"type":       relation.Type,
		"properties": relation.Properties,
		"source":     relation.Source,
		"score":      relation.Score,
		"created_at": relation.CreatedAt,
		"updated_at": relation.UpdatedAt,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"from_id": relation.FromEntity,
			"to_id":   relation.ToEntity,
			"props":   props,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	return err
}

// GetRelation 获取关系
func (r *graphRepository) GetRelation(ctx context.Context, id string) (*models.KnowledgeRelation, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH (a:Entity)-[r:RELATION {id: $id}]->(b:Entity) RETURN r, a.id as from_id, b.id as to_id"

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()

			relNode, _ := record.Get("r")
			fromID, _ := record.Get("from_id")
			toID, _ := record.Get("to_id")

			return map[string]interface{}{
				"relation": relNode,
				"from_id":  fromID,
				"to_id":    toID,
			}, nil
		}

		return nil, result.Err()
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("relation not found")
	}

	resultMap := result.(map[string]interface{})
	rel := resultMap["relation"].(neo4j.Relationship)

	relation := &models.KnowledgeRelation{
		ID:         getStringProp(rel.Props, "id"),
		Type:       getStringProp(rel.Props, "type"),
		FromEntity: resultMap["from_id"].(string),
		ToEntity:   resultMap["to_id"].(string),
		Properties: getMapProp(rel.Props, "properties"),
		Source:     getStringProp(rel.Props, "source"),
		Score:      getFloatProp(rel.Props, "score"),
		CreatedAt:  getTimeProp(rel.Props, "created_at"),
		UpdatedAt:  getTimeProp(rel.Props, "updated_at"),
	}

	return relation, nil
}

// UpdateRelation 更新关系
func (r *graphRepository) UpdateRelation(ctx context.Context, relation *models.KnowledgeRelation) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH ()-[r:RELATION {id: $id}]->() SET r += $props, r.updated_at = $updated_at RETURN r"

	props := map[string]interface{}{
		"type":       relation.Type,
		"properties": relation.Properties,
		"source":     relation.Source,
		"score":      relation.Score,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":         relation.ID,
			"props":      props,
			"updated_at": time.Now(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, result.Err()
	})

	return err
}

// DeleteRelation 删除关系
func (r *graphRepository) DeleteRelation(ctx context.Context, id string) error {
	session := neo4jClient.GetSession(ctx)
	defer session.Close(ctx)

	cypher := "MATCH ()-[r:RELATION {id: $id}]->() DELETE r"

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		return summary.Counters().RelationshipsDeleted(), nil
	})

	return err
}

// ListRelations 列出关系
func (r *graphRepository) ListRelations(ctx context.Context, fromEntity, toEntity, relationType string) ([]*models.KnowledgeRelation, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	// 构建查询条件
	var whereConditions []string
	params := make(map[string]any)

	if fromEntity != "" {
		whereConditions = append(whereConditions, "a.id = $from_entity")
		params["from_entity"] = fromEntity
	}

	if toEntity != "" {
		whereConditions = append(whereConditions, "b.id = $to_entity")
		params["to_entity"] = toEntity
	}

	if relationType != "" {
		whereConditions = append(whereConditions, "r.type = $relation_type")
		params["relation_type"] = relationType
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	cypher := fmt.Sprintf("MATCH (a:Entity)-[r:RELATION]->(b:Entity)%s RETURN r, a.id as from_id, b.id as to_id ORDER BY r.created_at DESC", whereClause)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var relations []*models.KnowledgeRelation
		for result.Next(ctx) {
			record := result.Record()

			relNode, _ := record.Get("r")
			fromID, _ := record.Get("from_id")
			toID, _ := record.Get("to_id")

			rel := relNode.(neo4j.Relationship)

			relation := &models.KnowledgeRelation{
				ID:         getStringProp(rel.Props, "id"),
				Type:       getStringProp(rel.Props, "type"),
				FromEntity: fromID.(string),
				ToEntity:   toID.(string),
				Properties: getMapProp(rel.Props, "properties"),
				Source:     getStringProp(rel.Props, "source"),
				Score:      getFloatProp(rel.Props, "score"),
				CreatedAt:  getTimeProp(rel.Props, "created_at"),
				UpdatedAt:  getTimeProp(rel.Props, "updated_at"),
			}

			relations = append(relations, relation)
		}

		return relations, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.KnowledgeRelation), nil
}

// TraverseGraph 图遍历
func (r *graphRepository) TraverseGraph(ctx context.Context, config *models.GraphTraversal) (*models.GraphTraversalResult, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	// 构建遍历查询
	cypher := "MATCH path = (start:Entity {id: $start_id})"

	// 添加方向和深度
	switch config.Direction {
	case "OUT":
		cypher += "-[*1.." + fmt.Sprintf("%d", config.MaxDepth) + "]->(end:Entity)"
	case "IN":
		cypher += "<-[*1.." + fmt.Sprintf("%d", config.MaxDepth) + "]-(end:Entity)"
	case "BOTH":
		cypher += "-[*1.." + fmt.Sprintf("%d", config.MaxDepth) + "]-(end:Entity)"
	default:
		cypher += "-[*1..3]->(end:Entity)"
	}

	// 添加过滤条件
	var whereConditions []string
	params := map[string]any{"start_id": config.StartEntity}

	if config.MinScore > 0 {
		whereConditions = append(whereConditions, "all(rel in relationships(path) WHERE rel.score >= $min_score)")
		params["min_score"] = config.MinScore
	}

	if len(whereConditions) > 0 {
		cypher += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	cypher += " RETURN path"

	if config.Limit > 0 {
		cypher += fmt.Sprintf(" LIMIT %d", config.Limit)
	}

	startTime := time.Now()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var paths []models.GraphPath
		for result.Next(ctx) {
			record := result.Record()

			pathValue, found := record.Get("path")
			if !found {
				continue
			}

			neo4jPath := pathValue.(neo4j.Path)

			// 转换路径
			var entities []models.KnowledgeEntity
			var relations []models.KnowledgeRelation

			for _, node := range neo4jPath.Nodes {
				entity := models.KnowledgeEntity{
					ID:         getStringProp(node.Props, "id"),
					Type:       getStringProp(node.Props, "type"),
					Name:       getStringProp(node.Props, "name"),
					Labels:     node.Labels[1:], // 去掉"Entity"标签
					Properties: getMapProp(node.Props, "properties"),
					Source:     getStringProp(node.Props, "source"),
					Score:      getFloatProp(node.Props, "score"),
					CreatedAt:  getTimeProp(node.Props, "created_at"),
					UpdatedAt:  getTimeProp(node.Props, "updated_at"),
				}
				entities = append(entities, entity)
			}

			for _, rel := range neo4jPath.Relationships {
				relation := models.KnowledgeRelation{
					ID:         getStringProp(rel.Props, "id"),
					Type:       getStringProp(rel.Props, "type"),
					Properties: getMapProp(rel.Props, "properties"),
					Source:     getStringProp(rel.Props, "source"),
					Score:      getFloatProp(rel.Props, "score"),
					CreatedAt:  getTimeProp(rel.Props, "created_at"),
					UpdatedAt:  getTimeProp(rel.Props, "updated_at"),
				}
				relations = append(relations, relation)
			}

			path := models.GraphPath{
				Entities:  entities,
				Relations: relations,
				Length:    len(neo4jPath.Relationships),
			}

			paths = append(paths, path)
		}

		return paths, result.Err()
	})

	if err != nil {
		return nil, err
	}

	paths := result.([]models.GraphPath)

	// 计算统计信息
	totalPaths := len(paths)
	avgPathLength := 0.0
	maxDepth := 0

	if totalPaths > 0 {
		totalLength := 0
		for _, path := range paths {
			totalLength += path.Length
			if path.Length > maxDepth {
				maxDepth = path.Length
			}
		}
		avgPathLength = float64(totalLength) / float64(totalPaths)
	}

	traversalResult := &models.GraphTraversalResult{
		Paths: paths,
	}
	traversalResult.Stats.TotalPaths = totalPaths
	traversalResult.Stats.AvgPathLength = avgPathLength
	traversalResult.Stats.MaxDepth = maxDepth
	traversalResult.Stats.ExecutionTime = int(time.Since(startTime).Milliseconds())

	return traversalResult, nil
}

// FindPath 查找路径
func (r *graphRepository) FindPath(ctx context.Context, fromEntity, toEntity string, maxDepth int) ([]*models.GraphPath, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	cypher := fmt.Sprintf(
		"MATCH path = shortestPath((start:Entity {id: $from_id})-[*1..%d]-(end:Entity {id: $to_id})) RETURN path",
		maxDepth,
	)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"from_id": fromEntity,
			"to_id":   toEntity,
		})
		if err != nil {
			return nil, err
		}

		var paths []*models.GraphPath
		for result.Next(ctx) {
			record := result.Record()

			pathValue, found := record.Get("path")
			if !found {
				continue
			}

			neo4jPath := pathValue.(neo4j.Path)

			// 转换路径
			var entities []models.KnowledgeEntity
			var relations []models.KnowledgeRelation

			for _, node := range neo4jPath.Nodes {
				entity := models.KnowledgeEntity{
					ID:         getStringProp(node.Props, "id"),
					Type:       getStringProp(node.Props, "type"),
					Name:       getStringProp(node.Props, "name"),
					Labels:     node.Labels[1:],
					Properties: getMapProp(node.Props, "properties"),
					Source:     getStringProp(node.Props, "source"),
					Score:      getFloatProp(node.Props, "score"),
					CreatedAt:  getTimeProp(node.Props, "created_at"),
					UpdatedAt:  getTimeProp(node.Props, "updated_at"),
				}
				entities = append(entities, entity)
			}

			for _, rel := range neo4jPath.Relationships {
				relation := models.KnowledgeRelation{
					ID:         getStringProp(rel.Props, "id"),
					Type:       getStringProp(rel.Props, "type"),
					Properties: getMapProp(rel.Props, "properties"),
					Source:     getStringProp(rel.Props, "source"),
					Score:      getFloatProp(rel.Props, "score"),
					CreatedAt:  getTimeProp(rel.Props, "created_at"),
					UpdatedAt:  getTimeProp(rel.Props, "updated_at"),
				}
				relations = append(relations, relation)
			}

			path := &models.GraphPath{
				Entities:  entities,
				Relations: relations,
				Length:    len(neo4jPath.Relationships),
			}

			paths = append(paths, path)
		}

		return paths, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.GraphPath), nil
}

// SearchEntities 搜索实体
func (r *graphRepository) SearchEntities(ctx context.Context, query string, entityTypes []string, limit int) ([]*models.KnowledgeEntity, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	// 构建查询条件
	var whereConditions []string
	params := make(map[string]any)

	// 文本搜索条件
	whereConditions = append(whereConditions, "(e.name CONTAINS $query OR any(prop in keys(e.properties) WHERE toString(e.properties[prop]) CONTAINS $query))")
	params["query"] = query

	// 实体类型过滤
	if len(entityTypes) > 0 {
		whereConditions = append(whereConditions, "e.type IN $entity_types")
		params["entity_types"] = entityTypes
	}

	whereClause := " WHERE " + strings.Join(whereConditions, " AND ")

	cypher := "MATCH (e:Entity)" + whereClause + " RETURN e ORDER BY e.score DESC"

	if limit > 0 {
		cypher += fmt.Sprintf(" LIMIT %d", limit)
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var entities []*models.KnowledgeEntity
		for result.Next(ctx) {
			record := result.Record()

			entityNode, found := record.Get("e")
			if !found {
				continue
			}

			node := entityNode.(neo4j.Node)

			entity := &models.KnowledgeEntity{
				ID:         getStringProp(node.Props, "id"),
				Type:       getStringProp(node.Props, "type"),
				Name:       getStringProp(node.Props, "name"),
				Labels:     node.Labels[1:],
				Properties: getMapProp(node.Props, "properties"),
				Source:     getStringProp(node.Props, "source"),
				Score:      getFloatProp(node.Props, "score"),
				CreatedAt:  getTimeProp(node.Props, "created_at"),
				UpdatedAt:  getTimeProp(node.Props, "updated_at"),
			}

			entities = append(entities, entity)
		}

		return entities, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.KnowledgeEntity), nil
}

// GetGraphStats 获取图统计信息
func (r *graphRepository) GetGraphStats(ctx context.Context) (*models.GraphStats, error) {
	session := neo4jClient.GetReadSession(ctx)
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 获取实体数量
		entityCountResult, err := tx.Run(ctx, "MATCH (e:Entity) RETURN count(e) as entity_count", nil)
		if err != nil {
			return nil, err
		}

		var entityCount int
		if entityCountResult.Next(ctx) {
			record := entityCountResult.Record()
			if count, found := record.Get("entity_count"); found {
				entityCount = int(count.(int64))
			}
		}

		// 获取关系数量
		relationCountResult, err := tx.Run(ctx, "MATCH ()-[r:RELATION]->() RETURN count(r) as relation_count", nil)
		if err != nil {
			return nil, err
		}

		var relationCount int
		if relationCountResult.Next(ctx) {
			record := relationCountResult.Record()
			if count, found := record.Get("relation_count"); found {
				relationCount = int(count.(int64))
			}
		}

		// 获取实体类型统计
		entityTypesResult, err := tx.Run(ctx, "MATCH (e:Entity) RETURN e.type as type, count(e) as count", nil)
		if err != nil {
			return nil, err
		}

		var entityTypes []struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
		}
		for entityTypesResult.Next(ctx) {
			record := entityTypesResult.Record()
			entityType, _ := record.Get("type")
			count, _ := record.Get("count")

			entityTypes = append(entityTypes, struct {
				Type  string `json:"type"`
				Count int    `json:"count"`
			}{
				Type:  entityType.(string),
				Count: int(count.(int64)),
			})
		}

		// 获取关系类型统计
		relationTypesResult, err := tx.Run(ctx, "MATCH ()-[r:RELATION]->() RETURN r.type as type, count(r) as count", nil)
		if err != nil {
			return nil, err
		}

		var relationTypes []struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
		}
		for relationTypesResult.Next(ctx) {
			record := relationTypesResult.Record()
			relationType, _ := record.Get("type")
			count, _ := record.Get("count")

			relationTypes = append(relationTypes, struct {
				Type  string `json:"type"`
				Count int    `json:"count"`
			}{
				Type:  relationType.(string),
				Count: int(count.(int64)),
			})
		}

		// 计算图密度
		density := 0.0
		if entityCount > 1 {
			maxPossibleEdges := entityCount * (entityCount - 1)
			density = float64(relationCount*2) / float64(maxPossibleEdges) // 无向图
		}

		return &models.GraphStats{
			TotalEntities:       entityCount,
			TotalRelations:      relationCount,
			EntityTypes:         entityTypes,
			RelationTypes:       relationTypes,
			Density:             density,
			ConnectedComponents: 1, // 简化处理
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.GraphStats), nil
}

// 辅助函数
func getStringProp(props map[string]any, key string) string {
	if val, ok := props[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatProp(props map[string]any, key string) float64 {
	if val, ok := props[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0.0
}

func getMapProp(props map[string]any, key string) map[string]interface{} {
	if val, ok := props[key]; ok && val != nil {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}

func getTimeProp(props map[string]any, key string) time.Time {
	if val, ok := props[key]; ok && val != nil {
		if t, ok := val.(time.Time); ok {
			return t
		}
		// 尝试解析字符串时间
		if str, ok := val.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, str); err == nil {
				return parsed
			}
		}
	}
	return time.Time{}
}
