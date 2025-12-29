package aws

import (
	"context"

	"github.com/cucumber/godog"
)

// RegisterSteps registers all AWS-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	// DynamoDB steps
	registerDynamoDBSteps(sc)

	// S3 steps
	registerS3Steps(sc)

	// RDS steps
	registerRDSSteps(sc)

	// EC2 steps
	registerEC2Steps(sc)

	// IAM steps
	registerIAMSteps(sc)

	// SQS steps
	registerSQSSteps(sc)

	// Lambda steps
	registerLambdaSteps(sc)

	// Generic AWS steps
	sc.Step(`^the AWS resource "([^"]*)" should exist$`, newAWSResourceExistsStep)
}

// Generic AWS Steps
func newAWSResourceExistsStep(ctx context.Context, resourceID string) error {
	// TODO - implement
	return nil
}
