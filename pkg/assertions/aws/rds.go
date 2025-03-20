package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/gruntwork-io/terratest/modules/testing"
)

// Ensure the `AWSAsserter` struct implements the `RDSAsserter` interface.
var _ RDSAsserter = (*AWSAsserter)(nil)

// RDSAsserter defines RDS-specific assertions
type RDSAsserter interface {
	AssertDBInstanceExists(t testing.TestingT, dbInstanceID string) error
	AssertDBInstanceClass(t testing.TestingT, dbInstanceID, instanceClass string) error
	AssertDBInstanceEngine(t testing.TestingT, dbInstanceID, engine string) error
	AssertDBInstanceStorage(t testing.TestingT, dbInstanceID string, allocatedStorage int32) error
	AssertDBInstanceMultiAZ(t testing.TestingT, dbInstanceID string, multiAZ bool) error
	AssertDBInstanceEncryption(t testing.TestingT, dbInstanceID string, encrypted bool) error
	AssertDBInstanceTags(t testing.TestingT, dbInstanceID string, expectedTags map[string]string) error
}

// AssertDBInstanceExists checks if a DB instance exists
func (a *AWSAsserter) AssertDBInstanceExists(t testing.TestingT, dbInstanceID string) error {
	// Create a new AWS RDS client
	config, err := awsgo.NewConfig(&aws.Config{Region: aws.String(a.region)})
	if err != nil {
		return err
	}

	client := rds.New(config)

	// Describe the DB instance
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: awsgo.String(dbInstanceID),
	}

	result, err := client.DescribeDBInstances(input)
	if err != nil {
		return fmt.Errorf("error describing DB instance %s: %v", dbInstanceID, err)
	}

	// Check if the DB instance exists
	if len(result.DBInstances) == 0 {
		return fmt.Errorf("DB instance %s does not exist", dbInstanceID)
	}

	return nil
}

// AssertDBInstanceClass checks if a DB instance has the expected instance class
func (a *AWSAsserter) AssertDBInstanceClass(t testing.TestingT, dbInstanceID, instanceClass string) error {
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	if aws.ToString(instance.DBInstanceClass) != instanceClass {
		return fmt.Errorf("expected instance class %s, but got %s", instanceClass, *instance.DBInstanceClass)
	}

	return nil
}

// AssertDBInstanceEngine checks if a DB instance has the expected engine
func (a *AWSAsserter) AssertDBInstanceEngine(t testing.TestingT, dbInstanceID, engine string) error {
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	if aws.ToString(instance.Engine) != engine {
		return fmt.Errorf("expected engine %s, but got %s", engine, *instance.Engine)
	}

	return nil
}

// AssertDBInstanceStorage checks if a DB instance has the expected allocated storage
func (a *AWSAsserter) AssertDBInstanceStorage(t testing.TestingT, dbInstanceID string, allocatedStorage int32) error {
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	if instance.AllocatedStorage != allocatedStorage {
		return fmt.Errorf("expected allocated storage %d, but got %d", allocatedStorage, instance.AllocatedStorage)
	}

	return nil
}

// AssertDBInstanceMultiAZ checks if a DB instance has the expected MultiAZ setting
func (a *AWSAsserter) AssertDBInstanceMultiAZ(t testing.TestingT, dbInstanceID string, multiAZ bool) error {
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	if instance.MultiAZ != multiAZ {
		return fmt.Errorf("expected MultiAZ %t, but got %t", multiAZ, instance.MultiAZ)
	}

	return nil
}

// AssertDBInstanceEncryption checks if a DB instance has the expected encryption setting
func (a *AWSAsserter) AssertDBInstanceEncryption(t testing.TestingT, dbInstanceID string, encrypted bool) error {
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	if instance.StorageEncrypted != encrypted {
		return fmt.Errorf("expected encryption %t, but got %t", encrypted, instance.StorageEncrypted)
	}

	return nil
}

// AssertDBInstanceTags checks if a DB instance has the expected tags
func (a *AWSAsserter) AssertDBInstanceTags(t testing.TestingT, dbInstanceID string, expectedTags map[string]string) error {
	// Create a new AWS RDS client
	config, err := awsgo.NewConfig(&aws.Config{Region: aws.String(a.region)})
	if err != nil {
		return err
	}

	client := rds.New(config)

	// First, get the DB instance ARN
	instance, err := a.getDBInstance(dbInstanceID)
	if err != nil {
		return err
	}

	// List tags for the DB instance
	input := &rds.ListTagsForResourceInput{
		ResourceName: instance.DBInstanceArn,
	}

	result, err := client.ListTagsForResource(input)
	if err != nil {
		return fmt.Errorf("error listing tags for DB instance %s: %v", dbInstanceID, err)
	}

	// Convert the tags to a map
	actualTags := make(map[string]string)
	for _, tag := range result.TagList {
		actualTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Compare the expected and actual tags
	for key, value := range expectedTags {
		actualValue, exists := actualTags[key]
		if !exists {
			return fmt.Errorf("expected tag %s not found", key)
		}
		if actualValue != value {
			return fmt.Errorf("expected tag %s to have value %s, but got %s", key, value, actualValue)
		}
	}

	return nil
}

// Helper method to get a DB instance
func (a *AWSAsserter) getDBInstance(dbInstanceID string) (*types.DBInstance, error) {
	// Create a new AWS RDS client
	config, err := awsgo.NewConfig(&aws.Config{Region: aws.String(a.region)})
	if err != nil {
		return nil, err
	}

	client := rds.New(config)

	// Describe the DB instance
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: awsgo.String(dbInstanceID),
	}

	result, err := client.DescribeDBInstances(input)
	if err != nil {
		return nil, fmt.Errorf("error describing DB instance %s: %v", dbInstanceID, err)
	}

	// Check if the DB instance exists
	if len(result.DBInstances) == 0 {
		return nil, fmt.Errorf("DB instance %s does not exist", dbInstanceID)
	}

	return &result.DBInstances[0], nil
}