package aws

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	t "github.com/robmorgan/infraspec/internal/testing"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

// RegisterSteps registers all AWS-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	// DynamoDB steps
	sc.Step(`^the DynamoDB table "([^"]*)" should have tags$`, newDynamoDBTagsStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have billing mode "([^"]*)"$`, newDynamoDBBillingModeStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have read capacity (\d+)$`, newDynamoDBReadCapacityStep)
	sc.Step(`^the DynamoDB table "([^"]*)" should have write capacity (\d+)$`, newDynamoDBWriteCapacityStep)

	// S3 steps
	// sc.Step(`^the S3 bucket "([^"]*)" should exist$`, newS3BucketExistsStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have versioning enabled$`, newS3BucketVersioningStep(ctx))
	// sc.Step(`^the S3 bucket "([^"]*)" should have encryption "([^"]*)"$`, newS3BucketEncryptionStep(ctx))

	// RDS steps
	sc.Step(`^the RDS instance "([^"]*)" should exist$`, newRDSInstanceExistsStep)
	sc.Step(`^the RDS instance "([^"]*)" should have instance class "([^"]*)"$`, newRDSInstanceClassStep)
	sc.Step(`^the RDS instance "([^"]*)" should have engine "([^"]*)"$`, newRDSInstanceEngineStep)
	sc.Step(`^the RDS instance "([^"]*)" should have allocated storage (\d+)$`, newRDSInstanceStorageStepWrapper)
	sc.Step(`^the RDS instance "([^"]*)" should have MultiAZ "(true|false)"$`, newRDSInstanceMultiAZStep)
	sc.Step(`^the RDS instance "([^"]*)" should have encryption "(true|false)"$`, newRDSInstanceEncryptionStep)
	sc.Step(`^the RDS instance "([^"]*)" should have tags$`, newRDSInstanceTagsStep)

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

	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	// TODO - you'll need to get the resource ID from the context
	resourceID := ""

	return asserter.AssertTags(t.GetT(), "", resourceID, tags)
}