package aws

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

// RegisterSteps registers all AWS-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	// DynamoDB steps
	registerDynamoDBSteps(sc)

	// S3 steps
	registerS3Steps(sc)

	// RDS steps
	registerRDSSteps(sc)

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

	return asserter.AssertTags("", resourceID, tags)
}
