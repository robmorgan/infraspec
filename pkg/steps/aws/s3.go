package aws

import (
	"fmt"

	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/pkg/assertions"
)

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
