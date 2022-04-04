package helpers

import (
	"dynamodb.data-migration/internal/domain"
	"github.com/aws/aws-sdk-go/aws"
	awsDynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
)

// ConvertToAWSAttributeDefinitions - converts a domain struct to aws struct.
func ConvertToAWSAttributeDefinitions(attrs []*domain.DynamoDBAttributeDefinition) []*awsDynamodb.AttributeDefinition {
	result := make([]*awsDynamodb.AttributeDefinition, len(attrs))
	for i, a := range attrs {
		result[i] = &awsDynamodb.AttributeDefinition{
			AttributeName: aws.String(a.AttributeName),
			AttributeType: aws.String(a.AttributeType),
		}
	}
	return result
}

// ConvertToAWSKeySchemaElement - converts a domain struct to aws struct.
func ConvertToAWSKeySchemaElement(schema []*domain.DynamoDBKeySchema) []*awsDynamodb.KeySchemaElement {
	result := make([]*awsDynamodb.KeySchemaElement, len(schema))
	for i, s := range schema {
		result[i] = &awsDynamodb.KeySchemaElement{
			AttributeName: aws.String(s.AttributeName),
			KeyType:       aws.String(s.KeyType),
		}
	}
	return result
}
