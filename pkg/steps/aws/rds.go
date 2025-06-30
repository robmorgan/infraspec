package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// RDS Step Definitions
func registerRDSSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I have access to AWS RDS service$`, newVerifyAWSRDSAccessStep)
	sc.Step(`^I have the necessary IAM permissions to describe RDS instances$`, newVerifyAWSRDSDescribeInstancesStep)
	sc.Step(`^I have an RDS instance with identifier "([^"]*)"$`, newExistingRDSInstanceWithIdentifierStep)
	sc.Step(`^I describe the RDS instance$`, newRDSDescribeInstanceStep)
	sc.Step(`^the existing RDS DB instance with the identifier "([^"]*)"$`, newRDSExistingDBInstanceStep)
	sc.Step(`^the RDS instance "([^"]*)" should exist$`, newRDSInstanceExistsStep)
	sc.Step(`^the RDS instance "([^"]*)" status should be "([^"]*)"$`, newRDSInstanceStatusStep)
	sc.Step(`^the RDS instance "([^"]*)" should have instance class "([^"]*)"$`, newRDSInstanceClassStep)
	sc.Step(`^the RDS instance "([^"]*)" should have engine "([^"]*)"$`, newRDSInstanceEngineStep)
	sc.Step(`^the RDS instance "([^"]*)" should have allocated storage (\d+)$`, newRDSInstanceStorageStep)
	sc.Step(`^the RDS instance "([^"]*)" should have MultiAZ "(true|false)"$`, newRDSInstanceMultiAZStep)
	sc.Step(`^the RDS instance "([^"]*)" should have encryption "(true|false)"$`, newRDSInstanceEncryptionStep)
	sc.Step(`^the RDS instance "([^"]*)" should have tags$`, newRDSInstanceTagsStep)
}

func newVerifyAWSRDSAccessStep(ctx context.Context) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	return rdsAssert.AssertRDSServiceAccess()
}

func newVerifyAWSRDSDescribeInstancesStep(ctx context.Context) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	return rdsAssert.AssertRDSDescribeInstances()
}

func newExistingRDSInstanceWithIdentifierStep(ctx context.Context, dbInstanceID string) error {
	contexthelpers.SetRDSDBInstanceID(ctx, dbInstanceID)
	return nil
}

func newRDSDescribeInstanceStep(ctx context.Context) error {
	// do nothing for now, as we pass the identifier to the steps
	return nil
}

func newRDSExistingDBInstanceStep(ctx context.Context, dbInstanceID string) error {
	contexthelpers.SetRDSDBInstanceID(ctx, dbInstanceID)
	return nil
}

func newRDSInstanceStatusStep(ctx context.Context, dbInstanceID string, status string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceStatus(dbInstanceID, status, region)
}

func newRDSInstanceExistsStep(ctx context.Context, dbInstanceID string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceExists(dbInstanceID, region)
}

func newRDSInstanceClassStep(ctx context.Context, dbInstanceID, instanceClass string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceClass(dbInstanceID, instanceClass, region)
}

func newRDSInstanceEngineStep(ctx context.Context, dbInstanceID, engine string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceEngine(dbInstanceID, engine, region)
}

func newRDSInstanceStorageStep(ctx context.Context, dbInstanceID string, allocatedStorage int32) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return rdsAssert.AssertDBInstanceStorage(dbInstanceID, allocatedStorage, region)
}

func newRDSInstanceStorageStepWrapper(ctx context.Context, dbInstanceID string, allocatedStorageStr string) error {
	allocatedStorage, err := strconv.ParseInt(allocatedStorageStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid allocated storage value: %s", allocatedStorageStr)
	}

	return newRDSInstanceStorageStep(ctx, dbInstanceID, int32(allocatedStorage))
}

func newRDSInstanceMultiAZStep(ctx context.Context, dbInstanceID string, multiAZStr string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
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

func newRDSInstanceEncryptionStep(ctx context.Context, dbInstanceID string, encryptedStr string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
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

func newRDSInstanceTagsStep(ctx context.Context, dbInstanceID string, table *godog.Table) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
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
