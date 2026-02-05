package aws

import (
	"strings"

	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// IAMNoWildcardActionRule checks that IAM policies don't have Action:"*" with Resource:"*".
type IAMNoWildcardActionRule struct {
	BaseRule
}

// NewIAMNoWildcardActionRule creates a new rule that checks for wildcard actions.
func NewIAMNoWildcardActionRule() *IAMNoWildcardActionRule {
	return &IAMNoWildcardActionRule{
		BaseRule: BaseRule{
			id:           "aws-iam-no-wildcard-action",
			description:  "IAM policies should not have Action:\"*\" with Resource:\"*\" (full admin access)",
			severity:     rules.Critical,
			resourceType: "aws_iam_policy",
		},
	}
}

// Check evaluates the rule against an aws_iam_policy resource.
func (r *IAMNoWildcardActionRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	policyDoc := resource.GetAfterString("policy")
	if policyDoc == "" {
		return r.passResult(resource, "IAM policy document is empty or not defined"), nil
	}

	policy, parseErr := parseIAMPolicy(policyDoc)
	if parseErr != nil {
		// If we can't parse the policy, we can't evaluate it - pass with warning
		// We intentionally return nil error here because a parse failure is not a rule check failure
		return r.passResult(resource, "Could not parse IAM policy document"), nil //nolint:nilerr
	}

	if hasWildcardActionAndResource(policy) {
		return r.failResult(resource, "IAM policy grants full admin access (Action:\"*\" with Resource:\"*\"); use specific actions and resources"), nil
	}

	return r.passResult(resource, "IAM policy does not grant full admin access"), nil
}

// IAMNoAdminPolicyRule checks that IAM role/user policy attachments don't use AdministratorAccess.
type IAMNoAdminPolicyRule struct {
	BaseRule
}

// NewIAMNoAdminPolicyRoleRule creates a new rule for role policy attachments.
func NewIAMNoAdminPolicyRoleRule() *IAMNoAdminPolicyRule {
	return &IAMNoAdminPolicyRule{
		BaseRule: BaseRule{
			id:           "aws-iam-no-admin-policy-role",
			description:  "IAM role policy attachments should not use AdministratorAccess managed policy",
			severity:     rules.Critical,
			resourceType: "aws_iam_role_policy_attachment",
		},
	}
}

// NewIAMNoAdminPolicyUserRule creates a new rule for user policy attachments.
func NewIAMNoAdminPolicyUserRule() *IAMNoAdminPolicyRule {
	return &IAMNoAdminPolicyRule{
		BaseRule: BaseRule{
			id:           "aws-iam-no-admin-policy-user",
			description:  "IAM user policy attachments should not use AdministratorAccess managed policy",
			severity:     rules.Critical,
			resourceType: "aws_iam_user_policy_attachment",
		},
	}
}

// Check evaluates the rule against an IAM policy attachment resource.
func (r *IAMNoAdminPolicyRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	policyARN := resource.GetAfterString("policy_arn")
	if policyARN == "" {
		return r.passResult(resource, "Policy ARN is empty or not defined"), nil
	}

	// Check for AWS managed AdministratorAccess policy
	if strings.Contains(policyARN, ":policy/AdministratorAccess") {
		return r.failResult(resource, "IAM policy attachment uses AdministratorAccess managed policy; use more restrictive policies"), nil
	}

	return r.passResult(resource, "IAM policy attachment does not use AdministratorAccess"), nil
}

// IAMNoUserInlinePolicyRule checks that inline policies are not attached directly to users.
type IAMNoUserInlinePolicyRule struct {
	BaseRule
}

// NewIAMNoUserInlinePolicyRule creates a new rule that checks for user inline policies.
func NewIAMNoUserInlinePolicyRule() *IAMNoUserInlinePolicyRule {
	return &IAMNoUserInlinePolicyRule{
		BaseRule: BaseRule{
			id:           "aws-iam-no-user-inline-policy",
			description:  "IAM inline policies should not be attached directly to users; use groups or roles instead",
			severity:     rules.Warning,
			resourceType: "aws_iam_user_policy",
		},
	}
}

// Check evaluates the rule against an aws_iam_user_policy resource.
// The existence of this resource type is itself a violation.
func (r *IAMNoUserInlinePolicyRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	// The mere existence of an aws_iam_user_policy resource is a violation
	userName := resource.GetAfterString("user")
	if userName == "" {
		userName = resource.GetAfterString("name")
	}

	msg := "IAM inline policy attached directly to user"
	if userName != "" {
		msg = "IAM inline policy attached directly to user '" + userName + "'"
	}
	msg += "; attach policies to groups or roles instead for better management"

	return r.failResult(resource, msg), nil
}
