package aws

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// DynamoDB Step Definitions
func newDynamoDBTagsStep(ctx context.Context, tableName string, tags *godog.Table) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	return dynamoAssert.AssertTableTags(tableName, tags)
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

	return dynamoAssert.AssertBillingMode(tableName, expectedMode)
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
	return dynamoAssert.AssertCapacity(tableName, capacity, -1)
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
	return dynamoAssert.AssertCapacity(tableName, -1, capacity)
}

func newDynamoDBGSIStep(ctx context.Context, tableName, indexName, keyAttribute string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}

	return dynamoAssert.AssertGSI(tableName, indexName, keyAttribute)
}
