// Package rules provides rule definitions and loading for the gatekeeper.
package rules

// Severity represents the severity level of a rule
type Severity int

const (
	// SeverityError is the highest severity, blocks deployment
	SeverityError Severity = iota
	// SeverityWarning is a warning that should be reviewed
	SeverityWarning
	// SeverityInfo is informational only
	SeverityInfo
)

// String returns the string representation of a severity
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// ParseSeverity parses a severity string into a Severity value
func ParseSeverity(s string) Severity {
	switch s {
	case "error":
		return SeverityError
	case "warning", "warn":
		return SeverityWarning
	case "info":
		return SeverityInfo
	default:
		return SeverityError
	}
}

// Rule represents a security or policy rule
type Rule struct {
	ID           string    `yaml:"id"`
	Name         string    `yaml:"name"`
	Description  string    `yaml:"description,omitempty"`
	Severity     Severity  `yaml:"-"`
	SeverityStr  string    `yaml:"severity"`
	ResourceType string    `yaml:"resource_type"`
	Condition    Condition `yaml:"condition"`
	Message      string    `yaml:"message"`
	Remediation  string    `yaml:"remediation,omitempty"`
	Tags         []string  `yaml:"tags,omitempty"`
}

// RuleSet represents a collection of rules loaded from a file
type RuleSet struct {
	Version  string   `yaml:"version"`
	Metadata Metadata `yaml:"metadata,omitempty"`
	Rules    []Rule   `yaml:"rules"`
}

// Metadata contains optional metadata about a rule set
type Metadata struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Author      string `yaml:"author,omitempty"`
}
