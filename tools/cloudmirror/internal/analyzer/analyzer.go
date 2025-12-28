package analyzer

import (
	"fmt"
	"os"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// Analyzer orchestrates the analysis of AWS API parity
type Analyzer struct {
	modelParser *AWSModelParser
	codeScanner *CodeScanner
	diffEngine  *DiffEngine
}

// NewAnalyzer creates a new analyzer with the given paths
func NewAnalyzer(sdkPath, servicesPath string) *Analyzer {
	return &Analyzer{
		modelParser: NewAWSModelParser(sdkPath),
		codeScanner: NewCodeScanner(servicesPath),
		diffEngine:  NewDiffEngine(),
	}
}

// AnalyzeService analyzes a single service for API parity
func (a *Analyzer) AnalyzeService(serviceName string) (*models.CoverageReport, error) {
	// Parse AWS model
	awsModel, err := a.modelParser.ParseService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AWS model for %s: %w", serviceName, err)
	}

	// Scan implementation
	impl, err := a.codeScanner.ScanService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to scan implementation for %s: %w", serviceName, err)
	}

	// Compare and generate report
	report := a.diffEngine.Compare(awsModel, impl)
	return report, nil
}

// AnalyzeAllServices analyzes all implemented services
func (a *Analyzer) AnalyzeAllServices() ([]*models.CoverageReport, error) {
	// Get list of implemented services
	serviceNames, err := a.codeScanner.GetImplementedServiceNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get implemented services: %w", err)
	}

	var reports []*models.CoverageReport
	var errors []string

	for _, name := range serviceNames {
		report, err := a.AnalyzeService(name)
		if err != nil {
			// Log individual service failures to stderr for visibility
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze service %s: %v\n", name, err)
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		reports = append(reports, report)
	}

	if len(errors) > 0 && len(reports) == 0 {
		return nil, fmt.Errorf("all services failed to analyze: %v", errors)
	}

	return reports, nil
}

// ListAWSServices lists available AWS services in the SDK
func (a *Analyzer) ListAWSServices() ([]string, error) {
	return a.modelParser.ListServices()
}

// ListImplementedServices lists services implemented in InfraSpec
func (a *Analyzer) ListImplementedServices() ([]string, error) {
	return a.codeScanner.GetImplementedServiceNames()
}

// CompareVersions compares two AWS model versions
func (a *Analyzer) CompareVersions(serviceName, oldSDKPath, newSDKPath string) (*models.VersionDiff, error) {
	oldParser := NewAWSModelParser(oldSDKPath)
	newParser := NewAWSModelParser(newSDKPath)

	oldModel, err := oldParser.ParseService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse old model: %w", err)
	}

	newModel, err := newParser.ParseService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse new model: %w", err)
	}

	return a.diffEngine.CompareTwoVersions(oldModel, newModel), nil
}

// GetMultiServiceReport creates a summary report for multiple services
func (a *Analyzer) GetMultiServiceReport(reports []*models.CoverageReport) *models.MultiServiceReport {
	return a.diffEngine.CreateMultiServiceReport(reports)
}

// GetAWSModel returns the parsed AWS model for a service (for scaffold generation)
func (a *Analyzer) GetAWSModel(serviceName string) (*models.AWSService, error) {
	return a.modelParser.ParseService(serviceName)
}

// CompareSDKVersions compares two SDK versions and generates a comprehensive change report
func (a *Analyzer) CompareSDKVersions(config *models.SDKCompareConfig) (*models.SDKChangeReport, error) {
	oldParser := NewAWSModelParser(config.OldSDKPath)
	newParser := NewAWSModelParser(config.NewSDKPath)

	// Get list of services to compare
	var servicesToCompare []string
	if len(config.Services) > 0 {
		servicesToCompare = config.Services
	} else {
		// Get union of services from both SDKs
		oldServices, _ := oldParser.ListServices()
		newServices, _ := newParser.ListServices()

		serviceSet := make(map[string]bool)
		for _, s := range oldServices {
			serviceSet[s] = true
		}
		for _, s := range newServices {
			serviceSet[s] = true
		}
		for s := range serviceSet {
			servicesToCompare = append(servicesToCompare, s)
		}
	}

	// Parse all service models
	oldModels := make(map[string]*models.AWSService)
	newModels := make(map[string]*models.AWSService)

	for _, serviceName := range servicesToCompare {
		oldModel, err := oldParser.ParseService(serviceName)
		if err == nil {
			oldModels[serviceName] = oldModel
		}

		newModel, err := newParser.ParseService(serviceName)
		if err == nil {
			newModels[serviceName] = newModel
		}
	}

	// Generate comparison report
	return a.diffEngine.CompareSDKVersions(oldModels, newModels, config), nil
}

// GetImplementationImpact analyzes how SDK changes affect current implementations
func (a *Analyzer) GetImplementationImpact(changeReport *models.SDKChangeReport) ([]models.ImplementationImpact, error) {
	var impacts []models.ImplementationImpact

	// Get implemented services
	implServices, err := a.codeScanner.GetImplementedServiceNames()
	if err != nil {
		return nil, err
	}
	implSet := make(map[string]bool)
	for _, s := range implServices {
		implSet[s] = true
	}

	// Check each service change
	for _, svcChange := range changeReport.Services {
		isImplemented := implSet[svcChange.Name]

		// Check new operations
		for _, newOp := range svcChange.NewOperations {
			impacts = append(impacts, models.ImplementationImpact{
				Service:              svcChange.Name,
				Operation:            newOp.Name,
				CurrentlyImplemented: false,
				RequiresNewHandler:   true,
				EstimatedComplexity:  estimateComplexity(newOp),
				SuggestedAction:      "Implement new handler for " + newOp.Name,
			})
		}

		// Check modified operations (only if service is implemented)
		if isImplemented {
			for _, modOp := range svcChange.ModifiedOps {
				// Scan for existing handler
				impl, _ := a.codeScanner.ScanService(svcChange.Name)
				var affectedFiles []string
				if impl != nil {
					if op, exists := impl.Operations[modOp.Name]; exists {
						affectedFiles = []string{op.File}
					}
				}

				impacts = append(impacts, models.ImplementationImpact{
					Service:              svcChange.Name,
					Operation:            modOp.Name,
					CurrentlyImplemented: len(affectedFiles) > 0,
					RequiresUpdate:       modOp.IsBreaking || len(modOp.NewParams) > 0,
					AffectedFiles:        affectedFiles,
					SuggestedAction:      generateSuggestedAction(modOp),
				})
			}

			// Check removed operations
			for _, removedOp := range svcChange.RemovedOps {
				impl, _ := a.codeScanner.ScanService(svcChange.Name)
				var affectedFiles []string
				if impl != nil {
					if op, exists := impl.Operations[removedOp]; exists {
						affectedFiles = []string{op.File}
					}
				}

				impacts = append(impacts, models.ImplementationImpact{
					Service:              svcChange.Name,
					Operation:            removedOp,
					CurrentlyImplemented: len(affectedFiles) > 0,
					RequiresUpdate:       true,
					AffectedFiles:        affectedFiles,
					SuggestedAction:      "Remove or deprecate handler for " + removedOp,
				})
			}
		}
	}

	return impacts, nil
}

// estimateComplexity estimates the complexity of implementing a new operation
func estimateComplexity(op models.OperationInfo) string {
	paramCount := len(op.Parameters)
	if paramCount <= 3 {
		return "low"
	} else if paramCount <= 8 {
		return "medium"
	}
	return "high"
}

// generateSuggestedAction creates a suggested action for a modified operation
func generateSuggestedAction(diff models.OperationDiff) string {
	if diff.IsBreaking {
		if len(diff.NewParams) > 0 {
			return fmt.Sprintf("Add handling for new required parameters: %v", getParamNames(diff.NewParams))
		}
		return "Update handler to handle breaking changes"
	}
	if len(diff.NewParams) > 0 {
		return fmt.Sprintf("Consider adding support for new parameters: %v", getParamNames(diff.NewParams))
	}
	return "Review changes for any impact on implementation"
}

func getParamNames(params []models.ParameterChange) []string {
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}
	return names
}
