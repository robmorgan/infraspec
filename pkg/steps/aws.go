package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

// registerAWSSteps registers all AWS-specific step definitions
func registerAWSSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	// DynamoDB steps
	sc.Step(`^the DynamoDB table "([^"]*)" should have billing mode "([^"]*)"$`, newDynamoDBBillingModeStep(ctx))
	sc.Step(`^the DynamoDB table "([^"]*)" should have read capacity (\d+)$`, newDynamoDBReadCapacityStep(ctx))
	sc.Step(`^the DynamoDB table "([^"]*)" should have write capacity (\d+)$`, newDynamoDBWriteCapacityStep(ctx))
	//sc.Step(`^the DynamoDB table "([^"]*)" should have GSI "([^"]*)" with key "([^"]*)"$`, newDynamoDBGSIStep(ctx))
	//sc.Step(`^the DynamoDB table "([^"]*)" should have point in time recovery enabled$`, newDynamoDBPITRStep(ctx))

	// S3 steps
	// sc.Step(`^the S3 bucket "([^"]*)" should exist$`, newS3BucketExistsStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have versioning enabled$`, newS3BucketVersioningStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have encryption "([^"]*)"$`, newS3BucketEncryptionStep(ctx))

	// Generic AWS steps
	sc.Step(`^the resource "([^"]*)" should have tags$`, newAWSTagsStep(ctx))
	// sc.Step(`^I wait for resource "([^"]*)" to be "([^"]*)"$`, newAWSWaitForStateStep(ctx))
}

// DynamoDB Step Definitions
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

// S3 Step Definitions
func newS3BucketExistsStep(ctx *context.TestContext) func(string) error {
	return func(bucketName string) error {
		bucketName = replaceVariables(ctx, bucketName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		s3Assert, ok := asserter.(assertions.S3Asserter)
		if !ok {
			return fmt.Errorf("asserter does not implement S3Asserter")
		}

		return s3Assert.AssertBucketExists(bucketName)
	}
}

func newS3BucketVersioningStep(ctx *context.TestContext) func(string) error {
	return func(bucketName string) error {
		bucketName = replaceVariables(ctx, bucketName)

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		s3Assert, ok := asserter.(assertions.S3Asserter)
		if !ok {
			return fmt.Errorf("asserter does not implement S3Asserter")
		}

		return s3Assert.AssertBucketVersioning(bucketName, true)
	}
}

// Generic AWS Steps
func newAWSTagsStep(ctx *context.TestContext) func(string, *godog.Table) error {
	return func(resourceID string, table *godog.Table) error {
		resourceID = replaceVariables(ctx, resourceID)

		tags := make(map[string]string)
		for _, row := range table.Rows[1:] { // Skip header row
			tags[row.Cells[0].Value] = row.Cells[1].Value
		}

		asserter, err := ctx.GetAsserter("aws")
		if err != nil {
			return err
		}

		return asserter.AssertTags("", resourceID, tags)
	}
}

// Helper Functions
func replaceVariables(ctx *context.TestContext, input string) string {
	if !strings.Contains(input, "${") {
		return input
	}

	result := input
	for key, value := range ctx.GetStoredValues() {
		result = strings.ReplaceAll(result, "${"+key+"}", value)
	}
	return result
}
