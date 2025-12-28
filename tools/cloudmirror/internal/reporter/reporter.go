package reporter

import (
	"fmt"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// OutputFormat represents the output format for reports
type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
	FormatBadge    OutputFormat = "badge"
)

// Reporter is the main reporter that delegates to format-specific reporters
type Reporter struct {
	json     *JSONReporter
	markdown *MarkdownReporter
	badge    *BadgeGenerator
}

// NewReporter creates a new reporter
func NewReporter() *Reporter {
	return &Reporter{
		json:     NewJSONReporter(),
		markdown: NewMarkdownReporter(),
		badge:    NewBadgeGenerator(),
	}
}

// GenerateReport generates a report in the specified format
func (r *Reporter) GenerateReport(report *models.CoverageReport, format OutputFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return r.json.GenerateReport(report)
	case FormatMarkdown:
		return r.markdown.GenerateReport(report)
	case FormatBadge:
		return r.badge.GenerateBadge(report)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GenerateMultiReport generates a multi-service report in the specified format
func (r *Reporter) GenerateMultiReport(reports []*models.CoverageReport, format OutputFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return r.json.GenerateMultiReport(reports)
	case FormatMarkdown:
		multi := &models.MultiServiceReport{
			Services: reports,
		}
		// Calculate totals
		for _, rep := range reports {
			multi.TotalOperations += rep.TotalOperations
			multi.TotalSupported += len(rep.Supported)
			multi.TotalMissing += len(rep.Missing)
		}
		if multi.TotalOperations > 0 {
			multi.OverallCoverage = float64(multi.TotalSupported) / float64(multi.TotalOperations) * 100
		}
		return r.markdown.GenerateMultiReport(multi)
	case FormatBadge:
		// Generate overall badge
		multi := &models.MultiServiceReport{
			Services: reports,
		}
		for _, rep := range reports {
			multi.TotalOperations += rep.TotalOperations
			multi.TotalSupported += len(rep.Supported)
		}
		if multi.TotalOperations > 0 {
			multi.OverallCoverage = float64(multi.TotalSupported) / float64(multi.TotalOperations) * 100
		}
		return r.badge.GenerateOverallBadge(multi)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GenerateDiffReport generates a diff report in the specified format
func (r *Reporter) GenerateDiffReport(diff *models.VersionDiff, format OutputFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return r.json.GenerateDiffReport(diff)
	case FormatMarkdown:
		return r.markdown.GenerateDiffReport(diff)
	default:
		return nil, fmt.Errorf("unsupported format for diff: %s", format)
	}
}
