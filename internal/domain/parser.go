package domain

// QueryParser - query parser.
type QueryParser interface {

	// ParseQuery - parses content query into dynamodb query.
	ParseContent(content []byte) ([]*DynamoDBQuery, error)
}
