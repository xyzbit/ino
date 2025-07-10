package indexer

const (
	toolNameStoreEntityAndRelation = "store_entity_and_relation"
)

type StorePairs []StorePair

type StorePair struct {
	From       Entity   `json:"from"`
	To         Entity   `json:"to"`
	Relation   Relation `json:"relation"`
	Confidence float64  `json:"confidence"`
}

type Entity struct {
	Type       string                 `json:"type"` // person, organization, concept, etc.
	Labels     []string               `json:"labels"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

type Relation struct {
	Type       string                 `json:"type"` // 关系类型 WORKS_AT, KNOWS, ...etc.
	Properties map[string]interface{} `json:"properties"`
}
