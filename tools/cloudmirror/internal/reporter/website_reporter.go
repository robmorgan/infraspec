package reporter

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// WebsiteReporter generates reports for the InfraSpec website
type WebsiteReporter struct{}

// NewWebsiteReporter creates a new website reporter
func NewWebsiteReporter() *WebsiteReporter {
	return &WebsiteReporter{}
}

// GenerateWebsiteReport generates a combined report for the InfraSpec website
// It takes Virtual Cloud coverage reports and InfraSpec assertion data
func (r *WebsiteReporter) GenerateWebsiteReport(
	vcReports []*models.CoverageReport,
	infraspecOps map[string][]models.InfraSpecOperation,
) (*models.WebsiteReport, error) {
	report := &models.WebsiteReport{
		GeneratedAt: time.Now().UTC(),
		Services:    []models.ServiceSummary{},
	}

	// Track which services we've processed
	processedServices := make(map[string]bool)

	// Process Virtual Cloud reports
	for _, vcReport := range vcReports {
		serviceName := vcReport.ServiceName
		processedServices[serviceName] = true

		summary := models.ServiceSummary{
			Name:     serviceName,
			FullName: getServiceFullName(serviceName, vcReport.ServiceFullName),
			Status:   models.StatusImplemented,
		}

		// Add Virtual Cloud coverage
		summary.VirtualCloud = &models.VirtualCloudStatus{
			Status:          models.StatusImplemented,
			CoveragePercent: vcReport.CoveragePercent,
			TotalOperations: vcReport.TotalOperations,
			Implemented:     len(vcReport.Supported),
			Operations:      convertToVCOperations(vcReport),
		}

		// Add InfraSpec coverage if available
		if ops, ok := infraspecOps[serviceName]; ok && len(ops) > 0 {
			summary.InfraSpec = &models.InfraSpecCoverage{
				Status:     models.StatusImplemented,
				Operations: ops,
			}
		}

		report.Services = append(report.Services, summary)
	}

	// Add InfraSpec-only services (services with assertions but not in Virtual Cloud)
	for serviceName, ops := range infraspecOps {
		if !processedServices[serviceName] && len(ops) > 0 {
			summary := models.ServiceSummary{
				Name:     serviceName,
				FullName: getServiceFullName(serviceName, ""),
				Status:   models.StatusImplemented,
				InfraSpec: &models.InfraSpecCoverage{
					Status:     models.StatusImplemented,
					Operations: ops,
				},
			}
			report.Services = append(report.Services, summary)
			processedServices[serviceName] = true
		}
	}

	// Add planned services
	for _, planned := range models.PlannedServices {
		if !processedServices[planned.Name] {
			summary := models.ServiceSummary{
				Name:     planned.Name,
				FullName: planned.FullName,
				Status:   models.StatusPlanned,
			}
			report.Services = append(report.Services, summary)
		}
	}

	// Sort services: implemented first, then planned, alphabetically within each group
	sort.Slice(report.Services, func(i, j int) bool {
		if report.Services[i].Status != report.Services[j].Status {
			// Implemented comes before planned
			return report.Services[i].Status == models.StatusImplemented
		}
		return report.Services[i].Name < report.Services[j].Name
	})

	return report, nil
}

// GenerateJSON generates the website report as JSON
func (r *WebsiteReporter) GenerateJSON(report *models.WebsiteReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// convertToVCOperations converts coverage report operations to VirtualCloudOperations
func convertToVCOperations(report *models.CoverageReport) []models.VirtualCloudOperation {
	operations := make([]models.VirtualCloudOperation, 0, len(report.Supported)+len(report.Missing))

	// Add supported operations
	for _, op := range report.Supported {
		operations = append(operations, models.VirtualCloudOperation{
			Name:        op.Name,
			Implemented: true,
			Priority:    op.Priority,
		})
	}

	// Add missing operations (only high priority for website display)
	for _, op := range report.Missing {
		if op.Priority == models.PriorityHigh {
			operations = append(operations, models.VirtualCloudOperation{
				Name:        op.Name,
				Implemented: false,
				Priority:    op.Priority,
			})
		}
	}

	// Sort by implemented (true first), then by name
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].Implemented != operations[j].Implemented {
			return operations[i].Implemented
		}
		return operations[i].Name < operations[j].Name
	})

	return operations
}

// getServiceFullName returns the full name for a service
func getServiceFullName(serviceName, reportFullName string) string {
	if reportFullName != "" {
		return reportFullName
	}
	if fullName, ok := models.ServiceFullNames[serviceName]; ok {
		return fullName
	}
	return serviceName
}
