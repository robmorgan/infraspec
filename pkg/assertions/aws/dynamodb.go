package aws

var _ DynamoDBAsserter = &AWSAsserter{}

// DynamoDBAsserter defines DynamoDB-specific assertions
type DynamoDBAsserter interface {
	AssertTableTags(tableName string, expectedTags map[string]string) error
	AssertBillingMode(tableName, expectedMode string) error
	AssertCapacity(tableName string, readCapacity, writeCapacity int64) error
	AssertGSI(tableName, indexName, keyAttribute string) error
}

func (a *AWSAsserter) AssertTableTags(tableName string, expectedTags map[string]string) error {
	// Implement DynamoDB-specific logic to check table tags
	return nil
}

func (a *AWSAsserter) AssertBillingMode(tableName, expectedMode string) error {
	// Implement DynamoDB-specific logic to check billing mode
	return nil
}

func (a *AWSAsserter) AssertCapacity(tableName string, readCapacity, writeCapacity int64) error {
	// Implement DynamoDB-specific logic to check capacity
	return nil
}

func (a *AWSAsserter) AssertGSI(tableName, indexName, keyAttribute string) error {
	// Implement DynamoDB-specific logic to check GSI
	return nil
}
