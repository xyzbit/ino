package indexer

type StorePairs []StorePair

type StorePair struct {
	From     Entity   `json:"from"`
	To       Entity   `json:"to"`
	Relation Relation `json:"relation"`
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
