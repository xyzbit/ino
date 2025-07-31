package indexer

const (
	toolNameStoreEntityAndRelation = "store_entity_and_relation"
)

type ExtractEntityAndRelation struct {
	Entities  []Entity   `json:"entities" jsonschema:"required,description=the entities, example: [{type: person, name: John Doe, properties: {age: 30, gender: male}}, {type: organization, name: Google, properties: {industry: technology, location: California}}]"`
	Relations []Relation `json:"relations" jsonschema:"required,description=the relations between the entities, example: [{from_entity_name: John Doe, to_entity_name: Google, type: WORKS_AT, properties: {start_date: 2020-01-01}}]"`
}

type Entity struct {
	Type       string                 `json:"type" jsonschema:"required,description=the type of the entity, such as person, organization, concept,... etc."`
	Name       string                 `json:"name" jsonschema:"required,description=the name of the entity, must be a specific and clear name, such as '张三', 'k8s', '消息中心', ... etc."`
	Properties map[string]interface{} `json:"properties,omitempty" jsonschema:"description=the properties of the entity, extract only necessary information, no more than 3 properties, example: if the entity is a person, the properties can be age, gender, ... etc."`
}

type Relation struct {
	FromEntityName string                 `json:"from_entity_name" jsonschema:"required,description=the from entity name"`
	ToEntityName   string                 `json:"to_entity_name" jsonschema:"required,description=the to entity name"`
	Type           string                 `json:"type" jsonschema:"required,description=the type of the relation, such as WORKS_AT, KNOWS, ...etc."`
	Properties     map[string]interface{} `json:"properties,omitempty" jsonschema:"description=the properties of the relation, extract only necessary information, no more than 5 properties, example: if the relation is WORKS_AT, the properties can be start_date, ... etc."`
}

type StorePairs []StorePair

type StorePair struct {
	From     Entity   `json:"from"`
	To       Entity   `json:"to"`
	Relation Relation `json:"relation"`
}
