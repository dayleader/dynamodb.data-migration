package dynamodb

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"dynamodb.data-migration/internal/domain"
	"dynamodb.data-migration/internal/helpers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsDynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	awsErrorResourceNotFound = "ResourceNotFoundException"
	awsErrorResourceInUse    = "ResourceInUseException"
)

const (
	fieldVersion       = "version"
	fieldName          = "name"
	fieldMetadata      = "metadata"
	fieldStartTime     = "start_time"
	fieldExecutionTime = "execution_time"
)

type migrationRepo struct {
	db              *awsDynamodb.DynamoDB
	migrationsTable string
}

// NewMigrationRepository creates a new repository.
func NewMigrationRepository(session *awsSession.Session, migrationsTable string) domain.MigrationRepository {
	r := &migrationRepo{
		db:              awsDynamodb.New(session),
		migrationsTable: migrationsTable,
	}
	if err := r.ensureMigrationsTableExist(); err != nil {
		panic(err)
	}
	return r
}

func (r *migrationRepo) ensureMigrationsTableExist() error {

	// Check table name.
	if len(r.migrationsTable) == 0 {
		return errors.New("Migrations table name required")
	}

	// Check if table exist.
	isExist, err := r.isTableExist(r.migrationsTable)
	if err != nil {
		return err
	}
	if isExist {
		return nil // return, table already exist.
	}

	// Create table.
	_, err = r.db.CreateTable(&awsDynamodb.CreateTableInput{
		AttributeDefinitions: []*awsDynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(fieldVersion),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*awsDynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(fieldVersion),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &awsDynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(r.migrationsTable),
	})
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() != awsErrorResourceInUse {
			return aerr
		}
	}
	// Wait for table.
	return r.db.WaitUntilTableExists(&awsDynamodb.DescribeTableInput{
		TableName: aws.String(r.migrationsTable),
	})
}

func (r *migrationRepo) isTableExist(tableName string) (bool, error) {

	// Check table name.
	if len(tableName) == 0 {
		return false, errors.New("Table name required")
	}

	// Check if the table exist.
	describeTableOutput, err := r.db.DescribeTable(&awsDynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() != awsErrorResourceNotFound {
			return false, errors.New(aerr.Error())
		}
	} else if err != nil {
		// Returned internal server error.
		// The application does not know if the table exists or not.
		// Thus, it cannot query the server, so we panic this error.
		return false, err
	}
	return describeTableOutput.Table != nil, nil
}

func (r *migrationRepo) IsMigrationRecordExist(ver domain.Version) (bool, error) {

	// Build the query input parameters.
	getInput := &awsDynamodb.GetItemInput{
		TableName: aws.String(r.migrationsTable),
		Key: map[string]*awsDynamodb.AttributeValue{
			fieldVersion: {
				S: aws.String(ver.ID()),
			},
		},
	}

	// Make the DynamoDB Query API call.
	result, err := r.db.GetItem(getInput)
	if err != nil {
		return false, fmt.Errorf("Query API call failed: %s", err)
	}

	// Return result.
	return len(result.Item) > 0, nil
}

func (r *migrationRepo) CreateMigrationRecord(migrationRecord domain.MigrationRecord) error {

	// Marshal Go value type to a map of AttributeValues.
	items := map[string]*awsDynamodb.AttributeValue{
		fieldVersion: {S: aws.String(migrationRecord.Version.ID())},
		fieldName:    {S: aws.String(migrationRecord.Name)},
		fieldMetadata: {M: map[string]*awsDynamodb.AttributeValue{
			fieldStartTime:     {N: aws.String(strconv.Itoa(int(migrationRecord.Metadata.StartTime)))},
			fieldExecutionTime: {N: aws.String(strconv.Itoa(int(migrationRecord.Metadata.ExecutionTime)))},
		}},
	}

	transaction := &awsDynamodb.TransactWriteItemsInput{
		TransactItems: []*awsDynamodb.TransactWriteItem{
			{
				Put: &awsDynamodb.Put{
					TableName:           aws.String(r.migrationsTable),
					ConditionExpression: aws.String("attribute_not_exists(#pk)"),
					ExpressionAttributeNames: map[string]*string{
						"#pk": aws.String(fieldVersion),
					},
					Item: items,
				},
			},
		},
	}

	// Run transaction.
	req, _ := r.db.TransactWriteItemsRequest(transaction)
	if err := req.Send(); err != nil {
		return err
	}
	return nil
}

func (r *migrationRepo) ExecuteQueries(queries []*domain.DynamoDBQuery) error {
	var (
		createTableInputs = make([]*awsDynamodb.CreateTableInput, 0)
		dataTransactions  = make([]*awsDynamodb.TransactWriteItem, 0)
	)
	for _, q := range queries {
		if err := q.Validate(); err != nil {
			return err
		}
		for _, schema := range q.Schema {
			createTableInputs = append(createTableInputs, &awsDynamodb.CreateTableInput{
				AttributeDefinitions: helpers.ConvertToAWSAttributeDefinitions(schema.AttributeDefinitions),
				KeySchema:            helpers.ConvertToAWSKeySchemaElement(schema.KeySchema),
				ProvisionedThroughput: &awsDynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
				TableName: aws.String(q.TableName),
			})
		}
		for _, data := range q.Data {
			// Marshal Go value type to a map of AttributeValues.
			item, err := dynamodbattribute.MarshalMap(data)
			if err != nil {
				return err
			}
			if len(item) == 0 {
				return fmt.Errorf("Items cannot be empty for %v", q.TableName)
			}
			dataTransactions = append(dataTransactions, &awsDynamodb.TransactWriteItem{
				Put: &awsDynamodb.Put{
					TableName: aws.String(q.TableName),
					Item:      item,
				},
			})
		}
	}

	// Create tables.
	for _, createTableInput := range createTableInputs {
		isTableExist, err := r.isTableExist(*createTableInput.TableName)
		if err != nil {
			return err
		}
		if isTableExist {
			log.Printf("Skipping a table %s because the table already exist\n", *createTableInput.TableName)
			continue
		}
		_, err = r.db.CreateTable(createTableInput)
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() != awsErrorResourceInUse {
				return aerr
			}
		}
		// Wait for table.
		if err := r.db.WaitUntilTableExists(&awsDynamodb.DescribeTableInput{TableName: createTableInput.TableName}); err != nil {
			return err
		}
	}

	// Run data migrations if present.
	if len(dataTransactions) > 0 {
		req, _ := r.db.TransactWriteItemsRequest(&awsDynamodb.TransactWriteItemsInput{
			TransactItems: dataTransactions,
		})
		if err := req.Send(); err != nil {
			return err
		}
	}
	return nil
}
