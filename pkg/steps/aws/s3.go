package aws

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

// S3 Step Definitions
func newS3BucketExistsStep(ctx context.Context, bucketName string) error {
	asserter, err := contexthelpers.GetAsserter(ctx)
	if err != nil {
		return err
	}

	s3Assert, ok := asserter.(assertions.S3Asserter)
	if !ok {
		return fmt.Errorf("asserter does not implement S3Asserter")
	}

	return s3Assert.AssertBucketExists(bucketName)
}

func newS3BucketVersioningStep(ctx context.Context, bucketName string) error {
	asserter, err := contexthelpers.GetAsserter(ctx)
	if err != nil {
		return err
	}

	s3Assert, ok := asserter.(assertions.S3Asserter)
	if !ok {
		return fmt.Errorf("asserter does not implement S3Asserter")
	}

	return s3Assert.AssertBucketVersioning(bucketName, true)
}
