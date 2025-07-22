package aws

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Ensure the `AWSAsserter` struct implements the `DynamoDBAsserter` interface.
var _ DynamoDBAsserter = (*AWSAsserter)(nil)

// DynamoDBAsserter defines DynamoDB-specific assertions
type DynamoDBAsserter interface {
	AssertTableExists(tableName string) error
	AssertTableTags(tableName string, expectedTags map[string]string) error
	AssertBillingMode(tableName, expectedMode string) error
	AssertCapacity(tableName string, readCapacity, writeCapacity int64) error
}

// AssertTableExists checks if the DynamoDB table exists.
func (a *AWSAsserter) AssertTableExists(tableName string) error {
	client, err := a.createDynamoDBClient()
	if err != nil {
		return err
	}

	// List tables
	input := &dynamodb.ListTablesInput{}
	result, err := client.ListTables(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error listing tables: %w", err)
	}

	// Check if the table exists
	if slices.Contains(result.TableNames, tableName) {
		return nil
	}

	return fmt.Errorf("table %s does not exist", tableName)
}

// AssertTableTags checks if the DynamoDB table has the expected tags.
func (a *AWSAsserter) AssertTableTags(tableName string, expectedTags map[string]string) error {
	client, err := a.createDynamoDBClient()
	if err != nil {
		return err
	}

	// First, get the table ARN
	table, err := a.getDynamoDBTable(tableName)
	if err != nil {
		return err
	}

	// List tags for the table
	input := &dynamodb.ListTagsOfResourceInput{
		ResourceArn: table.TableArn,
	}

	result, err := client.ListTagsOfResource(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error listing tags for table %s: %w", tableName, err)
	}

	// Convert the tags to a map
	actualTags := make(map[string]string)
	for _, tag := range result.Tags {
		actualTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Compare the expected and actual tags
	for key, value := range expectedTags {
		actualValue, exists := actualTags[key]
		if !exists {
			return fmt.Errorf("expected tag %s not found", key)
		}
		if actualValue != value {
			return fmt.Errorf("expected tag %s to have value %s, but got %s", key, value, actualValue)
		}
	}

	return nil
}

// AssertBillingMode checks if the DynamoDB table has the expected billing mode.
func (a *AWSAsserter) AssertBillingMode(tableName, expectedMode string) error {
	table, err := a.getDynamoDBTable(tableName)
	if err != nil {
		return err
	}

	billingMode, err := getDynamoDBBTableBillingMode(table)
	if err != nil {
		return err
	}

	// check the expected billing mode
	if billingMode != types.BillingMode(expectedMode) {
		return fmt.Errorf("expected billing mode %s, but got %s", expectedMode, billingMode)
	}

	return nil
}

// AssertCapacity checks if the DynamoDB table has the expected read and write capacity.
func (a *AWSAsserter) AssertCapacity(tableName string, readCapacity, writeCapacity int64) error {
	table, err := a.getDynamoDBTable(tableName)
	if err != nil {
		return err
	}

	billingMode, err := getDynamoDBBTableBillingMode(table)
	if err != nil {
		return err
	}

	// if the billing mode is not set to PROVISIONED, we can't check the capacity, so return an error
	if billingMode != types.BillingModeProvisioned {
		return fmt.Errorf("billing mode is PAY_PER_REQUEST, cannot check capacity")
	}

	// if an expected read capacity is provided, check that it matches the table's read capacity
	if readCapacity != -1 && *table.ProvisionedThroughput.ReadCapacityUnits != readCapacity {
		return fmt.Errorf("expected read capacity %d, but got %d", readCapacity, table.ProvisionedThroughput.ReadCapacityUnits)
	}

	// if an expected write capacity is provided, check that it matches the table's write capacity
	if writeCapacity != -1 && *table.ProvisionedThroughput.WriteCapacityUnits != writeCapacity {
		return fmt.Errorf("expected write capacity %d, but got %d", writeCapacity, table.ProvisionedThroughput.WriteCapacityUnits)
	}

	return nil
}

// Helper method to get a DynamoDB table
func (a *AWSAsserter) getDynamoDBTable(tableName string) (*types.TableDescription, error) {
	client, err := a.createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	// Describe the table
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}

	result, err := client.DescribeTable(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing DynamoDB table %s: %w", tableName, err)
	}

	return result.Table, nil
}

// Helper method to create a DynamoDB client
func (a *AWSAsserter) createDynamoDBClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

// Helper method to get the billing mode of a DynamoDB table
func getDynamoDBBTableBillingMode(tableDesc *types.TableDescription) (types.BillingMode, error) {
	if tableDesc == nil {
		return "", fmt.Errorf("table description is nil")
	}

	// Note: as per https://github.com/aws/aws-sdk-go-v2/blob/main/service/dynamodb/types/types.go#L600-L601
	// if we don't receive a response that includes the BillingModeSummary, it means the table is in PROVISIONED mode.
	billingMode := types.BillingModeProvisioned
	if tableDesc.BillingModeSummary != nil {
		billingMode = tableDesc.BillingModeSummary.BillingMode
	}

	return billingMode, nil
}
