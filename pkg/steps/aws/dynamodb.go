package aws

import (
	"fmt"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

// DynamoDB Step Definitions
func newDynamoDBTagsStep(ctx *context.TestContext) func(string, *godog.Table) error {
	return func(tableName string, tags *godog.Table) error {
		// TODO - implement
		return nil
	}
}

func newDynamoDBBillingModeStep(ctx *context.TestContext) func(string, string) error {
	return func(tableName, expectedMode string) error {
		// Replace any variables in the table name
		tableName = replaceVariables(ctx, tableName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		dynamoAssert, ok := asserter.(assertions.DynamoDBAsserter)
		if !ok {
			return fmt.Errorf("asserter does not implement DynamoDBAsserter")
		}

		return dynamoAssert.AssertBillingMode(tableName, expectedMode)
	}
}

func newDynamoDBReadCapacityStep(ctx *context.TestContext) func(string, int64) error {
	return func(tableName string, capacity int64) error {
		tableName = replaceVariables(ctx, tableName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		dynamoAssert, ok := asserter.(assertions.DynamoDBAsserter)
		if !ok {
			return fmt.Errorf("asserter does not implement DynamoDBAsserter")
		}

		// We only check read capacity here
		return dynamoAssert.AssertCapacity(tableName, capacity, -1)
	}
}

func newDynamoDBWriteCapacityStep(ctx *context.TestContext) func(string, int64) error {
	return func(tableName string, capacity int64) error {
		tableName = replaceVariables(ctx, tableName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		dynamoAssert, ok := asserter.(assertions.DynamoDBAsserter)
		if !ok {
			return fmt.Errorf("asserter does not implement DynamoDBAsserter")
		}

		// We only check write capacity here
		return dynamoAssert.AssertCapacity(tableName, -1, capacity)
	}
}

func newDynamoDBGSIStep(ctx *context.TestContext) func(string, string, string) error {
	return func(tableName, indexName, keyAttribute string) error {
		tableName = replaceVariables(ctx, tableName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		dynamoAssert, ok := asserter.(assertions.DynamoDBAsserter)
		if !ok {
			return fmt.Errorf("asserter does not implement DynamoDBAsserter")
		}

		return dynamoAssert.AssertGSI(tableName, indexName, keyAttribute)
	}
}
