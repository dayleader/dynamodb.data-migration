package domain

import "errors"

// Field names.
const (
	JSONFieldTableName = "table_name"
	JSONFieldSchema    = "schema"
	JSONFieldData      = "data"
)

// DynamoDBAttributeDefinition - represents an attribute for describing the key schema for the table and indexes.
type DynamoDBAttributeDefinition struct {
	AttributeName string `json:"name"`
	AttributeType string `json:"type"`
}

// DynamoDBKeySchema - represents a single element of a key schema.
type DynamoDBKeySchema struct {
	AttributeName string `json:"name"`
	KeyType       string `json:"type"`
}

// DynamoDBSchema - represents a dynamodb schema format.
type DynamoDBSchema struct {
	AttributeDefinitions []*DynamoDBAttributeDefinition `json:"attribute_definitions"`
	KeySchema            []*DynamoDBKeySchema           `json:"key_schema"`
}

// DynamoDBQuery - represents a dynamodb query format.
type DynamoDBQuery struct {
	TableName string                   `json:"table_name"`
	Schema    []*DynamoDBSchema        `json:"schema"`
	Data      []map[string]interface{} `json:"data"`
}

// Validate - checks if the dynamodb query is valid.
func (q *DynamoDBQuery) Validate() error {
	if len(q.TableName) == 0 {
		return errors.New("Table name required")
	}
	if len(q.Schema) == 0 && len(q.Data) == 0 {
		return errors.New("Either schema or data must be specified")
	}
	return nil
}
