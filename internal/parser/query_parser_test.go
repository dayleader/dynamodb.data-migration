package parser

import (
	"reflect"
	"testing"

	"dynamodb.data-migration/internal/domain"
)

func TestQueryParser(t *testing.T) {
	validSchemaDataQuery := `[
		{
			"table_name": "users",
			"schema": [
				{
					"attribute_definitions": [
						{
							"name": "id",
							"type": "S"
						}
					],
					"key_schema": [
						{
							"name": "id",
							"type": "HASH"
						}
					]
				}
			],
			"data": [
				{
					"id": "1",
					"username": "test1",
					"email": "test1@email.com"
				},
				{
					"id": "2",
					"username": "test2",
					"email": "test2@email.com"
				},
				{
					"id": "3",
					"username": "test3",
					"email": "test3@email.com"
				}
			]
		},
		{
			"table_name": "roles",
			"schema": [
				{
					"attribute_definitions": [
						{
							"name": "id",
							"type": "S"
						}
					],
					"key_schema": [
						{
							"name": "id",
							"type": "HASH"
						}
					]
				}
			],
			"data": [
				{
					"id": "1",
					"name": "role1"
				},
				{
					"id": "2",
					"name": "role2"
				}
			]
		}
	]`
	validSchemaQuery := `[
		{
			"table_name": "contacts",
			"schema": [
				{
					"attribute_definitions": [
						{
							"name": "id",
							"type": "S"
						}
					],
					"key_schema": [
						{
							"name": "id",
							"type": "HASH"
						}
					]
				}
			]
		},
		{
			"table_name": "user_contacts",
			"schema": [
				{
					"attribute_definitions": [
						{
							"name": "user_id",
							"type": "S"
						},
						{
							"name": "contact_id",
							"type": "S"
						}
					],
					"key_schema": [
						{
							"name": "user_id",
							"type": "HASH"
						},
						{
							"name": "contact_id",
							"type": "RANGE"
						}
					]
				}
			]
		}
	]`
	validDataQuery := `[
		{
			"table_name": "groups",
			"data": [
				{
					"name": "admins",
					"position": 1
				},
				{
					"name": "managers",
					"position": 2
				},
				{
					"name": "users",
					"position": 3
				}
			]
		},
		{
			"table_name": "organizations",
			"data": [
				{
					"name": "Test 1",
					"active": true
				},
				{
					"name": "Test 2",
					"active": false
				}
			]
		}
	]`

	// Test.
	type arguments struct {
		query string
	}
	type result struct {
		query []*domain.DynamoDBQuery
	}
	tests := []struct {
		name        string
		arguments   arguments
		expected    result
		expectError bool
	}{
		{
			name: "Success: schema and data",
			arguments: arguments{
				query: validSchemaDataQuery,
			},
			expected: result{
				query: []*domain.DynamoDBQuery{
					{
						TableName: "users",
						Schema: []*domain.DynamoDBSchema{
							{
								AttributeDefinitions: []*domain.DynamoDBAttributeDefinition{
									{
										AttributeName: "id",
										AttributeType: "S",
									},
								},
								KeySchema: []*domain.DynamoDBKeySchema{
									{
										AttributeName: "id",
										KeyType:       "HASH",
									},
								},
							},
						},
						Data: []map[string]interface{}{
							{
								"id":       "1",
								"username": "test1",
								"email":    "test1@email.com",
							},
							{
								"id":       "2",
								"username": "test2",
								"email":    "test2@email.com",
							},
							{
								"id":       "3",
								"username": "test3",
								"email":    "test3@email.com",
							},
						},
					},
					{
						TableName: "roles",
						Schema: []*domain.DynamoDBSchema{
							{
								AttributeDefinitions: []*domain.DynamoDBAttributeDefinition{
									{
										AttributeName: "id",
										AttributeType: "S",
									},
								},
								KeySchema: []*domain.DynamoDBKeySchema{
									{
										AttributeName: "id",
										KeyType:       "HASH",
									},
								},
							},
						},
						Data: []map[string]interface{}{
							{
								"id":   "1",
								"name": "role1",
							},
							{
								"id":   "2",
								"name": "role2",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Success: schema only",
			arguments: arguments{
				query: validSchemaQuery,
			},
			expected: result{
				query: []*domain.DynamoDBQuery{
					{
						TableName: "contacts",
						Schema: []*domain.DynamoDBSchema{
							{
								AttributeDefinitions: []*domain.DynamoDBAttributeDefinition{
									{
										AttributeName: "id",
										AttributeType: "S",
									},
								},
								KeySchema: []*domain.DynamoDBKeySchema{
									{
										AttributeName: "id",
										KeyType:       "HASH",
									},
								},
							},
						},
						Data: []map[string]interface{}{},
					},
					{
						TableName: "user_contacts",
						Schema: []*domain.DynamoDBSchema{
							{
								AttributeDefinitions: []*domain.DynamoDBAttributeDefinition{
									{
										AttributeName: "user_id",
										AttributeType: "S",
									},
									{
										AttributeName: "contact_id",
										AttributeType: "S",
									},
								},
								KeySchema: []*domain.DynamoDBKeySchema{
									{
										AttributeName: "user_id",
										KeyType:       "HASH",
									},
									{
										AttributeName: "contact_id",
										KeyType:       "RANGE",
									},
								},
							},
						},
						Data: []map[string]interface{}{},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Success: data only",
			arguments: arguments{
				query: validDataQuery,
			},
			expected: result{
				query: []*domain.DynamoDBQuery{
					{
						TableName: "groups",
						Schema:    []*domain.DynamoDBSchema{},
						Data: []map[string]interface{}{
							{
								"name":     "admins",
								"position": float64(1),
							},
							{
								"name":     "managers",
								"position": float64(2),
							},
							{
								"name":     "users",
								"position": float64(3),
							},
						},
					},
					{
						TableName: "organizations",
						Schema:    []*domain.DynamoDBSchema{},
						Data: []map[string]interface{}{
							{
								"name":   "Test 1",
								"active": true,
							},
							{
								"name":   "Test 2",
								"active": false,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Fail: empty query content",
			arguments: arguments{
				query: "",
			},
			expected: result{
				query: nil,
			},
			expectError: true,
		},
	}
	queryParser := NewQueryParser()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arguments := test.arguments
			expected := test.expected

			parsedQuery, err := queryParser.ParseContent([]byte(arguments.query))
			if !test.expectError {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				actual := result{
					query: parsedQuery,
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Error("parsed and expected structures are diffrent")
				}
			} else {
				if err == nil {
					t.Error("expected error but got nothing")
				}
			}
		})
	}
}
