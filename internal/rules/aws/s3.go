package aws

import (
	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// S3NoPublicACLRule checks that S3 buckets don't use public ACLs.
type S3NoPublicACLRule struct {
	BaseRule
}

// NewS3NoPublicACLRule creates a new rule that checks for public ACLs.
func NewS3NoPublicACLRule() *S3NoPublicACLRule {
	return &S3NoPublicACLRule{
		BaseRule: BaseRule{
			id:           "aws-s3-no-public-acl",
			description:  "S3 buckets should not use public-read or public-read-write ACLs",
			severity:     rules.Critical,
			resourceType: "aws_s3_bucket",
		},
	}
}

// Check evaluates the rule against an aws_s3_bucket resource.
func (r *S3NoPublicACLRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	acl := resource.GetAfterString("acl")
	if acl == "public-read" || acl == "public-read-write" {
		return r.failResult(resource, "S3 bucket uses a public ACL ("+acl+"); use private ACL and bucket policies instead"), nil
	}
	return r.passResult(resource, "S3 bucket does not use a public ACL"), nil
}

// S3EncryptionEnabledRule checks that S3 buckets have server-side encryption configured.
type S3EncryptionEnabledRule struct {
	BaseRule
}

// NewS3EncryptionEnabledRule creates a new rule that checks for encryption.
func NewS3EncryptionEnabledRule() *S3EncryptionEnabledRule {
	return &S3EncryptionEnabledRule{
		BaseRule: BaseRule{
			id:           "aws-s3-encryption-enabled",
			description:  "S3 buckets should have server-side encryption configured",
			severity:     rules.Critical,
			resourceType: "aws_s3_bucket",
		},
	}
}

// Check evaluates the rule against an aws_s3_bucket resource.
func (r *S3EncryptionEnabledRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	// Check for server_side_encryption_configuration block
	sseConfig, ok := resource.GetAfter("server_side_encryption_configuration")
	if !ok || sseConfig == nil {
		return r.failResult(resource, "S3 bucket does not have server-side encryption configured; add server_side_encryption_configuration block"), nil
	}

	// Verify it's not an empty slice
	if slice, ok := sseConfig.([]interface{}); ok && len(slice) == 0 {
		return r.failResult(resource, "S3 bucket does not have server-side encryption configured; add server_side_encryption_configuration block"), nil
	}

	return r.passResult(resource, "S3 bucket has server-side encryption configured"), nil
}

// S3VersioningEnabledRule checks that S3 buckets have versioning enabled.
type S3VersioningEnabledRule struct {
	BaseRule
}

// NewS3VersioningEnabledRule creates a new rule that checks for versioning.
func NewS3VersioningEnabledRule() *S3VersioningEnabledRule {
	return &S3VersioningEnabledRule{
		BaseRule: BaseRule{
			id:           "aws-s3-versioning-enabled",
			description:  "S3 buckets should have versioning enabled for data protection",
			severity:     rules.Warning,
			resourceType: "aws_s3_bucket",
		},
	}
}

// Check evaluates the rule against an aws_s3_bucket resource.
func (r *S3VersioningEnabledRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	// Check for versioning[0].enabled = true
	versioning, ok := resource.GetAfter("versioning")
	if !ok {
		return r.failResult(resource, "S3 bucket does not have versioning enabled; add versioning { enabled = true } block"), nil
	}

	versioningSlice, ok := versioning.([]interface{})
	if !ok || len(versioningSlice) == 0 {
		return r.failResult(resource, "S3 bucket does not have versioning enabled; add versioning { enabled = true } block"), nil
	}

	versioningConfig, ok := versioningSlice[0].(map[string]interface{})
	if !ok {
		return r.failResult(resource, "S3 bucket does not have versioning enabled; add versioning { enabled = true } block"), nil
	}

	enabled, ok := versioningConfig["enabled"].(bool)
	if !ok || !enabled {
		return r.failResult(resource, "S3 bucket versioning is not enabled; set versioning { enabled = true }"), nil
	}

	return r.passResult(resource, "S3 bucket has versioning enabled"), nil
}

// S3NoPublicPolicyRule checks that S3 public access blocks have block_public_policy enabled.
type S3NoPublicPolicyRule struct {
	BaseRule
}

// NewS3NoPublicPolicyRule creates a new rule that checks for public policy blocking.
func NewS3NoPublicPolicyRule() *S3NoPublicPolicyRule {
	return &S3NoPublicPolicyRule{
		BaseRule: BaseRule{
			id:           "aws-s3-no-public-policy",
			description:  "S3 public access blocks should have block_public_policy enabled",
			severity:     rules.Critical,
			resourceType: "aws_s3_bucket_public_access_block",
		},
	}
}

// Check evaluates the rule against an aws_s3_bucket_public_access_block resource.
func (r *S3NoPublicPolicyRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	blockPublicPolicy := resource.GetAfterBool("block_public_policy")
	if !blockPublicPolicy {
		return r.failResult(resource, "S3 public access block does not block public policies; set block_public_policy = true"), nil
	}
	return r.passResult(resource, "S3 public access block has block_public_policy enabled"), nil
}
