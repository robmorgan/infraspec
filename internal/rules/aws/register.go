package aws

import "github.com/robmorgan/infraspec/internal/rules"

func init() {
	rules.RegisterProvider(RegisterAll)
}

// RegisterAll registers all AWS rules with the given registry.
func RegisterAll(registry *rules.Registry) {
	registry.RegisterAll(
		// Security Group rules (aws_security_group)
		NewSGNoPublicSSHRule(),
		NewSGNoPublicRDPRule(),
		NewSGNoPublicMySQLRule(),
		NewSGNoPublicPostgresRule(),
		NewSGNoUnrestrictedIngressRule(),

		// Security Group Rule rules (aws_security_group_rule)
		NewSGRuleNoPublicSSHRule(),
		NewSGRuleNoPublicRDPRule(),
		NewSGRuleNoPublicMySQLRule(),
		NewSGRuleNoPublicPostgresRule(),
		NewSGRuleNoUnrestrictedIngressRule(),

		// S3 rules
		NewS3NoPublicACLRule(),
		NewS3EncryptionEnabledRule(),
		NewS3VersioningEnabledRule(),
		NewS3NoPublicPolicyRule(),

		// IAM rules
		NewIAMNoWildcardActionRule(),
		NewIAMNoAdminPolicyRoleRule(),
		NewIAMNoAdminPolicyUserRule(),
		NewIAMNoUserInlinePolicyRule(),

		// RDS rules
		NewRDSEncryptionEnabledRule(),
		NewRDSNoPublicAccessRule(),
		NewRDSBackupEnabledRule(),
	)
}
