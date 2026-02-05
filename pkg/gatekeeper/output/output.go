// Package output provides formatters for gatekeeper check results.
package output

import (
	"fmt"
	"io"
	"time"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/engine"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

// Result contains the results to be formatted
type Result struct {
	Violations     []engine.Violation
	FilesScanned   int
	ResourcesFound int
	RulesEvaluated int
	Duration       time.Duration
	Passed         bool
}

// RuleSummary contains rule information for listing
type RuleSummary struct {
	ID          string
	Name        string
	Description string
	Severity    rules.Severity
	Tags        []string
}

// Formatter is the interface for output formatters
type Formatter interface {
	// Format formats the check result
	Format(result Result, w io.Writer) error
	// FormatRules formats a list of rules
	FormatRules(rules interface{}, w io.Writer) error
}

// NewFormatter creates a formatter based on the format string
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "text", "":
		return NewTextFormatter(), nil
	case "json":
		return NewJSONFormatter(), nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: text, json)", format)
	}
}
