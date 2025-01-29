package assertions

import "fmt"

// Asserter defines the interface for all cloud resource assertions
type Asserter interface {
	// Common assertions
	AssertExists(resourceType, resourceName string) error
	AssertTags(resourceType, resourceName string, tags map[string]string) error

	// Provider-specific assertions must be implemented by concrete types
}

// Factory function to create new asserters
func New(provider, region string) (Asserter, error) {
	switch provider {
	case "aws":
		return NewAWSAsserter(region)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// AWSAsserter implements assertions for AWS resources
type AWSAsserter struct {
	region string
}

// NewAWSAsserter creates a new AWSAsserter instance
func NewAWSAsserter(region string) (*AWSAsserter, error) {
	return &AWSAsserter{
		region: region,
	}, nil
}

// AssertExists checks if a resource exists
func (a *AWSAsserter) AssertExists(resourceType, resourceName string) error {
	// Implement AWS-specific logic to check resource existence
	return nil
}

// AssertTags checks if a resource has the expected tags
func (a *AWSAsserter) AssertTags(resourceType, resourceName string, tags map[string]string) error {
	// Implement AWS-specific logic to check resource tags
	return nil
}

// DynamoDBAsserter defines DynamoDB-specific assertions
type DynamoDBAsserter interface {
	Asserter
	AssertBillingMode(tableName, expectedMode string) error
	AssertCapacity(tableName string, readCapacity, writeCapacity int64) error
	AssertGSI(tableName, indexName, keyAttribute string) error
}

// S3Asserter defines S3-specific assertions
type S3Asserter interface {
	Asserter
	AssertBucketExists(bucketName string) error
	AssertBucketVersioning(bucketName string, enabled bool) error
	AssertBucketEncryption(bucketName string, encryptionType string) error
}
