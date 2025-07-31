package jobs

// TODO: 处理数据冲突
/*
知识数据可能存在过时、冲突、冗余等问题，需要监听冲突事件，进行异步处理。
这样设计的原因如下：
其实实时在写入时解决冲突，也是一个不错的方案，但是冲突解决的效果由大模型决定，不稳定，如果某次冲突解决不彻底，可能导致冲突数据一直留存在数据库中。
但如果在查询时实时解决，会严重影响查询性能，所以我们在查询时已有的 rerank 阶段进行判断，如果存在冲突，发送冲突事件，由消费者进行异步处理（同时 rerank 阶段会将冲突数据中较新的数据返回，一次保证查询效果）。
*/

// func (i *Indexer) handleConflictRelations(ctx context.Context, newPairs StorePairs, opts *ImplOptions) error {
// 	database := opts.Database
// 	if database == "" {
// 		database = i.config.Database
// 	}
// 	session := i.config.Driver.NewSession(ctx, neo4j.SessionConfig{
// 		AccessMode:   neo4j.AccessModeWrite,
// 		DatabaseName: database,
// 	})
// 	defer session.Close(ctx)

// 	records, err := i.searchSimilarGraph(ctx, session, newPairs, opts.SimilarityThreshold)
// 	if err != nil {
// 		return fmt.Errorf("failed to search similar graph: %w", err)
// 	}

// 	oldRelations := make([]StorePair, 0)
// 	for _, record := range records {
// 		pair, err := i.recordToStorePair(record)
// 		if err != nil {
// 			return fmt.Errorf("failed to convert record to store pair: %w", err)
// 		}
// 		oldRelations = append(oldRelations, *pair)
// 	}

// 	// 大模型判断是否需要删除
// 	msgs, err := PromptGraphExtractEntityAndRelation.Format(ctx, map[string]any{
// 		"new_relations": newRelations,
// 		"old_relations": oldRelations,
// 		"resp_schema":   i.config.respSchema,
// 	})
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// }

// searchSimilarGraph 搜索相似的图谱
// func (i *Indexer) searchSimilarGraph(ctx context.Context, session neo4j.SessionWithContext, pairs StorePairs, threshold float64) ([]*neo4j.Record, error) {
// 	filter := map[string]string{}
// 	embeddings := make([][]float64, 0)
// 	for _, pair := range pairs {
// 		if pair.From.Properties[constants.CollectionKey] != "" {
// 			filter[constants.CollectionKey] = pair.From.Properties[constants.CollectionKey].(string)
// 		}
// 		if pair.To.Properties[constants.CollectionKey] != "" {
// 			filter[constants.CollectionKey] = pair.To.Properties[constants.CollectionKey].(string)
// 		}
// 		ebs, err := i.config.EmbeddingModel.EmbedStrings(ctx, []string{pair.From.Name, pair.To.Name})
// 		if err != nil {
// 			return nil, err
// 		}
// 		embeddings = append(embeddings, ebs...)
// 	}

// 	records := make([]*neo4j.Record, 0, len(embeddings))
// 	for _, embedding := range embeddings {
// 		rs, err := infraNeo4j.GraphSearch(ctx, session, embedding, &infraNeo4j.SearchOptions{
// 			TopK:                1,
// 			Limit:               3,
// 			SimilarityThreshold: threshold,
// 			MaxDepth:            1,
// 			Filter:              filter,
// 		})
// 		if err != nil {
// 			log.Println("failed to search graph", err)
// 			continue
// 		}
// 		records = append(records, rs...)
// 	}

// 	return records, nil
// }

// func (i *Indexer) deleteGraph(ctx context.Context, session neo4j.SessionWithContext, pair *StorePair) error {
// 	params := map[string]interface{}{
// 		"source_name": pair.From.Name,
// 		"dest_name":   pair.To.Name,
// 	}
// 	relationshipType := sanitizeRelationshipType(pair.Relation.Type)
// 	fromType := sanitizeEntityType(pair.From.Type)
// 	toType := sanitizeEntityType(pair.To.Type)

// 	var (
// 		fromProperties string
// 		toProperties   string
// 	)
// 	for k, v := range pair.From.Properties {
// 		fromProperties += fmt.Sprintf(", %s: %v", k, v)
// 	}
// 	for k, v := range pair.To.Properties {
// 		toProperties += fmt.Sprintf(", %s: %v", k, v)
// 	}

// 	cypher := fmt.Sprintf(`
// 		MATCH (n:%s:Entity {name: $source_name%s})
// 		-[r:%s]->
// 		(m:%s:Entity {name: $dest_name%s})
// 	DELETE r
// 	RETURN
// 		n.name AS source,
// 		m.name AS target,
// 		type(r) AS relationship
// 	`, fromType, fromProperties, relationshipType, toType, toProperties)

// 	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
// 		result, err := tx.Run(ctx, cypher, params)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return result.Consume(ctx)
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to delete graph: %w", err)
// 	}

// 	return nil
// }
