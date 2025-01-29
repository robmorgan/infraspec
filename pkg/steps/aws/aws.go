package aws

import (
	"strings"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/context"
)

// RegisterSteps registers all AWS-specific step definitions
func RegisterSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	// DynamoDB steps
	sc.Step(`^the DynamoDB table "([^"]*)" should have tags$`, newDynamoDBTagsStep(ctx))
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
	sc.Step(`^the AWS resource "([^"]*)" should exist$`, newAWSResourceExistsStep(ctx))
	sc.Step(`^the resource "([^"]*)" should have tags$`, newAWSTagsStep(ctx))
	// sc.Step(`^I wait for resource "([^"]*)" to be "([^"]*)"$`, newAWSWaitForStateStep(ctx))
}

// Generic AWS Steps
func newAWSResourceExistsStep(ctx *context.TestContext) func(string) error {
	return func(resourceID string) error {
		// TODO - implement
		return nil
	}
}

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
