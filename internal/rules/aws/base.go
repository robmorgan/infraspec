// Package aws provides AWS-specific safety rules for Terraform plan evaluation.
package aws

import (
	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// BaseRule provides common fields and methods for AWS rules.
type BaseRule struct {
	id           string
	description  string
	severity     rules.Severity
	resourceType string
}

// ID returns the unique identifier for this rule.
func (r *BaseRule) ID() string {
	return r.id
}

// Description returns a human-readable description of what the rule checks.
func (r *BaseRule) Description() string {
	return r.description
}

// Severity returns the severity level of violations from this rule.
func (r *BaseRule) Severity() rules.Severity {
	return r.severity
}

// Provider returns "aws" for all AWS rules.
func (r *BaseRule) Provider() string {
	return "aws"
}

// ResourceType returns the Terraform resource type this rule applies to.
func (r *BaseRule) ResourceType() string {
	return r.resourceType
}

// passResult creates a passing result for the given resource.
func (r *BaseRule) passResult(resource *plan.ResourceChange, message string) *rules.Result {
	return &rules.Result{
		Passed:          true,
		Message:         message,
		ResourceAddress: resource.Address,
		RuleID:          r.id,
		Severity:        r.severity,
	}
}

// failResult creates a failing result for the given resource.
func (r *BaseRule) failResult(resource *plan.ResourceChange, message string) *rules.Result {
	return &rules.Result{
		Passed:          false,
		Message:         message,
		ResourceAddress: resource.Address,
		RuleID:          r.id,
		Severity:        r.severity,
	}
}
