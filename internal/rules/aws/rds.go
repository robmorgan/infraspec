package aws

import (
	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// RDSEncryptionEnabledRule checks that RDS instances have storage encryption enabled.
type RDSEncryptionEnabledRule struct {
	BaseRule
}

// NewRDSEncryptionEnabledRule creates a new rule that checks for RDS encryption.
func NewRDSEncryptionEnabledRule() *RDSEncryptionEnabledRule {
	return &RDSEncryptionEnabledRule{
		BaseRule: BaseRule{
			id:           "aws-rds-encryption-enabled",
			description:  "RDS instances should have storage encryption enabled",
			severity:     rules.Critical,
			resourceType: "aws_db_instance",
		},
	}
}

// Check evaluates the rule against an aws_db_instance resource.
func (r *RDSEncryptionEnabledRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	storageEncrypted := resource.GetAfterBool("storage_encrypted")
	if !storageEncrypted {
		return r.failResult(resource, "RDS instance does not have storage encryption enabled; set storage_encrypted = true"), nil
	}
	return r.passResult(resource, "RDS instance has storage encryption enabled"), nil
}

// RDSNoPublicAccessRule checks that RDS instances are not publicly accessible.
type RDSNoPublicAccessRule struct {
	BaseRule
}

// NewRDSNoPublicAccessRule creates a new rule that checks for public RDS access.
func NewRDSNoPublicAccessRule() *RDSNoPublicAccessRule {
	return &RDSNoPublicAccessRule{
		BaseRule: BaseRule{
			id:           "aws-rds-no-public-access",
			description:  "RDS instances should not be publicly accessible",
			severity:     rules.Critical,
			resourceType: "aws_db_instance",
		},
	}
}

// Check evaluates the rule against an aws_db_instance resource.
func (r *RDSNoPublicAccessRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	publiclyAccessible := resource.GetAfterBool("publicly_accessible")
	if publiclyAccessible {
		return r.failResult(resource, "RDS instance is publicly accessible; set publicly_accessible = false"), nil
	}
	return r.passResult(resource, "RDS instance is not publicly accessible"), nil
}

// RDSBackupEnabledRule checks that RDS instances have backups enabled.
type RDSBackupEnabledRule struct {
	BaseRule
}

// NewRDSBackupEnabledRule creates a new rule that checks for RDS backups.
func NewRDSBackupEnabledRule() *RDSBackupEnabledRule {
	return &RDSBackupEnabledRule{
		BaseRule: BaseRule{
			id:           "aws-rds-backup-enabled",
			description:  "RDS instances should have automated backups enabled (backup_retention_period > 0)",
			severity:     rules.Warning,
			resourceType: "aws_db_instance",
		},
	}
}

// Check evaluates the rule against an aws_db_instance resource.
func (r *RDSBackupEnabledRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	retentionPeriod := resource.GetAfterInt("backup_retention_period")
	if retentionPeriod <= 0 {
		return r.failResult(resource, "RDS instance does not have automated backups enabled; set backup_retention_period to a value greater than 0"), nil
	}
	return r.passResult(resource, "RDS instance has automated backups enabled"), nil
}
