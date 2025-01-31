package aws

import (
	"fmt"
	"reflect"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/testing"
)

var _ DynamoDBAsserter = &AWSAsserter{}

// DynamoDBAsserter defines DynamoDB-specific assertions
type DynamoDBAsserter interface {
	AssertTableTags(t testing.TestingT, tableName string, expectedTags map[string]string) error
	AssertBillingMode(t testing.TestingT, tableName, expectedMode string) error
	AssertCapacity(t testing.TestingT, tableName string, readCapacity, writeCapacity int64) error
	AssertGSI(t testing.TestingT, tableName, indexName, keyAttribute string) error
}

func (a *AWSAsserter) AssertTableTags(t testing.TestingT, tableName string, expectedTags map[string]string) error {
	actualTags, err := aws.GetDynamoDbTableTagsE(t, a.region, tableName)
	if err != nil {
		return err
	}

	// convert the terratest tags to a map[string]string
	actualTagsMap := make(map[string]string)
	for _, tag := range actualTags {
		actualTagsMap[*tag.Key] = *tag.Value
	}

	// Compare expected and actual tags
	if !reflect.DeepEqual(expectedTags, actualTagsMap) {
		return fmt.Errorf("expected tags %v, but got %v", expectedTags, actualTagsMap)
	}
	return nil
}

func (a *AWSAsserter) AssertBillingMode(t testing.TestingT, tableName, expectedMode string) error {
	// Implement DynamoDB-specific logic to check billing mode
	return nil
}

func (a *AWSAsserter) AssertCapacity(t testing.TestingT, tableName string, readCapacity, writeCapacity int64) error {
	// Implement DynamoDB-specific logic to check capacity
	return nil
}

func (a *AWSAsserter) AssertGSI(t testing.TestingT, tableName, indexName, keyAttribute string) error {
	// Implement DynamoDB-specific logic to check GSI
	return nil
}
