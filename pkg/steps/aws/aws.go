package aws

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
)

// RegisterSteps registers all AWS-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	// DynamoDB steps
	sc.Step(`^the DynamoDB table "([^"]*)" should have tags$`, newDynamoDBTagsStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have billing mode "([^"]*)"$`, newDynamoDBBillingModeStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have read capacity (\d+)$`, newDynamoDBReadCapacityStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have write capacity (\d+)$`, newDynamoDBWriteCapacityStep)
	//sc.Step(`^the DynamoDB table "([^"]*)" should have GSI "([^"]*)" with key "([^"]*)"$`, newDynamoDBGSIStep(ctx))
	//sc.Step(`^the DynamoDB table "([^"]*)" should have point in time recovery enabled$`, newDynamoDBPITRStep(ctx))

	// S3 steps
	// sc.Step(`^the S3 bucket "([^"]*)" should exist$`, newS3BucketExistsStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have versioning enabled$`, newS3BucketVersioningStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have encryption "([^"]*)"$`, newS3BucketEncryptionStep(ctx))

	// Generic AWS steps
	sc.Step(`^the AWS resource "([^"]*)" should exist$`, newAWSResourceExistsStep)
	sc.Step(`^the resource "([^"]*)" should have tags$`, newAWSTagsStep)
	// sc.Step(`^I wait for resource "([^"]*)" to be "([^"]*)"$`, newAWSWaitForStateStep(ctx))
}

// Generic AWS Steps
func newAWSResourceExistsStep(ctx context.Context, resourceID string) error {
	// TODO - implement
	return nil
}

func newAWSTagsStep(ctx context.Context, table *godog.Table) error {
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}

	asserter, err := contexthelpers.GetAsserter(ctx)
	if err != nil {
		return err
	}

	// TODO - you'll need to get the resource ID from the context
	resourceID := ""

	return asserter.AssertTags("", resourceID, tags)
}
