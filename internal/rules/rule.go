// Package rules provides a rule interface and registry for evaluating Terraform plan resources.
package rules

import "github.com/robmorgan/infraspec/internal/plan"

// Severity levels for rule violations.
type Severity int

const (
	Info Severity = iota
	Warning
	Critical
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Critical:
		return "critical"
	default:
		return "unknown"
	}
}

// Rule interface for all plan evaluation rules.
type Rule interface {
	// ID returns the unique identifier for this rule.
	ID() string
	// Description returns a human-readable description of what the rule checks.
	Description() string
	// Severity returns the severity level of violations from this rule.
	Severity() Severity
	// Provider returns the cloud provider this rule applies to (e.g., "aws", "gcp").
	Provider() string
	// ResourceType returns the Terraform resource type this rule applies to (e.g., "aws_security_group").
	ResourceType() string
	// Check evaluates the rule against a resource change and returns the result.
	Check(resource *plan.ResourceChange) (*Result, error)
}

// Result of a rule evaluation.
type Result struct {
	Passed          bool
	Message         string
	ResourceAddress string
	RuleID          string
	Severity        Severity
}
