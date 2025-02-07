package aws

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/testing"
)

var _ DynamoDBAsserter = &AWSAsserter{}

// DynamoDBAsserter defines DynamoDB-specific assertions
type DynamoDBAsserter interface {
	AssertTableTags(t testing.TestingT, tableName string, expectedTags map[string]string) error
	AssertBillingMode(t testing.TestingT, tableName, expectedMode string) error
	AssertCapacity(t testing.TestingT, tableName string, readCapacity, writeCapacity int64) error
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
	table, err := aws.GetDynamoDBTableE(t, a.region, tableName)
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

func (a *AWSAsserter) AssertCapacity(t testing.TestingT, tableName string, readCapacity, writeCapacity int64) error {
	table, err := aws.GetDynamoDBTableE(t, a.region, tableName)
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

func getDynamoDBBTableBillingMode(tableDesc *types.TableDescription) (types.BillingMode, error) {
	// Note: as per https://github.com/aws/aws-sdk-go-v2/blob/main/service/dynamodb/types/types.go#L600-L601
	// if we don't receive a response that includes the BillingModeSummary, it means the table is in PROVISIONED mode.
	billingMode := types.BillingModeProvisioned
	if tableDesc.BillingModeSummary != nil {
		billingMode = tableDesc.BillingModeSummary.BillingMode
	}

	return billingMode, nil
}
