package runner

import (
	"fmt"
	"io"
	"os"
	"time"
)

// SummaryPrinter handles printing aggregated results.
type SummaryPrinter struct {
	writer io.Writer
}

// NewSummaryPrinter creates a new summary printer.
func NewSummaryPrinter(w io.Writer) *SummaryPrinter {
	if w == nil {
		w = os.Stdout
	}
	return &SummaryPrinter{writer: w}
}

// PrintResults prints the detailed results for each feature.
func (sp *SummaryPrinter) PrintResults(results *AggregatedResults) {
	fmt.Fprintln(sp.writer, "\n=== Results ===")

	for _, r := range results.Results {
		status := sp.formatStatus(r.Status)
		duration := r.Duration.Round(time.Millisecond)
		displayPath := shortenPath(r.FeaturePath)

		fmt.Fprintf(sp.writer, "%s %s (%s)\n", status, displayPath, duration)

		if r.Error != nil {
			// Indent error message
			fmt.Fprintf(sp.writer, "       Error: %s\n", r.Error.Error())
		}
	}
}

// PrintSummary prints the final summary statistics.
func (sp *SummaryPrinter) PrintSummary(results *AggregatedResults) {
	fmt.Fprintln(sp.writer, "\n=== Summary ===")
	fmt.Fprintf(sp.writer, "Features: %d total, %d passed, %d failed\n",
		results.TotalFeatures,
		results.PassedFeatures,
		results.FailedFeatures,
	)
	fmt.Fprintf(sp.writer, "Duration: %s\n", results.TotalDuration.Round(time.Millisecond))

	if results.FailedFeatures > 0 {
		fmt.Fprintln(sp.writer, "\nResult: FAILED")
	} else {
		fmt.Fprintln(sp.writer, "\nResult: PASSED")
	}
}

// PrintParallelSummary prints both detailed results and summary.
func (sp *SummaryPrinter) PrintParallelSummary(results *AggregatedResults) {
	sp.PrintResults(results)
	sp.PrintSummary(results)
}

// formatStatus returns a formatted status string for results.
func (sp *SummaryPrinter) formatStatus(status FeatureStatus) string {
	switch status {
	case StatusPassed:
		return "[PASS]"
	case StatusFailed:
		return "[FAIL]"
	case StatusTimeout:
		return "[TIMEOUT]"
	case StatusCanceled:
		return "[CANCELED]"
	default:
		return "[UNKNOWN]"
	}
}

// PrintParallelResults is a convenience function to print parallel execution results.
func PrintParallelResults(results *AggregatedResults) {
	printer := NewSummaryPrinter(os.Stdout)
	printer.PrintParallelSummary(results)
}
