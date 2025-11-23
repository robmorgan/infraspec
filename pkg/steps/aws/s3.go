package aws

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// S3 Step Definitions
func registerS3Steps(sc *godog.ScenarioContext) {
	sc.Step(`^I have the necessary IAM permissions to describe S3 buckets$`, newVerifyAWSS3DescribeBucketsStep)
	sc.Step(`^the S3 bucket "([^"]*)" should exist$`, newS3BucketExistsStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a versioning configuration$`, newS3BucketVersioningStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a public access block$`, newS3BucketPublicAccessBlockStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a server access logging configuration$`, newS3BucketServerAccessLoggingStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have an encryption configuration$`, newS3BucketEncryptionStep)

	// Steps that read bucket name from Terraform output
	sc.Step(`^the S3 bucket from output "([^"]*)" should exist$`, newS3BucketFromOutputExistsStep)
	sc.Step(`^the S3 bucket from output "([^"]*)" should have a versioning configuration$`, newS3BucketFromOutputVersioningStep)
	sc.Step(`^the S3 bucket from output "([^"]*)" should have a public access block$`, newS3BucketFromOutputPublicAccessBlockStep)
	sc.Step(`^the S3 bucket from output "([^"]*)" should have a server access logging configuration$`, newS3BucketFromOutputServerAccessLoggingStep)
	sc.Step(`^the S3 bucket from output "([^"]*)" should have an encryption configuration$`, newS3BucketFromOutputEncryptionStep)
}

func newVerifyAWSS3DescribeBucketsStep(ctx context.Context) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertS3DescribeBuckets()
}

func newS3BucketExistsStep(ctx context.Context, bucketName string) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertBucketExists(bucketName)
}

func newS3BucketVersioningStep(ctx context.Context, bucketName string) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertBucketVersioning(bucketName)
}

func newS3BucketPublicAccessBlockStep(ctx context.Context, bucketName string) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertBucketPublicAccessBlock(bucketName)
}

func newS3BucketServerAccessLoggingStep(ctx context.Context, bucketName string) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertBucketServerAccessLogging(bucketName)
}

func newS3BucketEncryptionStep(ctx context.Context, bucketName string) error {
	s3Assert, err := getS3Asserter(ctx)
	if err != nil {
		return err
	}
	return s3Assert.AssertBucketEncryption(bucketName)
}

func getS3Asserter(ctx context.Context) (aws.S3Asserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	s3Assert, ok := asserter.(aws.S3Asserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement S3Asserter")
	}
	return s3Assert, nil
}

// Step functions that read bucket name from Terraform output

func newS3BucketFromOutputExistsStep(ctx context.Context, outputName string) error {
	bucketName, err := getBucketNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newS3BucketExistsStep(ctx, bucketName)
}

func newS3BucketFromOutputVersioningStep(ctx context.Context, outputName string) error {
	bucketName, err := getBucketNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newS3BucketVersioningStep(ctx, bucketName)
}

func newS3BucketFromOutputPublicAccessBlockStep(ctx context.Context, outputName string) error {
	bucketName, err := getBucketNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newS3BucketPublicAccessBlockStep(ctx, bucketName)
}

func newS3BucketFromOutputServerAccessLoggingStep(ctx context.Context, outputName string) error {
	bucketName, err := getBucketNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newS3BucketServerAccessLoggingStep(ctx, bucketName)
}

func newS3BucketFromOutputEncryptionStep(ctx context.Context, outputName string) error {
	bucketName, err := getBucketNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newS3BucketEncryptionStep(ctx, bucketName)
}

// Helper function to get bucket name from Terraform output
func getBucketNameFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	bucketName, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket name from output %s: %w", outputName, err)
	}
	return bucketName, nil
}
