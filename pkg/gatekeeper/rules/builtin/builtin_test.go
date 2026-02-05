package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

func TestLoadBuiltinRules(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	// Should have loaded all built-in rules
	assert.Greater(t, len(loadedRules), 0, "should have loaded some rules")

	// Check for expected rule IDs
	ruleIDs := make(map[string]bool)
	for _, r := range loadedRules {
		ruleIDs[r.ID] = true
	}

	// S3 rules
	assert.True(t, ruleIDs["S3_001"], "should have S3_001 rule")
	assert.True(t, ruleIDs["S3_002"], "should have S3_002 rule")
	assert.True(t, ruleIDs["S3_003"], "should have S3_003 rule")
	assert.True(t, ruleIDs["S3_004"], "should have S3_004 rule")

	// Security group rules
	assert.True(t, ruleIDs["SG_001"], "should have SG_001 rule")
	assert.True(t, ruleIDs["SG_002"], "should have SG_002 rule")
	assert.True(t, ruleIDs["SG_003"], "should have SG_003 rule")
	assert.True(t, ruleIDs["SG_004"], "should have SG_004 rule")

	// VPC rules
	assert.True(t, ruleIDs["VPC_001"], "should have VPC_001 rule")
	assert.True(t, ruleIDs["VPC_002"], "should have VPC_002 rule")

	// IAM rules
	assert.True(t, ruleIDs["IAM_001"], "should have IAM_001 rule")
	assert.True(t, ruleIDs["IAM_002"], "should have IAM_002 rule")
	assert.True(t, ruleIDs["IAM_003"], "should have IAM_003 rule")
}

func TestBuiltinRules_ValidStructure(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	for _, rule := range loadedRules {
		t.Run(rule.ID, func(t *testing.T) {
			// Required fields
			assert.NotEmpty(t, rule.ID, "rule should have ID")
			assert.NotEmpty(t, rule.Name, "rule should have name")
			assert.NotEmpty(t, rule.ResourceType, "rule should have resource_type")
			assert.NotEmpty(t, rule.Message, "rule should have message")

			// Severity should be valid
			assert.True(t, rule.Severity >= rules.SeverityError && rule.Severity <= rules.SeverityInfo,
				"rule should have valid severity")

			// Condition should be valid
			err := rule.Condition.Validate()
			assert.NoError(t, err, "rule condition should be valid")
		})
	}
}

func TestBuiltinRules_SeverityDistribution(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	var errors, warnings, infos int
	for _, rule := range loadedRules {
		switch rule.Severity {
		case rules.SeverityError:
			errors++
		case rules.SeverityWarning:
			warnings++
		case rules.SeverityInfo:
			infos++
		}
	}

	// Should have a mix of severities
	assert.Greater(t, errors, 0, "should have error-level rules")
	assert.Greater(t, warnings, 0, "should have warning-level rules")
	// Info rules are optional

	t.Logf("Severity distribution: %d errors, %d warnings, %d infos", errors, warnings, infos)
}

func TestBuiltinRules_ResourceTypes(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	resourceTypes := make(map[string]int)
	for _, rule := range loadedRules {
		resourceTypes[rule.ResourceType]++
	}

	// Should cover expected resource types
	expectedTypes := []string{
		"aws_s3_bucket",
		"aws_s3_bucket_public_access_block",
		"aws_security_group",
		"aws_flow_log",
		"aws_default_security_group",
		"aws_iam_role",
		"aws_iam_policy",
	}

	for _, rt := range expectedTypes {
		assert.Greater(t, resourceTypes[rt], 0, "should have rules for %s", rt)
	}

	t.Logf("Resource types covered: %v", resourceTypes)
}

func TestBuiltinRules_Tags(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	tags := make(map[string]int)
	for _, rule := range loadedRules {
		for _, tag := range rule.Tags {
			tags[tag]++
		}
	}

	// Should have common tags
	assert.Greater(t, tags["security"], 0, "should have security-tagged rules")

	t.Logf("Tags used: %v", tags)
}

func TestBuiltinRules_UniqueIDs(t *testing.T) {
	loadedRules, err := LoadBuiltinRules()
	require.NoError(t, err)

	seen := make(map[string]bool)
	for _, rule := range loadedRules {
		assert.False(t, seen[rule.ID], "rule ID %s should be unique", rule.ID)
		seen[rule.ID] = true
	}
}
