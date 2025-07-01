package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `RDSAsserter` interface.
var _ RDSAsserter = (*AWSAsserter)(nil)

// RDSAsserter defines RDS-specific assertions
type RDSAsserter interface {
	AssertRDSServiceAccess() error
	AssertRDSDescribeInstances() error
	AssertDBInstanceExists(dbInstanceID, region string) error
	AssertDBInstanceStatus(dbInstanceID, status, region string) error
	AssertDBInstanceClass(dbInstanceID, instanceClass, region string) error
	AssertDBInstanceEngine(dbInstanceID, engine, region string) error
	AssertDBInstanceStorage(dbInstanceID string, allocatedStorage int32, region string) error
	AssertDBInstanceMultiAZ(dbInstanceID string, multiAZ bool, region string) error
	AssertDBInstanceEncryption(dbInstanceID string, encrypted bool, region string) error
	AssertDBInstancePubliclyAccessible(dbInstanceID string, publiclyAccessible bool, region string) error
	AssertDBInstanceTags(dbInstanceID string, expectedTags map[string]string, region string) error
}

// AssertRDSServiceAccess checks if the AWS account has permission to access the RDS service
//
// TODO: This doesn't work on LocalStack as the API isn't supported, so we're best off leaving this call undocumented,
// until its ported to use something like the IAM policy simulator instead.
func (a *AWSAsserter) AssertRDSServiceAccess() error {
	client, err := awshelpers.NewRdsClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.DescribeAccountAttributes(context.TODO(), &rds.DescribeAccountAttributesInput{})
	if err != nil {
		return fmt.Errorf("error accessing the RDS service: %v", err)
	}

	return nil
}

// AssertRDSDescribeInstances checks if the AWS account has permission to describe RDS instances
func (a *AWSAsserter) AssertRDSDescribeInstances() error {
	// Use the default region
	client, err := awshelpers.NewRdsClientWithDefaultRegion()
	if err != nil {
		return err
	}

	// Describe the DB instances
	_, err = client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return fmt.Errorf("error describing DB instances: %v", err)
	}

	return nil
}

// AssertDBInstanceExists checks if a DB instance exists
func (a *AWSAsserter) AssertDBInstanceExists(dbInstanceID, region string) error {
	client, err := awshelpers.NewRdsClient(region)
	if err != nil {
		return err
	}

	// Describe the DB instance
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
	}

	result, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error describing DB instance %s: %v", dbInstanceID, err)
	}

	// Check if the DB instance exists
	if len(result.DBInstances) == 0 {
		return fmt.Errorf("DB instance %s does not exist", dbInstanceID)
	}

	return nil
}

// AssertDBInstanceStatus checks if a DB instance has the expected status
func (a *AWSAsserter) AssertDBInstanceStatus(dbInstanceID, status, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.DBInstanceStatus) != status {
		return fmt.Errorf("expected instance status %s, but got %s", status, *instance.DBInstanceStatus)
	}

	return nil
}

// AssertDBInstanceClass checks if a DB instance has the expected instance class
func (a *AWSAsserter) AssertDBInstanceClass(dbInstanceID, instanceClass, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.DBInstanceClass) != instanceClass {
		return fmt.Errorf("expected instance class %s, but got %s", instanceClass, *instance.DBInstanceClass)
	}

	return nil
}

// AssertDBInstanceEngine checks if a DB instance has the expected engine
func (a *AWSAsserter) AssertDBInstanceEngine(dbInstanceID, engine, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.Engine) != engine {
		return fmt.Errorf("expected engine %s, but got %s", engine, aws.ToString(instance.Engine))
	}

	return nil
}

// AssertDBInstanceStorage checks if a DB instance has the expected allocated storage
func (a *AWSAsserter) AssertDBInstanceStorage(dbInstanceID string, allocatedStorage int32, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToInt32(instance.AllocatedStorage) != allocatedStorage {
		return fmt.Errorf("expected allocated storage %d, but got %d", allocatedStorage, aws.ToInt32(instance.AllocatedStorage))
	}

	return nil
}

// AssertDBInstanceMultiAZ checks if a DB instance has the expected MultiAZ setting
func (a *AWSAsserter) AssertDBInstanceMultiAZ(dbInstanceID string, multiAZ bool, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToBool(instance.MultiAZ) != multiAZ {
		return fmt.Errorf("expected MultiAZ %t, but got %t", multiAZ, aws.ToBool(instance.MultiAZ))
	}

	return nil
}

// AssertDBInstanceEncryption checks if a DB instance has the expected encryption setting
func (a *AWSAsserter) AssertDBInstanceEncryption(dbInstanceID string, encrypted bool, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToBool(instance.StorageEncrypted) != encrypted {
		return fmt.Errorf("expected encryption %t, but got %t", encrypted, aws.ToBool(instance.StorageEncrypted))
	}

	return nil
}

func (a *AWSAsserter) AssertDBInstancePubliclyAccessible(dbInstanceID string, publiclyAccessible bool, region string) error {
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	if aws.ToBool(instance.PubliclyAccessible) != publiclyAccessible {
		return fmt.Errorf("expected publicly accessible %t, but got %t", publiclyAccessible, aws.ToBool(instance.PubliclyAccessible))
	}

	return nil
}

// AssertDBInstanceTags checks if a DB instance has the expected tags
func (a *AWSAsserter) AssertDBInstanceTags(dbInstanceID string, expectedTags map[string]string, region string) error {
	client, err := awshelpers.NewRdsClientWithDefaultRegion()
	if err != nil {
		return err
	}

	// First, get the DB instance ARN
	instance, err := a.getDBInstance(dbInstanceID, region)
	if err != nil {
		return err
	}

	// List tags for the DB instance
	input := &rds.ListTagsForResourceInput{
		ResourceName: instance.DBInstanceArn,
	}

	result, err := client.ListTagsForResource(context.TODO(), input)
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
func (a *AWSAsserter) getDBInstance(dbInstanceID string, region string) (*types.DBInstance, error) {
	client, err := awshelpers.NewRdsClient(region)
	if err != nil {
		return nil, err
	}

	// Describe the DB instance
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
	}

	result, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing DB instance %s: %v", dbInstanceID, err)
	}

	// Check if the DB instance exists
	if len(result.DBInstances) == 0 {
		return nil, fmt.Errorf("DB instance %s does not exist", dbInstanceID)
	}

	return &result.DBInstances[0], nil
}
