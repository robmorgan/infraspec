package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// RDS Step Definitions
func registerRDSSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I have access to AWS RDS service$`, newVerifyAWSRDSAccessStep)
	sc.Step(`^I have the necessary IAM permissions to describe RDS instances$`, newVerifyAWSRDSDescribeInstancesStep)
	sc.Step(`^I describe the RDS instance$`, newRDSDescribeInstanceStep)
	sc.Step(`^the RDS instance "([^"]*)" should exist$`, newRDSInstanceExistsStep)
	sc.Step(`^the RDS instance "([^"]*)" status should be "([^"]*)"$`, newRDSInstanceStatusStep)
	sc.Step(`^the RDS instance "([^"]*)" instance class should be "([^"]*)"$`, newRDSInstanceClassStep)
	sc.Step(`^the RDS instance "([^"]*)" engine should be "([^"]*)"$`, newRDSInstanceEngineStep)
	sc.Step(`^the RDS instance "([^"]*)" allocated storage should be (\d+)$`, newRDSInstanceStorageStep)
	sc.Step(`^the RDS instance "([^"]*)" MultiAZ should be "(true|false)"$`, newRDSInstanceMultiAZStep)
	sc.Step(`^the RDS instance "([^"]*)" encryption should be "(true|false)"$`, newRDSInstanceEncryptionStep)
	sc.Step(`^the RDS instance "([^"]*)" should not be publicly accessible$`, newRDSInstanceNotPubliclyAccessibleStep)
	sc.Step(`^the RDS instance "([^"]*)" should have the tags$`, newRDSInstanceTagsStep)

	// Steps that read DB instance identifier from Terraform output
	sc.Step(`^the RDS instance from output "([^"]*)" should exist$`, newRDSInstanceFromOutputExistsStep)
	sc.Step(`^the RDS instance from output "([^"]*)" status should be "([^"]*)"$`, newRDSInstanceFromOutputStatusStep)
	sc.Step(`^the RDS instance from output "([^"]*)" instance class should be "([^"]*)"$`, newRDSInstanceFromOutputClassStep)
	sc.Step(`^the RDS instance from output "([^"]*)" engine should be "([^"]*)"$`, newRDSInstanceFromOutputEngineStep)
	sc.Step(`^the RDS instance from output "([^"]*)" allocated storage should be (\d+)$`, newRDSInstanceFromOutputStorageStep)
	sc.Step(`^the RDS instance from output "([^"]*)" MultiAZ should be "(true|false)"$`, newRDSInstanceFromOutputMultiAZStep)
	sc.Step(`^the RDS instance from output "([^"]*)" encryption should be "(true|false)"$`, newRDSInstanceFromOutputEncryptionStep)
	sc.Step(`^the RDS instance from output "([^"]*)" should not be publicly accessible$`, newRDSInstanceFromOutputNotPubliclyAccessibleStep)
	sc.Step(`^the RDS instance from output "([^"]*)" should have the tags$`, newRDSInstanceFromOutputTagsStep)
}

func newVerifyAWSRDSAccessStep(ctx context.Context) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}
	return rdsAssert.AssertRDSServiceAccess()
}

func newVerifyAWSRDSDescribeInstancesStep(ctx context.Context) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}
	return rdsAssert.AssertRDSDescribeInstances()
}

func newRDSDescribeInstanceStep(ctx context.Context) error {
	// do nothing for now, as we pass the identifier to the steps
	return nil
}

func newRDSInstanceStatusStep(ctx context.Context, dbInstanceID, status string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceStatus(dbInstanceID, status, region)
}

func newRDSInstanceExistsStep(ctx context.Context, dbInstanceID string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceExists(dbInstanceID, region)
}

func newRDSInstanceClassStep(ctx context.Context, dbInstanceID, instanceClass string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceClass(dbInstanceID, instanceClass, region)
}

func newRDSInstanceEngineStep(ctx context.Context, dbInstanceID, engine string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceEngine(dbInstanceID, engine, region)
}

func newRDSInstanceStorageStep(ctx context.Context, dbInstanceID string, allocatedStorage int32) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceStorage(dbInstanceID, allocatedStorage, region)
}

func newRDSInstanceMultiAZStep(ctx context.Context, dbInstanceID, multiAZStr string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	multiAZ, err := strconv.ParseBool(multiAZStr)
	if err != nil {
		return fmt.Errorf("invalid MultiAZ value: %s", multiAZStr)
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceMultiAZ(dbInstanceID, multiAZ, region)
}

func newRDSInstanceEncryptionStep(ctx context.Context, dbInstanceID, encryptedStr string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	encrypted, err := strconv.ParseBool(encryptedStr)
	if err != nil {
		return fmt.Errorf("invalid encryption value: %s", encryptedStr)
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceEncryption(dbInstanceID, encrypted, region)
}

func newRDSInstanceNotPubliclyAccessibleStep(ctx context.Context, dbInstance string) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstancePubliclyAccessible(dbInstance, false, region)
}

func newRDSInstanceTagsStep(ctx context.Context, dbInstanceID string, table *godog.Table) error {
	rdsAssert, err := getRdsAsserter(ctx)
	if err != nil {
		return err
	}

	// Convert the table to a map of tags
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceTags(dbInstanceID, tags, region)
}

func getRdsAsserter(ctx context.Context) (aws.RDSAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement RDSAsserter")
	}
	return rdsAssert, nil
}

// Step functions that read DB instance identifier from Terraform output

func newRDSInstanceFromOutputExistsStep(ctx context.Context, outputName string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceExistsStep(ctx, dbInstanceID)
}

func newRDSInstanceFromOutputStatusStep(ctx context.Context, outputName, status string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceStatusStep(ctx, dbInstanceID, status)
}

func newRDSInstanceFromOutputClassStep(ctx context.Context, outputName, instanceClass string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceClassStep(ctx, dbInstanceID, instanceClass)
}

func newRDSInstanceFromOutputEngineStep(ctx context.Context, outputName, engine string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceEngineStep(ctx, dbInstanceID, engine)
}

func newRDSInstanceFromOutputStorageStep(ctx context.Context, outputName string, allocatedStorage int32) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceStorageStep(ctx, dbInstanceID, allocatedStorage)
}

func newRDSInstanceFromOutputMultiAZStep(ctx context.Context, outputName, multiAZStr string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceMultiAZStep(ctx, dbInstanceID, multiAZStr)
}

func newRDSInstanceFromOutputEncryptionStep(ctx context.Context, outputName, encryptedStr string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceEncryptionStep(ctx, dbInstanceID, encryptedStr)
}

func newRDSInstanceFromOutputNotPubliclyAccessibleStep(ctx context.Context, outputName string) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceNotPubliclyAccessibleStep(ctx, dbInstanceID)
}

func newRDSInstanceFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	dbInstanceID, err := getDBInstanceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newRDSInstanceTagsStep(ctx, dbInstanceID, table)
}

// Helper function to get DB instance identifier from Terraform output
func getDBInstanceIDFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	dbInstanceID, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get DB instance ID from output %s: %w", outputName, err)
	}
	return dbInstanceID, nil
}
