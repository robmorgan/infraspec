package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	t "github.com/robmorgan/infraspec/internal/testing"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

// RDS Step Definitions
func newRDSInstanceExistsStep(ctx context.Context, dbInstanceID string) error {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return err
	}

	rdsAssert, ok := asserter.(aws.RDSAsserter)
	if !ok {
		return fmt.Errorf("asserter does not implement RDSAsserter")
	}

	return rdsAssert.AssertDBInstanceExists(t.GetT(), dbInstanceID)
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

	return rdsAssert.AssertDBInstanceClass(t.GetT(), dbInstanceID, instanceClass)
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

	return rdsAssert.AssertDBInstanceEngine(t.GetT(), dbInstanceID, engine)
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

	return rdsAssert.AssertDBInstanceStorage(t.GetT(), dbInstanceID, allocatedStorage)
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

	return rdsAssert.AssertDBInstanceMultiAZ(t.GetT(), dbInstanceID, multiAZ)
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

	return rdsAssert.AssertDBInstanceEncryption(t.GetT(), dbInstanceID, encrypted)
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

	return rdsAssert.AssertDBInstanceTags(t.GetT(), dbInstanceID, tags)
}