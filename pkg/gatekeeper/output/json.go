package output

import (
	"encoding/json"
	"io"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/engine"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// JSONResult is the JSON output structure
type JSONResult struct {
	Passed     bool            `json:"passed"`
	Summary    JSONSummary     `json:"summary"`
	Violations []JSONViolation `json:"violations"`
}

// JSONSummary contains summary statistics
type JSONSummary struct {
	FilesScanned   int                 `json:"files_scanned"`
	ResourcesFound int                 `json:"resources_found"`
	RulesEvaluated int                 `json:"rules_evaluated"`
	DurationMs     int64               `json:"duration_ms"`
	Violations     JSONViolationCounts `json:"violations"`
}

// JSONViolationCounts contains violation counts by severity
type JSONViolationCounts struct {
	Total    int `json:"total"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

// JSONViolation represents a violation in JSON format
type JSONViolation struct {
	RuleID       string `json:"rule_id"`
	RuleName     string `json:"rule_name"`
	Severity     string `json:"severity"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Message      string `json:"message"`
	Remediation  string `json:"remediation,omitempty"`
}

// Format formats the check result as JSON
func (f *JSONFormatter) Format(result Result, w io.Writer) error {
	// Count violations by severity
	var errors, warnings, infos int
	for _, v := range result.Violations {
		switch v.Severity.String() {
		case "error":
			errors++
		case "warning":
			warnings++
		case "info":
			infos++
		}
	}

	// Convert violations
	jsonViolations := make([]JSONViolation, len(result.Violations))
	for i, v := range result.Violations {
		jsonViolations[i] = f.convertViolation(v)
	}

	output := JSONResult{
		Passed: result.Passed,
		Summary: JSONSummary{
			FilesScanned:   result.FilesScanned,
			ResourcesFound: result.ResourcesFound,
			RulesEvaluated: result.RulesEvaluated,
			DurationMs:     result.Duration.Milliseconds(),
			Violations: JSONViolationCounts{
				Total:    len(result.Violations),
				Errors:   errors,
				Warnings: warnings,
				Infos:    infos,
			},
		},
		Violations: jsonViolations,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// FormatRules formats a list of rules as JSON
func (f *JSONFormatter) FormatRules(rulesData interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(rulesData)
}

func (f *JSONFormatter) convertViolation(v engine.Violation) JSONViolation {
	return JSONViolation{
		RuleID:       v.RuleID,
		RuleName:     v.RuleName,
		Severity:     v.Severity.String(),
		ResourceType: v.ResourceType,
		ResourceName: v.ResourceName,
		File:         v.File,
		Line:         v.Line,
		Message:      v.Message,
		Remediation:  v.Remediation,
	}
}
