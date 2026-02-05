//nolint:dupl // Test functions have intentional structural similarity for readability
package aws

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// runRuleTest is a helper function to run a rule against a specific resource in a plan file.
func runRuleTest(t *testing.T, rule rules.Rule, planFile, resourceAddr string, expectPass bool, msgContains string) {
	t.Helper()

	p, err := plan.ParsePlanFile("testdata/plans/" + planFile)
	require.NoError(t, err, "Failed to parse plan file")

	resource := p.ResourceByAddress(resourceAddr)
	require.NotNil(t, resource, "Resource %s not found in plan", resourceAddr)

	result, err := rule.Check(resource)
	require.NoError(t, err, "Rule check returned error")
	require.NotNil(t, result, "Rule check returned nil result")

	if expectPass {
		assert.True(t, result.Passed, "Expected rule to pass, but it failed: %s", result.Message)
	} else {
		assert.False(t, result.Passed, "Expected rule to fail, but it passed: %s", result.Message)
		if msgContains != "" {
			assert.True(t, strings.Contains(result.Message, msgContains),
				"Expected message to contain %q, got %q", msgContains, result.Message)
		}
	}

	assert.Equal(t, rule.ID(), result.RuleID)
	assert.Equal(t, resourceAddr, result.ResourceAddress)
}

// ========================================
// Security Group Rules Tests
// ========================================

func TestSGNoPublicSSH(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when SSH open to world",
			planFile:     "sg_public_ssh.json",
			resourceAddr: "aws_security_group.bad",
			expectPass:   false,
			msgContains:  "SSH (port 22)",
		},
		{
			name:         "passes when SSH restricted to private network",
			planFile:     "sg_secure.json",
			resourceAddr: "aws_security_group.good",
			expectPass:   true,
		},
	}

	rule := NewSGNoPublicSSHRule()
	assert.Equal(t, "aws-sg-no-public-ssh", rule.ID())
	assert.Equal(t, "aws", rule.Provider())
	assert.Equal(t, "aws_security_group", rule.ResourceType())
	assert.Equal(t, rules.Critical, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestSGRuleNoPublicSSH(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when SSH rule open to world",
			planFile:     "sg_rule_public_ssh.json",
			resourceAddr: "aws_security_group_rule.bad_ssh",
			expectPass:   false,
			msgContains:  "SSH (port 22)",
		},
		{
			name:         "passes when SSH rule restricted",
			planFile:     "sg_rule_public_ssh.json",
			resourceAddr: "aws_security_group_rule.good_ssh",
			expectPass:   true,
		},
		{
			name:         "passes for egress rules",
			planFile:     "sg_rule_public_ssh.json",
			resourceAddr: "aws_security_group_rule.egress",
			expectPass:   true,
		},
	}

	rule := NewSGRuleNoPublicSSHRule()
	assert.Equal(t, "aws-sg-rule-no-public-ssh", rule.ID())
	assert.Equal(t, "aws_security_group_rule", rule.ResourceType())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestSGNoPublicRDP(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when RDP open to world",
			planFile:     "sg_public_databases.json",
			resourceAddr: "aws_security_group.public_rdp",
			expectPass:   false,
			msgContains:  "RDP (port 3389)",
		},
		{
			name:         "passes when RDP not configured",
			planFile:     "sg_secure.json",
			resourceAddr: "aws_security_group.good",
			expectPass:   true,
		},
	}

	rule := NewSGNoPublicRDPRule()
	assert.Equal(t, "aws-sg-no-public-rdp", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestSGNoPublicMySQL(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when MySQL open to world",
			planFile:     "sg_public_databases.json",
			resourceAddr: "aws_security_group.public_mysql",
			expectPass:   false,
			msgContains:  "MySQL (port 3306)",
		},
		{
			name:         "passes when MySQL not configured",
			planFile:     "sg_secure.json",
			resourceAddr: "aws_security_group.good",
			expectPass:   true,
		},
	}

	rule := NewSGNoPublicMySQLRule()
	assert.Equal(t, "aws-sg-no-public-mysql", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestSGNoPublicPostgres(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when PostgreSQL open to world",
			planFile:     "sg_public_databases.json",
			resourceAddr: "aws_security_group.public_postgres",
			expectPass:   false,
			msgContains:  "PostgreSQL (port 5432)",
		},
		{
			name:         "passes when PostgreSQL not configured",
			planFile:     "sg_secure.json",
			resourceAddr: "aws_security_group.good",
			expectPass:   true,
		},
	}

	rule := NewSGNoPublicPostgresRule()
	assert.Equal(t, "aws-sg-no-public-postgres", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestSGNoUnrestrictedIngress(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when all traffic allowed from world",
			planFile:     "sg_unrestricted.json",
			resourceAddr: "aws_security_group.unrestricted",
			expectPass:   false,
			msgContains:  "all traffic",
		},
		{
			name:         "passes when traffic restricted",
			planFile:     "sg_secure.json",
			resourceAddr: "aws_security_group.good",
			expectPass:   true,
		},
	}

	rule := NewSGNoUnrestrictedIngressRule()
	assert.Equal(t, "aws-sg-no-unrestricted-ingress", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

// ========================================
// S3 Rules Tests
// ========================================

func TestS3NoPublicACL(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails with public-read ACL",
			planFile:     "s3_public_acl.json",
			resourceAddr: "aws_s3_bucket.public_read",
			expectPass:   false,
			msgContains:  "public ACL",
		},
		{
			name:         "fails with public-read-write ACL",
			planFile:     "s3_public_acl.json",
			resourceAddr: "aws_s3_bucket.public_read_write",
			expectPass:   false,
			msgContains:  "public ACL",
		},
		{
			name:         "passes with private ACL",
			planFile:     "s3_secure.json",
			resourceAddr: "aws_s3_bucket.secure",
			expectPass:   true,
		},
	}

	rule := NewS3NoPublicACLRule()
	assert.Equal(t, "aws-s3-no-public-acl", rule.ID())
	assert.Equal(t, "aws_s3_bucket", rule.ResourceType())
	assert.Equal(t, rules.Critical, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestS3EncryptionEnabled(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails without encryption config",
			planFile:     "s3_no_encryption.json",
			resourceAddr: "aws_s3_bucket.no_encryption",
			expectPass:   false,
			msgContains:  "server-side encryption",
		},
		{
			name:         "fails with empty encryption config",
			planFile:     "s3_no_encryption.json",
			resourceAddr: "aws_s3_bucket.empty_encryption",
			expectPass:   false,
			msgContains:  "server-side encryption",
		},
		{
			name:         "passes with encryption configured",
			planFile:     "s3_secure.json",
			resourceAddr: "aws_s3_bucket.secure",
			expectPass:   true,
		},
	}

	rule := NewS3EncryptionEnabledRule()
	assert.Equal(t, "aws-s3-encryption-enabled", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestS3VersioningEnabled(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails without versioning",
			planFile:     "s3_no_versioning.json",
			resourceAddr: "aws_s3_bucket.no_versioning",
			expectPass:   false,
			msgContains:  "versioning",
		},
		{
			name:         "fails with versioning disabled",
			planFile:     "s3_no_versioning.json",
			resourceAddr: "aws_s3_bucket.versioning_disabled",
			expectPass:   false,
			msgContains:  "versioning",
		},
		{
			name:         "passes with versioning enabled",
			planFile:     "s3_secure.json",
			resourceAddr: "aws_s3_bucket.secure",
			expectPass:   true,
		},
	}

	rule := NewS3VersioningEnabledRule()
	assert.Equal(t, "aws-s3-versioning-enabled", rule.ID())
	assert.Equal(t, rules.Warning, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestS3NoPublicPolicy(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when block_public_policy is false",
			planFile:     "s3_public_access_block_insecure.json",
			resourceAddr: "aws_s3_bucket_public_access_block.insecure",
			expectPass:   false,
			msgContains:  "block_public_policy",
		},
		{
			name:         "passes when block_public_policy is true",
			planFile:     "s3_secure.json",
			resourceAddr: "aws_s3_bucket_public_access_block.secure",
			expectPass:   true,
		},
	}

	rule := NewS3NoPublicPolicyRule()
	assert.Equal(t, "aws-s3-no-public-policy", rule.ID())
	assert.Equal(t, "aws_s3_bucket_public_access_block", rule.ResourceType())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

// ========================================
// IAM Rules Tests
// ========================================

func TestIAMNoWildcardAction(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails with Action:* and Resource:*",
			planFile:     "iam_wildcard_policy.json",
			resourceAddr: "aws_iam_policy.admin",
			expectPass:   false,
			msgContains:  "full admin access",
		},
		{
			name:         "passes with specific actions",
			planFile:     "iam_secure.json",
			resourceAddr: "aws_iam_policy.s3_read",
			expectPass:   true,
		},
	}

	rule := NewIAMNoWildcardActionRule()
	assert.Equal(t, "aws-iam-no-wildcard-action", rule.ID())
	assert.Equal(t, "aws_iam_policy", rule.ResourceType())
	assert.Equal(t, rules.Critical, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestIAMNoAdminPolicyRole(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails with AdministratorAccess attached to role",
			planFile:     "iam_admin_attachment.json",
			resourceAddr: "aws_iam_role_policy_attachment.admin",
			expectPass:   false,
			msgContains:  "AdministratorAccess",
		},
		{
			name:         "passes with specific policy attached to role",
			planFile:     "iam_secure.json",
			resourceAddr: "aws_iam_role_policy_attachment.s3_read",
			expectPass:   true,
		},
	}

	rule := NewIAMNoAdminPolicyRoleRule()
	assert.Equal(t, "aws-iam-no-admin-policy-role", rule.ID())
	assert.Equal(t, "aws_iam_role_policy_attachment", rule.ResourceType())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestIAMNoAdminPolicyUser(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails with AdministratorAccess attached to user",
			planFile:     "iam_admin_attachment.json",
			resourceAddr: "aws_iam_user_policy_attachment.admin",
			expectPass:   false,
			msgContains:  "AdministratorAccess",
		},
	}

	rule := NewIAMNoAdminPolicyUserRule()
	assert.Equal(t, "aws-iam-no-admin-policy-user", rule.ID())
	assert.Equal(t, "aws_iam_user_policy_attachment", rule.ResourceType())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestIAMNoUserInlinePolicy(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when inline policy attached to user",
			planFile:     "iam_user_inline_policy.json",
			resourceAddr: "aws_iam_user_policy.inline",
			expectPass:   false,
			msgContains:  "inline policy attached directly to user",
		},
	}

	rule := NewIAMNoUserInlinePolicyRule()
	assert.Equal(t, "aws-iam-no-user-inline-policy", rule.ID())
	assert.Equal(t, "aws_iam_user_policy", rule.ResourceType())
	assert.Equal(t, rules.Warning, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

// ========================================
// RDS Rules Tests
// ========================================

func TestRDSEncryptionEnabled(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when storage not encrypted",
			planFile:     "rds_no_encryption.json",
			resourceAddr: "aws_db_instance.insecure",
			expectPass:   false,
			msgContains:  "storage encryption",
		},
		{
			name:         "passes when storage encrypted",
			planFile:     "rds_secure.json",
			resourceAddr: "aws_db_instance.secure",
			expectPass:   true,
		},
	}

	rule := NewRDSEncryptionEnabledRule()
	assert.Equal(t, "aws-rds-encryption-enabled", rule.ID())
	assert.Equal(t, "aws_db_instance", rule.ResourceType())
	assert.Equal(t, rules.Critical, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestRDSNoPublicAccess(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when publicly accessible",
			planFile:     "rds_no_encryption.json",
			resourceAddr: "aws_db_instance.insecure",
			expectPass:   false,
			msgContains:  "publicly accessible",
		},
		{
			name:         "passes when not publicly accessible",
			planFile:     "rds_secure.json",
			resourceAddr: "aws_db_instance.secure",
			expectPass:   true,
		},
	}

	rule := NewRDSNoPublicAccessRule()
	assert.Equal(t, "aws-rds-no-public-access", rule.ID())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

func TestRDSBackupEnabled(t *testing.T) {
	tests := []struct {
		name         string
		planFile     string
		resourceAddr string
		expectPass   bool
		msgContains  string
	}{
		{
			name:         "fails when backup retention is 0",
			planFile:     "rds_no_encryption.json",
			resourceAddr: "aws_db_instance.insecure",
			expectPass:   false,
			msgContains:  "backup",
		},
		{
			name:         "passes when backup retention > 0",
			planFile:     "rds_secure.json",
			resourceAddr: "aws_db_instance.secure",
			expectPass:   true,
		},
	}

	rule := NewRDSBackupEnabledRule()
	assert.Equal(t, "aws-rds-backup-enabled", rule.ID())
	assert.Equal(t, rules.Warning, rule.Severity())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runRuleTest(t, rule, tc.planFile, tc.resourceAddr, tc.expectPass, tc.msgContains)
		})
	}
}

// ========================================
// Registration Tests
// ========================================

func TestRegisterAll(t *testing.T) {
	registry := rules.NewRegistry()
	RegisterAll(registry)

	allRules := registry.AllRules()

	// We should have all 22 rules registered (15 unique + duplicates for SG/SGRule variants)
	// 5 SG rules + 5 SGRule rules + 4 S3 rules + 4 IAM rules + 3 RDS rules = 21 rules
	assert.GreaterOrEqual(t, len(allRules), 21, "Expected at least 21 rules to be registered")

	// Verify some specific rules exist
	ruleIDs := make(map[string]bool)
	for _, r := range allRules {
		ruleIDs[r.ID()] = true
	}

	expectedRules := []string{
		"aws-sg-no-public-ssh",
		"aws-sg-rule-no-public-ssh",
		"aws-sg-no-public-rdp",
		"aws-sg-no-public-mysql",
		"aws-sg-no-public-postgres",
		"aws-sg-no-unrestricted-ingress",
		"aws-s3-no-public-acl",
		"aws-s3-encryption-enabled",
		"aws-s3-versioning-enabled",
		"aws-s3-no-public-policy",
		"aws-iam-no-wildcard-action",
		"aws-iam-no-admin-policy-role",
		"aws-iam-no-admin-policy-user",
		"aws-iam-no-user-inline-policy",
		"aws-rds-encryption-enabled",
		"aws-rds-no-public-access",
		"aws-rds-backup-enabled",
	}

	for _, id := range expectedRules {
		assert.True(t, ruleIDs[id], "Expected rule %s to be registered", id)
	}
}

// ========================================
// Helper Function Tests
// ========================================

func TestIsPublicCIDR(t *testing.T) {
	tests := []struct {
		cidr     string
		expected bool
	}{
		{"0.0.0.0/0", true},
		{"::/0", true},
		{"10.0.0.0/8", false},
		{"192.168.1.0/24", false},
		{"172.16.0.0/12", false},
	}

	for _, tc := range tests {
		t.Run(tc.cidr, func(t *testing.T) {
			assert.Equal(t, tc.expected, isPublicCIDR(tc.cidr))
		})
	}
}

func TestContainsPublicCIDR(t *testing.T) {
	tests := []struct {
		name     string
		cidrs    []string
		expected bool
	}{
		{"empty slice", []string{}, false},
		{"only private", []string{"10.0.0.0/8", "192.168.0.0/16"}, false},
		{"contains IPv4 public", []string{"10.0.0.0/8", "0.0.0.0/0"}, true},
		{"contains IPv6 public", []string{"10.0.0.0/8", "::/0"}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, containsPublicCIDR(tc.cidrs))
		})
	}
}

func TestParseIAMPolicy(t *testing.T) {
	policyJSON := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`

	policy, err := parseIAMPolicy(policyJSON)
	require.NoError(t, err)
	assert.Equal(t, "2012-10-17", policy.Version)
	assert.Len(t, policy.Statement, 1)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "*", policy.Statement[0].Action)
	assert.Equal(t, "*", policy.Statement[0].Resource)
}

func TestHasWildcardActionAndResource(t *testing.T) {
	tests := []struct {
		name     string
		policy   IAMPolicy
		expected bool
	}{
		{
			name: "wildcard action and resource",
			policy: IAMPolicy{
				Statement: []IAMStatement{{Effect: "Allow", Action: "*", Resource: "*"}},
			},
			expected: true,
		},
		{
			name: "wildcard action only",
			policy: IAMPolicy{
				Statement: []IAMStatement{{Effect: "Allow", Action: "*", Resource: "arn:aws:s3:::bucket"}},
			},
			expected: false,
		},
		{
			name: "specific actions",
			policy: IAMPolicy{
				Statement: []IAMStatement{{Effect: "Allow", Action: "s3:GetObject", Resource: "*"}},
			},
			expected: false,
		},
		{
			name: "deny statement with wildcards",
			policy: IAMPolicy{
				Statement: []IAMStatement{{Effect: "Deny", Action: "*", Resource: "*"}},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, hasWildcardActionAndResource(&tc.policy))
		})
	}
}

func TestPortInRange(t *testing.T) {
	tests := []struct {
		name       string
		fromPort   int
		toPort     int
		targetPort int
		expected   bool
	}{
		{"exact match", 22, 22, 22, true},
		{"in range", 20, 25, 22, true},
		{"below range", 20, 25, 19, false},
		{"above range", 20, 25, 26, false},
		{"all ports (0-0)", 0, 0, 22, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, portInRange(tc.fromPort, tc.toPort, tc.targetPort))
		})
	}
}
