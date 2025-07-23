package formatter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

var (
	// Colors and styles similar to Claude Code
	primaryColor    = lipgloss.Color("#2563eb") // Blue
	successColor    = lipgloss.Color("#16a34a") // Green
	errorColor      = lipgloss.Color("#dc2626") // Red
	warningColor    = lipgloss.Color("#ea580c") // Orange
	mutedColor      = lipgloss.Color("#6b7280") // Gray
	backgroundColor = lipgloss.Color("#f8fafc") // Light gray background

	// Style definitions
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	stepStyle = lipgloss.NewStyle().
			Padding(0, 2)

	scenarioStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	featureStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 0)

	progressBarStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("#ffffff"))

	summaryStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(1, 0)
)

// TUIFormatter implements a Claude Code-like TUI formatter for Godog
type TUIFormatter struct {
	writer         io.Writer
	startTime      time.Time
	currentSuite   string
	features       []*messages.GherkinDocument
	pickles        []*messages.Pickle
	stepResults    map[string]stepResult
	totalSteps     int
	passedSteps    int
	failedSteps    int
	skippedSteps   int
	pendingSteps   int
	undefinedSteps int
	currentPickle  *messages.Pickle
}

type stepResult struct {
	status string
	err    error
}

// NewTUIFormatter creates a new TUI formatter
func NewTUIFormatter(suiteName string, writer io.Writer) formatters.Formatter {
	return &TUIFormatter{
		writer:       writer,
		currentSuite: suiteName,
		stepResults:  make(map[string]stepResult),
		startTime:    time.Now(),
	}
}

// TestRunStarted is called when the test run starts
func (f *TUIFormatter) TestRunStarted() {
	f.clearScreen()
	f.printHeader()
}

// Feature is called when a feature is encountered
func (f *TUIFormatter) Feature(gherkinDocument *messages.GherkinDocument, uri string, content []byte) {
	f.features = append(f.features, gherkinDocument)

	feature := gherkinDocument.Feature
	if feature == nil {
		return
	}

	featureHeader := featureStyle.Render(fmt.Sprintf("📋 Feature: %s", feature.Name))
	fmt.Fprintf(f.writer, "%s\n", featureHeader)

	if feature.Description != "" {
		description := mutedStyle.Render(fmt.Sprintf("   %s", strings.TrimSpace(feature.Description)))
		fmt.Fprintf(f.writer, "%s\n\n", description)
	}
}

// Pickle is called when a scenario starts
func (f *TUIFormatter) Pickle(pickle *messages.Pickle) {
	f.pickles = append(f.pickles, pickle)
	f.currentPickle = pickle

	scenarioHeader := scenarioStyle.Render(fmt.Sprintf("🎯 Scenario: %s", pickle.Name))
	fmt.Fprintf(f.writer, "%s\n", scenarioHeader)
}

// Defined is called when a step is defined and ready to execute
func (f *TUIFormatter) Defined(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.totalSteps++

	// Show step as running
	stepText := stepStyle.Render(fmt.Sprintf("   %s %s", f.getSpinner(), step.Text))
	fmt.Fprintf(f.writer, "%s", stepText)
}

// Passed is called when a step passes
func (f *TUIFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.passedSteps++
	f.stepResults[step.Text] = stepResult{status: "PASSED", err: nil}

	f.printStepResult(step, "✅", "PASSED", successStyle, nil)
}

// Failed is called when a step fails
func (f *TUIFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition, err error) {
	f.failedSteps++
	f.stepResults[step.Text] = stepResult{status: "FAILED", err: err}

	f.printStepResult(step, "❌", "FAILED", errorStyle, err)
}

// Skipped is called when a step is skipped
func (f *TUIFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.skippedSteps++
	f.stepResults[step.Text] = stepResult{status: "SKIPPED", err: nil}

	f.printStepResult(step, "⏭️", "SKIPPED", warningStyle, nil)
}

// Undefined is called when a step is undefined
func (f *TUIFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.undefinedSteps++
	f.stepResults[step.Text] = stepResult{status: "UNDEFINED", err: nil}

	f.printStepResult(step, "❓", "UNDEFINED", mutedStyle, nil)
}

// Pending is called when a step is pending
func (f *TUIFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.pendingSteps++
	f.stepResults[step.Text] = stepResult{status: "PENDING", err: nil}

	f.printStepResult(step, "⏸️", "PENDING", warningStyle, nil)
}

// Ambiguous is called when a step is ambiguous
func (f *TUIFormatter) Ambiguous(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition, err error) {
	f.failedSteps++
	f.stepResults[step.Text] = stepResult{status: "AMBIGUOUS", err: err}

	f.printStepResult(step, "❗", "AMBIGUOUS", errorStyle, err)
}

// Summary is called when the test run finishes
func (f *TUIFormatter) Summary() {
	duration := time.Since(f.startTime)
	f.printSummary(duration)
}

// Helper methods

func (f *TUIFormatter) printStepResult(step *messages.PickleStep, icon, statusText string, style lipgloss.Style, err error) {
	// Clear the current line and rewrite with result
	fmt.Fprintf(f.writer, "\r")

	stepText := stepStyle.Render(fmt.Sprintf("   %s %s", icon, step.Text))
	statusBadge := style.Render(fmt.Sprintf(" [%s]", statusText))

	fmt.Fprintf(f.writer, "%s%s\n", stepText, statusBadge)

	// Show error details if failed
	if err != nil {
		errorDetails := errorStyle.Render(fmt.Sprintf("      Error: %s", err.Error()))
		fmt.Fprintf(f.writer, "%s\n", errorDetails)
	}

	// Print scenario completion if this is the last step
	f.checkScenarioComplete()
}

func (f *TUIFormatter) checkScenarioComplete() {
	if f.currentPickle == nil {
		return
	}

	// Check if we have results for all steps in the current scenario
	completed := 0
	hasError := false

	for _, step := range f.currentPickle.Steps {
		if result, exists := f.stepResults[step.Text]; exists {
			completed++
			if result.err != nil || result.status == "FAILED" || result.status == "AMBIGUOUS" {
				hasError = true
			}
		}
	}

	// If all steps are complete, show scenario result
	if completed == len(f.currentPickle.Steps) {
		var icon, statusText string
		var style lipgloss.Style

		if hasError {
			icon = "❌"
			statusText = "FAILED"
			style = errorStyle
		} else {
			icon = "✅"
			statusText = "PASSED"
			style = successStyle
		}

		scenarioResult := style.Render(fmt.Sprintf("   %s Scenario %s", icon, statusText))
		fmt.Fprintf(f.writer, "%s\n\n", scenarioResult)

		f.currentPickle = nil // Reset for next scenario
	}
}

func (f *TUIFormatter) clearScreen() {
	fmt.Fprintf(f.writer, "\033[2J\033[H")
}

func (f *TUIFormatter) printHeader() {
	header := headerStyle.Render("🚀 InfraSpec Test Runner")
	subtitle := mutedStyle.Render("Testing your cloud infrastructure with confidence")

	fmt.Fprintf(f.writer, "%s\n", header)
	fmt.Fprintf(f.writer, "%s\n\n", subtitle)

	// Progress bar area (will be updated during execution)
	f.printProgressBar()
}

func (f *TUIFormatter) printProgressBar() {
	if f.totalSteps == 0 {
		return
	}

	completed := f.passedSteps + f.failedSteps + f.skippedSteps + f.pendingSteps + f.undefinedSteps
	progress := float64(completed) / float64(f.totalSteps)
	barWidth := 40
	filledWidth := int(progress * float64(barWidth))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)
	progressText := fmt.Sprintf(" %d/%d steps", completed, f.totalSteps)

	progressBar := progressBarStyle.Render(bar) + progressText
	fmt.Fprintf(f.writer, "%s\n\n", progressBar)
}

func (f *TUIFormatter) printSummary(duration time.Duration) {
	var summaryLines []string

	// Test results summary
	summaryLines = append(summaryLines, headerStyle.Render("📊 Test Summary"))
	summaryLines = append(summaryLines, "")

	// Overall status
	if f.failedSteps > 0 || f.undefinedSteps > 0 {
		summaryLines = append(summaryLines, errorStyle.Render("❌ Tests FAILED"))
	} else {
		summaryLines = append(summaryLines, successStyle.Render("✅ All tests PASSED"))
	}

	summaryLines = append(summaryLines, "")

	// Step statistics
	summaryLines = append(summaryLines, fmt.Sprintf("Total Steps: %d", f.totalSteps))
	if f.passedSteps > 0 {
		summaryLines = append(summaryLines, successStyle.Render(fmt.Sprintf("✅ Passed: %d", f.passedSteps)))
	}
	if f.failedSteps > 0 {
		summaryLines = append(summaryLines, errorStyle.Render(fmt.Sprintf("❌ Failed: %d", f.failedSteps)))
	}
	if f.skippedSteps > 0 {
		summaryLines = append(summaryLines, warningStyle.Render(fmt.Sprintf("⏭️  Skipped: %d", f.skippedSteps)))
	}
	if f.pendingSteps > 0 {
		summaryLines = append(summaryLines, warningStyle.Render(fmt.Sprintf("⏸️  Pending: %d", f.pendingSteps)))
	}
	if f.undefinedSteps > 0 {
		summaryLines = append(summaryLines, mutedStyle.Render(fmt.Sprintf("❓ Undefined: %d", f.undefinedSteps)))
	}

	summaryLines = append(summaryLines, "")
	summaryLines = append(summaryLines, mutedStyle.Render(fmt.Sprintf("⏱️  Duration: %s", duration.Round(time.Millisecond))))

	summary := summaryStyle.Render(strings.Join(summaryLines, "\n"))
	fmt.Fprintf(f.writer, "%s\n", summary)
}

func (f *TUIFormatter) getSpinner() string {
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinnerFrames[time.Now().UnixMilli()/100%int64(len(spinnerFrames))]
}

// Ensure TUIFormatter implements the Formatter interface
var _ formatters.Formatter = (*TUIFormatter)(nil)
