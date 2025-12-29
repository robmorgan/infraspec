package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `S3Asserter` interface.
var _ S3Asserter = (*AWSAsserter)(nil)

// S3Asserter defines S3-specific assertions
type S3Asserter interface {
	AssertS3DescribeBuckets() error
	AssertBucketExists(bucketName string) error
	AssertBucketVersioning(bucketName string) error
	AssertBucketEncryption(bucketName string) error
	AssertBucketPublicAccessBlock(bucketName string) error
	AssertBucketServerAccessLogging(bucketName string) error
}

// AssertS3DescribeBuckets checks if the AWS account has permission to describe S3 buckets
func (a *AWSAsserter) AssertS3DescribeBuckets() error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	// List buckets to verify access
	_, err = client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("error listing S3 buckets: %w", err)
	}

	return nil
}

func (a *AWSAsserter) AssertBucketExists(bucketName string) error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	// Use HeadBucket to check if bucket exists and we have access
	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("bucket %s does not exist or is not accessible: %w", bucketName, err)
	}

	return nil
}

func (a *AWSAsserter) AssertBucketVersioning(bucketName string) error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	input := &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	}

	result, err := client.GetBucketVersioning(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error getting bucket versioning for %s: %w", bucketName, err)
	}

	if result.Status != types.BucketVersioningStatusEnabled {
		return fmt.Errorf("bucket %s versioning is not enabled, status: %s", bucketName, result.Status)
	}

	return nil
}

func (a *AWSAsserter) AssertBucketEncryption(bucketName string) error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	input := &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}

	result, err := client.GetBucketEncryption(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error getting bucket encryption for %s: %w", bucketName, err)
	}

	if result.ServerSideEncryptionConfiguration == nil || len(result.ServerSideEncryptionConfiguration.Rules) == 0 {
		return fmt.Errorf("bucket %s does not have encryption configuration", bucketName)
	}

	return nil
}

func (a *AWSAsserter) AssertBucketPublicAccessBlock(bucketName string) error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	input := &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	}

	result, err := client.GetPublicAccessBlock(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error getting public access block for %s: %w", bucketName, err)
	}

	if result.PublicAccessBlockConfiguration == nil {
		return fmt.Errorf("bucket %s does not have public access block configuration", bucketName)
	}

	return nil
}

func (a *AWSAsserter) AssertBucketServerAccessLogging(bucketName string) error {
	client, err := a.createS3Client()
	if err != nil {
		return err
	}

	input := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}

	result, err := client.GetBucketLogging(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error getting bucket logging for %s: %w", bucketName, err)
	}

	if result.LoggingEnabled == nil {
		return fmt.Errorf("bucket %s does not have server access logging configuration", bucketName)
	}

	return nil
}

// Helper method to create an S3 client
func (a *AWSAsserter) createS3Client() (*s3.Client, error) {
	cfg, err := awshelpers.NewAuthenticatedSessionWithDefaultRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	opts := make([]func(*s3.Options), 0, 2)

	if endpoint, ok := awshelpers.GetVirtualCloudEndpoint("s3"); ok {
		// When using virtual cloud, use virtual-hosted style URLs
		// (e.g., http://bucket.s3.infraspec.sh/key or http://bucket.s3.localhost:3687/key)
		// instead of path-style (e.g., http://s3.infraspec.sh/bucket/key)
		opts = append(opts, func(o *s3.Options) {
			o.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
			o.UsePathStyle = false
		})
	}

	return s3.NewFromConfig(*cfg, opts...), nil
}
