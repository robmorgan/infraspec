// Package check provides the check command implementation for evaluating Terraform plans against rules.
package check

import (
	"fmt"
	"strings"

	"github.com/robmorgan/infraspec/internal/rules"
)

// Options configures how the check command runs.
type Options struct {
	// PlanPath is the path to a Terraform plan JSON file.
	// Mutually exclusive with Dir.
	PlanPath string

	// Dir is the path to a Terraform directory to run plan against.
	// Defaults to current directory if neither PlanPath nor Dir is set.
	Dir string

	// MinSeverity is the minimum severity threshold for reporting violations.
	// Rules with severity below this threshold are skipped.
	MinSeverity rules.Severity

	// IgnoreRuleIDs is a list of rule IDs to skip during evaluation.
	IgnoreRuleIDs []string

	// Format is the output format: text, json.
	Format string
}

// Result represents the outcome of a single rule check.
type Result struct {
	RuleID          string         `json:"rule_id"`
	RuleDescription string         `json:"rule_description"`
	ResourceAddress string         `json:"resource_address"`
	ResourceType    string         `json:"resource_type"`
	Passed          bool           `json:"passed"`
	Message         string         `json:"message"`
	Severity        rules.Severity `json:"severity"`
	SeverityString  string         `json:"severity_string"`
}

// Summary aggregates all check results and provides summary statistics.
type Summary struct {
	Results        []Result `json:"results"`
	TotalChecks    int      `json:"total_checks"`
	Passed         int      `json:"passed"`
	Failed         int      `json:"failed"`
	CriticalFailed int      `json:"critical_failed"`
	WarningFailed  int      `json:"warning_failed"`
	InfoFailed     int      `json:"info_failed"`
	Skipped        int      `json:"skipped"`
	ExitCode       int      `json:"exit_code"`
}

// SeverityFromString parses a severity string and returns the corresponding Severity value.
// Valid values are: "critical", "warning", "info" (case-insensitive).
func SeverityFromString(s string) (rules.Severity, error) {
	switch strings.ToLower(s) {
	case "critical":
		return rules.Critical, nil
	case "warning":
		return rules.Warning, nil
	case "info":
		return rules.Info, nil
	default:
		return rules.Info, fmt.Errorf("invalid severity %q: must be critical, warning, or info", s)
	}
}
