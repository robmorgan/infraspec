package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/engine"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

var (
	// Colors
	primaryColor = lipgloss.Color("#2563eb") // Blue
	successColor = lipgloss.Color("#16a34a") // Green
	errorColor   = lipgloss.Color("#dc2626") // Red
	warningColor = lipgloss.Color("#ea580c") // Orange
	infoColor    = lipgloss.Color("#3b82f6") // Light blue
	mutedColor   = lipgloss.Color("#6b7280") // Gray

	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	summaryStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(1, 0)
)

// TextFormatter formats output as colored text for terminals
type TextFormatter struct {
	isTTY bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		isTTY: term.IsTerminal(int(os.Stdout.Fd())),
	}
}

// Format formats the check result as text
func (f *TextFormatter) Format(result Result, w io.Writer) error {
	// Header
	f.printHeader(w)

	// Stats line
	fmt.Fprintf(w, "Checking %d file(s)...\n\n", result.FilesScanned)

	// Group violations by severity
	errorViolations := f.filterBySeverity(result.Violations, rules.SeverityError)
	warningViolations := f.filterBySeverity(result.Violations, rules.SeverityWarning)
	infoViolations := f.filterBySeverity(result.Violations, rules.SeverityInfo)

	// Print violations
	if len(result.Violations) > 0 {
		fmt.Fprintln(w, "=== Violations ===")
		fmt.Fprintln(w)

		// Print errors first
		for _, v := range errorViolations {
			f.printViolation(w, v)
		}

		// Then warnings
		for _, v := range warningViolations {
			f.printViolation(w, v)
		}

		// Then info
		for _, v := range infoViolations {
			f.printViolation(w, v)
		}
	}

	// Summary
	f.printSummary(w, result, len(errorViolations), len(warningViolations), len(infoViolations))

	return nil
}

// FormatRules formats a list of rules as text
func (f *TextFormatter) FormatRules(rulesData interface{}, w io.Writer) error {
	// This is handled in the CLI for text format
	return nil
}

func (f *TextFormatter) printHeader(w io.Writer) {
	if f.isTTY {
		header := headerStyle.Render("InfraSpec Gatekeeper")
		fmt.Fprintln(w, header)
	} else {
		fmt.Fprintln(w, "InfraSpec Gatekeeper")
	}
	fmt.Fprintln(w)
}

func (f *TextFormatter) filterBySeverity(violations []engine.Violation, severity rules.Severity) []engine.Violation {
	var filtered []engine.Violation
	for _, v := range violations {
		if v.Severity == severity {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (f *TextFormatter) printViolation(w io.Writer, v engine.Violation) {
	// Severity badge
	var badge string
	var style lipgloss.Style

	switch v.Severity {
	case rules.SeverityError:
		badge = "[ERROR]"
		style = errorStyle
	case rules.SeverityWarning:
		badge = "[WARNING]"
		style = warningStyle
	case rules.SeverityInfo:
		badge = "[INFO]"
		style = infoStyle
	}

	// Print violation header
	if f.isTTY {
		fmt.Fprintf(w, "%s %s: %s\n",
			style.Render(badge),
			headerStyle.Render(v.RuleID),
			v.RuleName)
	} else {
		fmt.Fprintf(w, "%s %s: %s\n", badge, v.RuleID, v.RuleName)
	}

	// Resource info
	relPath := f.relativePath(v.File)
	if f.isTTY {
		fmt.Fprintf(w, "  %s %s.%s\n",
			mutedStyle.Render("Resource:"),
			v.ResourceType,
			v.ResourceName)
		fmt.Fprintf(w, "  %s %s:%d\n",
			mutedStyle.Render("File:"),
			relPath,
			v.Line)
	} else {
		fmt.Fprintf(w, "  Resource: %s.%s\n", v.ResourceType, v.ResourceName)
		fmt.Fprintf(w, "  File: %s:%d\n", relPath, v.Line)
	}

	// Message
	if v.Message != "" {
		if f.isTTY {
			fmt.Fprintf(w, "  %s %s\n",
				mutedStyle.Render("Message:"),
				v.Message)
		} else {
			fmt.Fprintf(w, "  Message: %s\n", v.Message)
		}
	}

	// Remediation
	if v.Remediation != "" {
		fmt.Fprintln(w)
		if f.isTTY {
			fmt.Fprintf(w, "  %s\n", infoStyle.Render("Remediation:"))
		} else {
			fmt.Fprintln(w, "  Remediation:")
		}
		for _, line := range strings.Split(v.Remediation, "\n") {
			fmt.Fprintf(w, "    %s\n", strings.TrimSpace(line))
		}
	}

	fmt.Fprintln(w)
}

func (f *TextFormatter) printSummary(w io.Writer, result Result, errors, warnings, infos int) {
	var lines []string

	lines = append(lines, "=== Summary ===")
	lines = append(lines, "")

	// Overall result
	var resultLine string
	if result.Passed {
		if f.isTTY {
			resultLine = successStyle.Render("Result: PASS")
		} else {
			resultLine = "Result: PASS"
		}
	} else {
		if f.isTTY {
			resultLine = errorStyle.Render("Result: FAIL")
		} else {
			resultLine = "Result: FAIL"
		}
	}
	lines = append(lines, resultLine)
	lines = append(lines, "")

	// Stats
	lines = append(lines, fmt.Sprintf("Files: %d | Resources: %d | Rules: %d",
		result.FilesScanned, result.ResourcesFound, result.RulesEvaluated))

	// Violation counts
	if len(result.Violations) > 0 {
		parts := []string{}
		if errors > 0 {
			if f.isTTY {
				parts = append(parts, errorStyle.Render(fmt.Sprintf("%d error(s)", errors)))
			} else {
				parts = append(parts, fmt.Sprintf("%d error(s)", errors))
			}
		}
		if warnings > 0 {
			if f.isTTY {
				parts = append(parts, warningStyle.Render(fmt.Sprintf("%d warning(s)", warnings)))
			} else {
				parts = append(parts, fmt.Sprintf("%d warning(s)", warnings))
			}
		}
		if infos > 0 {
			if f.isTTY {
				parts = append(parts, infoStyle.Render(fmt.Sprintf("%d info(s)", infos)))
			} else {
				parts = append(parts, fmt.Sprintf("%d info(s)", infos))
			}
		}
		lines = append(lines, fmt.Sprintf("Violations: %s", strings.Join(parts, ", ")))
	} else {
		lines = append(lines, "Violations: 0")
	}

	lines = append(lines, "")
	if f.isTTY {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("Duration: %s", result.Duration.Round(1e6))))
	} else {
		lines = append(lines, fmt.Sprintf("Duration: %s", result.Duration.Round(1e6)))
	}

	// Print with or without box
	if f.isTTY {
		box := summaryStyle.Render(strings.Join(lines, "\n"))
		fmt.Fprintln(w, box)
	} else {
		fmt.Fprintln(w)
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
	}
}

func (f *TextFormatter) relativePath(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return rel
}
