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
func registerDynamoDBSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the DynamoDB table "([^"]*)" should exist$`, newDynamoDBTableExistsStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have tags$`, newDynamoDBTagsStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have billing mode "([^"]*)"$`, newDynamoDBBillingModeStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have read capacity (\d+)$`, newDynamoDBReadCapacityStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have write capacity (\d+)$`, newDynamoDBWriteCapacityStep)
}

func newDynamoDBTableExistsStep(ctx context.Context, tableName string) error {
	dynamoAssert, err := getDynamoDBAsserter(ctx)
	if err != nil {
		return err
	}

	return dynamoAssert.AssertTableExists(tableName)
}

func newDynamoDBTagsStep(ctx context.Context, tableName string, table *godog.Table) error {
	dynamoAssert, err := getDynamoDBAsserter(ctx)
	if err != nil {
		return err
	}

	// convert the tags to map[string]string
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}

	return dynamoAssert.AssertTableTags(tableName, tags)
}

func newDynamoDBBillingModeStep(ctx context.Context, tableName, expectedMode string) error {
	dynamoAssert, err := getDynamoDBAsserter(ctx)
	if err != nil {
		return err
	}
	return dynamoAssert.AssertBillingMode(tableName, expectedMode)
}

func newDynamoDBReadCapacityStep(ctx context.Context, tableName string, capacity int64) error {
	dynamoAssert, err := getDynamoDBAsserter(ctx)
	if err != nil {
		return err
	}

	// We only check read capacity here
	return dynamoAssert.AssertCapacity(tableName, capacity, -1)
}

func newDynamoDBWriteCapacityStep(ctx context.Context, tableName string, capacity int64) error {
	dynamoAssert, err := getDynamoDBAsserter(ctx)
	if err != nil {
		return err
	}

	// We only check write capacity here
	return dynamoAssert.AssertCapacity(tableName, -1, capacity)
}

func getDynamoDBAsserter(ctx context.Context) (aws.DynamoDBAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	dynamoAssert, ok := asserter.(aws.DynamoDBAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement DynamoDBAsserter")
	}
	return dynamoAssert, nil
}
