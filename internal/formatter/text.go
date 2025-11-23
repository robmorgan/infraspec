package formatter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

// TextFormatter implements a plain-text (Terraform-style) formatter suitable for CI logs.
type TextFormatter struct {
	writer         io.Writer
	startTime      time.Time
	currentSuite   string
	currentPickle  *messages.Pickle
	stepResults    map[string]stepResult
	totalSteps     int
	passedSteps    int
	failedSteps    int
	skippedSteps   int
	pendingSteps   int
	undefinedSteps int
}

// NewTextFormatter creates a new TextFormatter.
func NewTextFormatter(suite string, writer io.Writer) formatters.Formatter {
	return &TextFormatter{
		writer:       writer,
		currentSuite: suite,
		startTime:    time.Now(),
		stepResults:  make(map[string]stepResult),
	}
}

// TestRunStarted is called when test run begins.
func (f *TextFormatter) TestRunStarted() {
	fmt.Fprintf(f.writer, "=== InfraSpec (%s) ===\n\n", f.currentSuite)
}

// Feature prints the feature header.
func (f *TextFormatter) Feature(gherkinDocument *messages.GherkinDocument, uri string, content []byte) {
	if gherkinDocument == nil || gherkinDocument.Feature == nil {
		return
	}

	feature := gherkinDocument.Feature
	fmt.Fprintf(f.writer, "Feature: %s\n", feature.Name)
	if feature.Description != "" {
		description := strings.TrimSpace(feature.Description)
		if description != "" {
			for _, line := range strings.Split(description, "\n") {
				fmt.Fprintf(f.writer, "  %s\n", line)
			}
		}
	}
	fmt.Fprintln(f.writer)
}

// Pickle prints the scenario header.
func (f *TextFormatter) Pickle(pickle *messages.Pickle) {
	f.currentPickle = pickle
	fmt.Fprintf(f.writer, "  Scenario: %s\n", pickle.Name)
}

// Defined increments the step counter.
func (f *TextFormatter) Defined(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.totalSteps++
}

// Passed logs a passed step.
func (f *TextFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.passedSteps++
	f.recordStepResult(step, "PASS", nil)
}

// Skipped logs a skipped step.
func (f *TextFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.skippedSteps++
	f.recordStepResult(step, "SKIP", nil)
}

// Undefined logs an undefined step.
func (f *TextFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.undefinedSteps++
	f.recordStepResult(step, "UNDEFINED", nil)
}

// Pending logs a pending step.
func (f *TextFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition) {
	f.pendingSteps++
	f.recordStepResult(step, "PENDING", nil)
}

// Failed logs a failed step.
func (f *TextFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition, err error) {
	f.failedSteps++
	f.recordStepResult(step, "FAIL", err)
}

// Ambiguous logs an ambiguous step.
func (f *TextFormatter) Ambiguous(pickle *messages.Pickle, step *messages.PickleStep, stepDef *formatters.StepDefinition, err error) {
	f.failedSteps++
	f.recordStepResult(step, "AMBIGUOUS", err)
}

// Summary prints the summary after completion.
func (f *TextFormatter) Summary() {
	duration := time.Since(f.startTime).Round(time.Millisecond)
	fmt.Fprintln(f.writer)
	fmt.Fprintln(f.writer, "=== Summary ===")
	if f.failedSteps > 0 || f.undefinedSteps > 0 {
		fmt.Fprintln(f.writer, "Result: FAIL")
	} else {
		fmt.Fprintln(f.writer, "Result: PASS")
	}
	fmt.Fprintf(f.writer, "Duration: %s\n", duration)
	fmt.Fprintf(f.writer, "Total Steps: %d\n", f.totalSteps)
	if f.passedSteps > 0 {
		fmt.Fprintf(f.writer, "  Passed: %d\n", f.passedSteps)
	}
	if f.failedSteps > 0 {
		fmt.Fprintf(f.writer, "  Failed: %d\n", f.failedSteps)
	}
	if f.skippedSteps > 0 {
		fmt.Fprintf(f.writer, "  Skipped: %d\n", f.skippedSteps)
	}
	if f.pendingSteps > 0 {
		fmt.Fprintf(f.writer, "  Pending: %d\n", f.pendingSteps)
	}
	if f.undefinedSteps > 0 {
		fmt.Fprintf(f.writer, "  Undefined: %d\n", f.undefinedSteps)
	}
}

func (f *TextFormatter) recordStepResult(step *messages.PickleStep, status string, err error) {
	stepID := fmt.Sprintf("%p", step)
	if step != nil && step.Id != "" {
		stepID = step.Id
	}
	f.stepResults[stepID] = stepResult{status: status, err: err}

	stepText := ""
	if step != nil {
		stepText = step.Text
	}
	fmt.Fprintf(f.writer, "    %-9s %s\n", status+":", stepText)
	if err != nil {
		fmt.Fprintf(f.writer, "      Error: %s\n", err)
	}

	f.checkScenarioComplete()
}

func (f *TextFormatter) checkScenarioComplete() {
	if f.currentPickle == nil {
		return
	}

	completed := 0
	hasError := false

	for _, step := range f.currentPickle.Steps {
		if result, ok := f.stepResults[step.Id]; ok {
			completed++
			if result.status == "FAIL" || result.status == "AMBIGUOUS" || result.err != nil {
				hasError = true
			}
		}
	}

	if completed == len(f.currentPickle.Steps) {
		status := "PASS"
		if hasError {
			status = "FAIL"
		}
		fmt.Fprintf(f.writer, "  Scenario Result: %s\n\n", status)
		f.currentPickle = nil
	}
}

var _ formatters.Formatter = (*TextFormatter)(nil)
