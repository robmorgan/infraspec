package check

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/robmorgan/infraspec/internal/rules"
)

// Color styles for terminal output.
var (
	passStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Bold(true)
	failStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#dc2626")).Bold(true)
	warnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ea580c")).Bold(true)
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#3b82f6"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2563eb")).Bold(true)
	resourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b5cf6"))
)

// TextFormatter outputs check results as colorized text for terminal display.
type TextFormatter struct{}

// Format writes the check summary as colorized text.
func (f *TextFormatter) Format(w io.Writer, summary *Summary) error {
	useColors := f.isTTY(w)

	// Print header
	header := "InfraSpec - Pre-flight Check"
	if useColors {
		header = headerStyle.Render(header)
	}
	fmt.Fprintln(w, header)
	fmt.Fprintln(w)

	// Sort results: failures first, then by severity (critical > warning > info), then by rule ID
	sorted := make([]Result, len(summary.Results))
	copy(sorted, summary.Results)
	sort.Slice(sorted, func(i, j int) bool {
		// Failures first
		if sorted[i].Passed != sorted[j].Passed {
			return !sorted[i].Passed
		}
		// Higher severity first
		if sorted[i].Severity != sorted[j].Severity {
			return sorted[i].Severity > sorted[j].Severity
		}
		// Then by rule ID
		return sorted[i].RuleID < sorted[j].RuleID
	})

	// Print each result
	for i := range sorted {
		f.printResult(w, &sorted[i], useColors)
	}

	// Print summary
	fmt.Fprintln(w)
	f.printSummary(w, summary, useColors)

	return nil
}

// printResult prints a single check result.
func (f *TextFormatter) printResult(w io.Writer, result *Result, useColors bool) {
	icon := f.formatStatusIcon(result.Passed, useColors)
	severity := f.formatSeverity(result.Severity, useColors)
	ruleID := f.formatRuleID(result.RuleID, useColors)

	fmt.Fprintf(w, "%s %s %s: %s\n", icon, severity, ruleID, result.Message)

	if !result.Passed {
		f.printResourceAddress(w, result.ResourceAddress, useColors)
	}
}

// formatStatusIcon returns the pass/fail icon with optional coloring.
func (f *TextFormatter) formatStatusIcon(passed, useColors bool) string {
	if passed {
		icon := "[PASS]"
		if useColors {
			return passStyle.Render(icon)
		}
		return icon
	}
	icon := "[FAIL]"
	if useColors {
		return failStyle.Render(icon)
	}
	return icon
}

// formatRuleID returns the rule ID with optional coloring.
func (f *TextFormatter) formatRuleID(ruleID string, useColors bool) string {
	if useColors {
		return mutedStyle.Render(ruleID)
	}
	return ruleID
}

// printResourceAddress prints the resource address on a new line.
func (f *TextFormatter) printResourceAddress(w io.Writer, addr string, useColors bool) {
	if useColors {
		addr = resourceStyle.Render(addr)
	}
	fmt.Fprintf(w, "     Resource: %s\n", addr)
}

// formatSeverity returns a formatted severity string.
func (f *TextFormatter) formatSeverity(sev rules.Severity, useColors bool) string {
	label := fmt.Sprintf("[%s]", sev.String())
	if !useColors {
		return label
	}

	switch sev {
	case rules.Critical:
		return failStyle.Render(label)
	case rules.Warning:
		return warnStyle.Render(label)
	case rules.Info:
		return infoStyle.Render(label)
	default:
		return label
	}
}

// printSummary prints the final summary line.
func (f *TextFormatter) printSummary(w io.Writer, summary *Summary, useColors bool) {
	passedText := f.formatCount(summary.Passed, "passed", &passStyle, useColors)
	failedText := f.formatCount(summary.Failed, "failed", &failStyle, useColors)

	fmt.Fprintf(w, "Summary: %s, %s", passedText, failedText)

	if summary.Failed > 0 {
		f.printFailureBreakdown(w, summary, useColors)
	}

	if summary.Skipped > 0 {
		skippedText := fmt.Sprintf(", %d skipped", summary.Skipped)
		if useColors {
			skippedText = mutedStyle.Render(skippedText)
		}
		fmt.Fprint(w, skippedText)
	}

	fmt.Fprintln(w)
}

// formatCount formats a count with label and optional styling.
func (f *TextFormatter) formatCount(count int, label string, style *lipgloss.Style, useColors bool) string {
	text := fmt.Sprintf("%d %s", count, label)
	if useColors && count > 0 {
		return style.Render(text)
	}
	return text
}

// printFailureBreakdown prints the breakdown of failures by severity.
func (f *TextFormatter) printFailureBreakdown(w io.Writer, summary *Summary, useColors bool) {
	parts := f.buildFailureParts(summary, useColors)
	if len(parts) == 0 {
		return
	}

	fmt.Fprint(w, " (")
	for i, part := range parts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, part)
	}
	fmt.Fprint(w, ")")
}

// buildFailureParts builds the list of failure count parts by severity.
func (f *TextFormatter) buildFailureParts(summary *Summary, useColors bool) []string {
	var parts []string
	if summary.CriticalFailed > 0 {
		parts = append(parts, f.formatCount(summary.CriticalFailed, "critical", &failStyle, useColors))
	}
	if summary.WarningFailed > 0 {
		parts = append(parts, f.formatCount(summary.WarningFailed, "warning", &warnStyle, useColors))
	}
	if summary.InfoFailed > 0 {
		parts = append(parts, f.formatCount(summary.InfoFailed, "info", &infoStyle, useColors))
	}
	return parts
}

// isTTY checks if the writer is a terminal that supports colors.
func (f *TextFormatter) isTTY(w io.Writer) bool {
	if file, ok := w.(*os.File); ok {
		return term.IsTerminal(int(file.Fd()))
	}
	return false
}
