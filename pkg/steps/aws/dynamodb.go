package aws

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	t "github.com/robmorgan/infraspec/internal/testing"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// DynamoDB Step Definitions
func newDynamoDBTagsStep(ctx context.Context, tableName string, table *godog.Table) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	// convert the tags to map[string]string
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}

	return dynamoAssert.AssertTableTags(t.GetT(), tableName, tags)
}

func newDynamoDBBillingModeStep(ctx context.Context, tableName, expectedMode string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	return dynamoAssert.AssertBillingMode(t.GetT(), tableName, expectedMode)
}

func newDynamoDBReadCapacityStep(ctx context.Context, tableName string, capacity int64) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	// We only check read capacity here
	return dynamoAssert.AssertCapacity(t.GetT(), tableName, capacity, -1)
}

func newDynamoDBWriteCapacityStep(ctx context.Context, tableName string, capacity int64) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	// We only check write capacity here
	return dynamoAssert.AssertCapacity(t.GetT(), tableName, -1, capacity)
}
