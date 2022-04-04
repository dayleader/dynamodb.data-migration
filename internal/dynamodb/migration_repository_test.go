package dynamodb

import (
	"fmt"
	"testing"

	"dynamodb.data-migration/internal/domain"
	aws "github.com/aws/aws-sdk-go/aws"
	awsDynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-test/deep"
)

func TestMigrationRecords(t *testing.T) {
	versions := []domain.Version{
		{
			Major: 1,
			Minor: 1,
			Patch: 148,
		},
		{
			Major: 1,
			Minor: 1,
			Patch: 149,
		},
		{
			Major: 1,
			Minor: 1,
			Patch: 150,
		},
	}
	for _, ver := range versions {
		// Migration record should not exist.
		isExist, err := testMigrationRepository.IsMigrationRecordExist(ver)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
		if isExist {
			t.Errorf("migration record should not exist for %v", ver)
		}

		// Create a new migration record.
		err = testMigrationRepository.CreateMigrationRecord(domain.MigrationRecord{
			Version: ver,
			Name:    "test",
			Metadata: domain.Metadata{
				StartTime:     123,
				ExecutionTime: 1,
			},
		})
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}

		// Check if the migration record created.
		isExist, err = testMigrationRepository.IsMigrationRecordExist(ver)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
		if !isExist {
			t.Errorf("migration record must exist for %v", ver)
		}
	}
}

func TestExecuteQueries(t *testing.T) {
	var (
		db    = awsDynamodb.New(testAwsSession)
		users = []map[string]interface{}{
			{
				"id":       "490fa2a9-967a-4488-a85d-c28db110470b",
				"username": "test1",
				"email":    "test1@email.com",
			},
			{
				"id":       "5ab0ca77-d913-4d4d-9d90-5da840c2f0cd",
				"username": "test2",
				"email":    "test2@email.com",
			},
		}
		roles = []map[string]interface{}{
			{
				"id":   "a38bce70-1551-4d8b-bdca-d2224d217790",
				"name": "admin",
			},
			{
				"id":   "3847ae2b-b9a1-42bf-a20a-259073718336",
				"name": "manager",
			},
			{
				"id":   "911c16ec-3214-4e02-bb65-26e10ac2a1e4",
				"name": "standard",
			},
		}
	)

	// Create table schemas with test data.
	err := testMigrationRepository.ExecuteQueries([]*domain.DynamoDBQuery{
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
			Data: users,
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
			Data: roles,
		},
	})
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// Find created users.
	for _, user := range users {
		getInput := &awsDynamodb.GetItemInput{
			TableName: aws.String("users"),
			Key: map[string]*awsDynamodb.AttributeValue{
				"id": {
					S: aws.String(fmt.Sprintf("%v", user["id"])),
				},
			},
		}
		result, err := db.GetItem(getInput)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
			continue
		}
		if len(result.Item) == 0 {
			t.Errorf("item not found by id %v", user["id"])
			continue
		}
		var (
			actualID         = *result.Item["id"].S
			expectedID       = fmt.Sprintf("%v", user["id"])
			actualUsername   = *result.Item["username"].S
			expectedUsername = fmt.Sprintf("%v", user["username"])
			actualEmail      = *result.Item["email"].S
			expectedEmail    = fmt.Sprintf("%v", user["email"])
		)
		if diff := deep.Equal(actualID, expectedID); diff != nil {
			t.Errorf("actual user: %v does not match expected: %v", actualID, expectedID)
		}
		if diff := deep.Equal(actualUsername, expectedUsername); diff != nil {
			t.Errorf("actual user: %v does not match expected: %v", actualUsername, expectedUsername)
		}
		if diff := deep.Equal(actualEmail, expectedEmail); diff != nil {
			t.Errorf("actual user: %v does not match expected: %v", actualEmail, expectedEmail)
		}
	}

	// Find created roles.
	for _, role := range roles {
		getInput := &awsDynamodb.GetItemInput{
			TableName: aws.String("roles"),
			Key: map[string]*awsDynamodb.AttributeValue{
				"id": {
					S: aws.String(fmt.Sprintf("%v", role["id"])),
				},
			},
		}
		result, err := db.GetItem(getInput)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
			continue
		}
		if len(result.Item) == 0 {
			t.Errorf("item not found by id %v", role["id"])
			continue
		}
		var (
			actualID     = *result.Item["id"].S
			expectedID   = fmt.Sprintf("%v", role["id"])
			actualName   = *result.Item["name"].S
			expectedName = fmt.Sprintf("%v", role["name"])
		)
		if diff := deep.Equal(actualID, expectedID); diff != nil {
			t.Errorf("actual user: %v does not match expected: %v", actualID, expectedID)
		}
		if diff := deep.Equal(actualName, expectedName); diff != nil {
			t.Errorf("actual user: %v does not match expected: %v", actualName, expectedName)
		}
	}
}
