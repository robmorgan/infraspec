package aws

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// S3 Step Definitions
func registerS3Steps(sc *godog.ScenarioContext) {
	sc.Step(`^the S3 bucket "([^"]*)" should exist$`, newS3BucketExistsStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a versioning configuration$`, newS3BucketVersioningStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a public access block$`, newS3BucketPublicAccessBlockStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have a server access logging configuration$`, newS3BucketServerAccessLoggingStep)
	sc.Step(`^the S3 bucket "([^"]*)" should have an encryption configuration$`, newS3BucketEncryptionStep)
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
