package reporter

import (
	"encoding/json"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// JSONReporter generates JSON reports
type JSONReporter struct{}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

// GenerateReport generates a JSON report for a single service
func (r *JSONReporter) GenerateReport(report *models.CoverageReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// GenerateMultiReport generates a JSON report for multiple services
func (r *JSONReporter) GenerateMultiReport(reports []*models.CoverageReport) ([]byte, error) {
	return json.MarshalIndent(reports, "", "  ")
}

// GenerateDiffReport generates a JSON report for version differences
func (r *JSONReporter) GenerateDiffReport(diff *models.VersionDiff) ([]byte, error) {
	return json.MarshalIndent(diff, "", "  ")
}

// GenerateSummaryReport generates a JSON summary report
func (r *JSONReporter) GenerateSummaryReport(multi *models.MultiServiceReport) ([]byte, error) {
	return json.MarshalIndent(multi, "", "  ")
}
