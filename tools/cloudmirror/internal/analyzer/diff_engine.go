package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// DiffEngine compares AWS models with implementations
type DiffEngine struct{}

// NewDiffEngine creates a new diff engine
func NewDiffEngine() *DiffEngine {
	return &DiffEngine{}
}

// Compare generates a coverage report comparing AWS model to implementation
func (d *DiffEngine) Compare(aws *models.AWSService, impl *models.ServiceImplementation) *models.CoverageReport {
	report := &models.CoverageReport{
		ServiceName:     aws.Name,
		ServiceFullName: aws.FullName,
		APIVersion:      aws.APIVersion,
		Protocol:        aws.Protocol,
		TotalOperations: len(aws.Operations),
		Supported:       []models.OperationStatus{},
		Missing:         []models.OperationStatus{},
		Deprecated:      []models.OperationStatus{},
		ParameterGaps:   []models.ParameterGap{},
		GeneratedAt:     time.Now(),
	}

	if impl == nil {
		// Service not implemented at all
		for opName, op := range aws.Operations {
			status := models.OperationStatus{
				Name:          opName,
				Documentation: op.Documentation,
				Deprecated:    op.Deprecated,
				DeprecatedMsg: op.DeprecatedMsg,
				Priority:      d.calculatePriority(opName),
				Implemented:   false,
			}

			if op.Deprecated {
				report.Deprecated = append(report.Deprecated, status)
			} else {
				report.Missing = append(report.Missing, status)
			}
		}
		report.CoveragePercent = 0
		d.sortReport(report)
		return report
	}

	for opName, awsOp := range aws.Operations {
		implOp, implemented := impl.Operations[opName]

		status := models.OperationStatus{
			Name:          opName,
			Documentation: awsOp.Documentation,
			Deprecated:    awsOp.Deprecated,
			DeprecatedMsg: awsOp.DeprecatedMsg,
			Priority:      d.calculatePriority(opName),
		}

		if awsOp.Deprecated {
			if implemented {
				status.Implemented = true
				status.Handler = implOp.Handler
				status.File = implOp.File
				status.Line = implOp.Line
			}
			report.Deprecated = append(report.Deprecated, status)
			continue
		}

		if implemented {
			status.Implemented = true
			status.Handler = implOp.Handler
			status.File = implOp.File
			status.Line = implOp.Line
			report.Supported = append(report.Supported, status)
		} else {
			report.Missing = append(report.Missing, status)
		}
	}

	// Calculate coverage (excluding deprecated operations)
	nonDeprecatedTotal := report.TotalOperations - len(report.Deprecated)
	if nonDeprecatedTotal > 0 {
		report.CoveragePercent = float64(len(report.Supported)) / float64(nonDeprecatedTotal) * 100
	}

	d.sortReport(report)
	return report
}

func (d *DiffEngine) calculatePriority(opName string) models.Priority {
	// High priority: Core CRUD operations
	highPrefixes := []string{
		"Create", "Describe", "Delete", "List", "Get", "Put",
	}
	for _, prefix := range highPrefixes {
		if strings.HasPrefix(opName, prefix) {
			return models.PriorityHigh
		}
	}

	// Medium priority: Modify/lifecycle operations
	mediumPrefixes := []string{
		"Modify", "Update", "Start", "Stop", "Reboot", "Enable", "Disable",
		"Add", "Remove", "Attach", "Detach", "Associate", "Disassociate",
	}
	for _, prefix := range mediumPrefixes {
		if strings.HasPrefix(opName, prefix) {
			return models.PriorityMedium
		}
	}

	// Low priority: Everything else (admin, batch, restore, etc.)
	return models.PriorityLow
}

func (d *DiffEngine) sortReport(report *models.CoverageReport) {
	// Sort by priority (high first), then alphabetically
	priorityOrder := map[models.Priority]int{
		models.PriorityHigh:   0,
		models.PriorityMedium: 1,
		models.PriorityLow:    2,
	}

	sortFunc := func(ops []models.OperationStatus) {
		sort.Slice(ops, func(i, j int) bool {
			if priorityOrder[ops[i].Priority] != priorityOrder[ops[j].Priority] {
				return priorityOrder[ops[i].Priority] < priorityOrder[ops[j].Priority]
			}
			return ops[i].Name < ops[j].Name
		})
	}

	sortFunc(report.Supported)
	sortFunc(report.Missing)
	sortFunc(report.Deprecated)
}

// CompareTwoVersions compares two AWS model versions to find changes
func (d *DiffEngine) CompareTwoVersions(oldModel, newModel *models.AWSService) *models.VersionDiff {
	diff := &models.VersionDiff{
		ServiceName:       newModel.Name,
		OldVersion:        oldModel.APIVersion,
		NewVersion:        newModel.APIVersion,
		NewOperations:     []string{},
		RemovedOperations: []string{},
		ChangedOperations: []models.OperationChange{},
	}

	// Find new operations
	for opName := range newModel.Operations {
		if _, exists := oldModel.Operations[opName]; !exists {
			diff.NewOperations = append(diff.NewOperations, opName)
		}
	}

	// Find removed operations
	for opName := range oldModel.Operations {
		if _, exists := newModel.Operations[opName]; !exists {
			diff.RemovedOperations = append(diff.RemovedOperations, opName)
		}
	}

	// Find changed operations (parameter changes, etc.)
	for opName, newOp := range newModel.Operations {
		if oldOp, exists := oldModel.Operations[opName]; exists {
			changes := d.findOperationChanges(oldOp, newOp)
			if len(changes) > 0 {
				diff.ChangedOperations = append(diff.ChangedOperations, models.OperationChange{
					Name:    opName,
					Changes: changes,
				})
			}
		}
	}

	// Sort for consistent output
	sort.Strings(diff.NewOperations)
	sort.Strings(diff.RemovedOperations)
	sort.Slice(diff.ChangedOperations, func(i, j int) bool {
		return diff.ChangedOperations[i].Name < diff.ChangedOperations[j].Name
	})

	return diff
}

func (d *DiffEngine) findOperationChanges(old, new *models.Operation) []string {
	var changes []string

	// Build maps for comparison
	oldParams := make(map[string]models.Parameter)
	for _, p := range old.Parameters {
		oldParams[p.Name] = p
	}

	newParams := make(map[string]models.Parameter)
	for _, p := range new.Parameters {
		newParams[p.Name] = p
	}

	// Check for new parameters
	for name, newParam := range newParams {
		if _, exists := oldParams[name]; !exists {
			if newParam.Required {
				changes = append(changes, fmt.Sprintf("New required parameter: %s", name))
			} else {
				changes = append(changes, fmt.Sprintf("New optional parameter: %s", name))
			}
		}
	}

	// Check for removed parameters
	for name := range oldParams {
		if _, exists := newParams[name]; !exists {
			changes = append(changes, fmt.Sprintf("Removed parameter: %s", name))
		}
	}

	// Check for changed parameters
	for name, newParam := range newParams {
		if oldParam, exists := oldParams[name]; exists {
			if !oldParam.Required && newParam.Required {
				changes = append(changes, fmt.Sprintf("Parameter now required: %s", name))
			}
			if oldParam.Required && !newParam.Required {
				changes = append(changes, fmt.Sprintf("Parameter now optional: %s", name))
			}
			if !oldParam.Deprecated && newParam.Deprecated {
				changes = append(changes, fmt.Sprintf("Parameter deprecated: %s", name))
			}
		}
	}

	return changes
}

// CompareReports compares two coverage reports to find changes
func (d *DiffEngine) CompareReports(baseline, current *models.CoverageReport) *models.VersionDiff {
	diff := &models.VersionDiff{
		ServiceName:       current.ServiceName,
		OldVersion:        baseline.APIVersion,
		NewVersion:        current.APIVersion,
		NewOperations:     []string{},
		RemovedOperations: []string{},
		ChangedOperations: []models.OperationChange{},
	}

	// Build maps of operations
	baselineOps := make(map[string]bool)
	for _, op := range baseline.Supported {
		baselineOps[op.Name] = true
	}
	for _, op := range baseline.Missing {
		baselineOps[op.Name] = true
	}

	currentOps := make(map[string]bool)
	for _, op := range current.Supported {
		currentOps[op.Name] = true
	}
	for _, op := range current.Missing {
		currentOps[op.Name] = true
	}

	// Find new operations in AWS API
	for name := range currentOps {
		if !baselineOps[name] {
			diff.NewOperations = append(diff.NewOperations, name)
		}
	}

	// Find removed operations from AWS API
	for name := range baselineOps {
		if !currentOps[name] {
			diff.RemovedOperations = append(diff.RemovedOperations, name)
		}
	}

	sort.Strings(diff.NewOperations)
	sort.Strings(diff.RemovedOperations)

	return diff
}

// CreateMultiServiceReport creates a summary report across multiple services
func (d *DiffEngine) CreateMultiServiceReport(reports []*models.CoverageReport) *models.MultiServiceReport {
	multi := &models.MultiServiceReport{
		Services:    reports,
		GeneratedAt: time.Now(),
	}

	for _, r := range reports {
		multi.TotalOperations += r.TotalOperations
		multi.TotalSupported += len(r.Supported)
		multi.TotalMissing += len(r.Missing)
	}

	if multi.TotalOperations > 0 {
		multi.OverallCoverage = float64(multi.TotalSupported) / float64(multi.TotalOperations) * 100
	}

	return multi
}

// CompareSDKVersions compares two SDK versions across all services and generates a comprehensive change report
func (d *DiffEngine) CompareSDKVersions(oldModels, newModels map[string]*models.AWSService, config *models.SDKCompareConfig) *models.SDKChangeReport {
	report := &models.SDKChangeReport{
		OldVersion:      config.OldSDKPath,
		NewVersion:      config.NewSDKPath,
		Services:        []models.ServiceChange{},
		BreakingChanges: []models.BreakingChange{},
		GeneratedAt:     time.Now(),
	}

	// Get union of all service names
	allServices := make(map[string]bool)
	for name := range oldModels {
		allServices[name] = true
	}
	for name := range newModels {
		allServices[name] = true
	}

	// Filter to specified services if configured
	if len(config.Services) > 0 {
		filterSet := make(map[string]bool)
		for _, s := range config.Services {
			filterSet[strings.ToLower(s)] = true
		}
		for name := range allServices {
			if !filterSet[strings.ToLower(name)] {
				delete(allServices, name)
			}
		}
	}

	// Filter to allowlist if enabled
	if config.UseAllowlist {
		for name := range allServices {
			if !models.IsServiceAllowed(name) {
				delete(allServices, name)
			}
		}
	}

	// Process each service
	for serviceName := range allServices {
		oldModel := oldModels[serviceName]
		newModel := newModels[serviceName]

		serviceChange := d.compareServiceModels(serviceName, oldModel, newModel)
		if serviceChange.HasChanges {
			report.Services = append(report.Services, serviceChange)

			// Extract breaking changes
			breakingChanges := d.extractBreakingChanges(serviceName, oldModel, newModel)
			report.BreakingChanges = append(report.BreakingChanges, breakingChanges...)
		}
	}

	// Sort services alphabetically
	sort.Slice(report.Services, func(i, j int) bool {
		return report.Services[i].Name < report.Services[j].Name
	})

	// Calculate summary
	report.Summary = d.calculateChangeSummary(report)

	// Filter to only breaking changes if configured
	if config.OnlyBreaking {
		var filteredServices []models.ServiceChange
		for _, svc := range report.Services {
			hasBreaking := false
			for _, bc := range report.BreakingChanges {
				if bc.Service == svc.Name {
					hasBreaking = true
					break
				}
			}
			if hasBreaking || len(svc.RemovedOps) > 0 {
				filteredServices = append(filteredServices, svc)
			}
		}
		report.Services = filteredServices
	}

	return report
}

// compareServiceModels compares old and new versions of a single service
func (d *DiffEngine) compareServiceModels(serviceName string, oldModel, newModel *models.AWSService) models.ServiceChange {
	change := models.ServiceChange{
		Name:       serviceName,
		HasChanges: false,
	}

	// Handle case where service is new
	if oldModel == nil && newModel != nil {
		change.FullName = newModel.FullName
		change.HasChanges = true
		for opName, op := range newModel.Operations {
			change.NewOperations = append(change.NewOperations, models.OperationInfo{
				Name:          opName,
				HTTPMethod:    op.HTTPMethod,
				HTTPPath:      op.HTTPPath,
				Documentation: truncateDoc(op.Documentation),
				Priority:      d.calculatePriority(opName),
				InputShape:    op.InputShape,
				OutputShape:   op.OutputShape,
			})
		}
		sort.Slice(change.NewOperations, func(i, j int) bool {
			return change.NewOperations[i].Name < change.NewOperations[j].Name
		})
		return change
	}

	// Handle case where service was removed
	if oldModel != nil && newModel == nil {
		change.FullName = oldModel.FullName
		change.HasChanges = true
		for opName := range oldModel.Operations {
			change.RemovedOps = append(change.RemovedOps, opName)
		}
		sort.Strings(change.RemovedOps)
		return change
	}

	// Compare operations between versions
	change.FullName = newModel.FullName

	// Find new operations
	for opName, newOp := range newModel.Operations {
		if _, exists := oldModel.Operations[opName]; !exists {
			change.NewOperations = append(change.NewOperations, models.OperationInfo{
				Name:          opName,
				HTTPMethod:    newOp.HTTPMethod,
				HTTPPath:      newOp.HTTPPath,
				Documentation: truncateDoc(newOp.Documentation),
				Priority:      d.calculatePriority(opName),
				InputShape:    newOp.InputShape,
				OutputShape:   newOp.OutputShape,
			})
			change.HasChanges = true
		}
	}

	// Find removed operations
	for opName := range oldModel.Operations {
		if _, exists := newModel.Operations[opName]; !exists {
			change.RemovedOps = append(change.RemovedOps, opName)
			change.HasChanges = true
		}
	}

	// Find modified operations
	for opName, newOp := range newModel.Operations {
		oldOp, exists := oldModel.Operations[opName]
		if !exists {
			continue
		}

		opDiff := d.compareOperations(opName, oldOp, newOp)
		if opDiff != nil {
			change.ModifiedOps = append(change.ModifiedOps, *opDiff)
			change.HasChanges = true
		}

		// Track newly deprecated operations
		if !oldOp.Deprecated && newOp.Deprecated {
			change.DeprecatedOps = append(change.DeprecatedOps, opName)
			change.HasChanges = true
		}
	}

	// Sort results
	sort.Slice(change.NewOperations, func(i, j int) bool {
		return change.NewOperations[i].Name < change.NewOperations[j].Name
	})
	sort.Strings(change.RemovedOps)
	sort.Strings(change.DeprecatedOps)
	sort.Slice(change.ModifiedOps, func(i, j int) bool {
		return change.ModifiedOps[i].Name < change.ModifiedOps[j].Name
	})

	return change
}

// compareOperations compares two versions of an operation
func (d *DiffEngine) compareOperations(opName string, oldOp, newOp *models.Operation) *models.OperationDiff {
	diff := &models.OperationDiff{
		Name: opName,
	}

	hasChanges := false

	// Build parameter maps
	oldParams := make(map[string]models.Parameter)
	for _, p := range oldOp.Parameters {
		oldParams[p.Name] = p
	}

	newParams := make(map[string]models.Parameter)
	for _, p := range newOp.Parameters {
		newParams[p.Name] = p
	}

	// Find new parameters
	for name, newParam := range newParams {
		if _, exists := oldParams[name]; !exists {
			diff.NewParams = append(diff.NewParams, models.ParameterChange{
				Name:     name,
				Type:     newParam.Type,
				Required: newParam.Required,
			})
			hasChanges = true

			// New required params are breaking
			if newParam.Required {
				diff.IsBreaking = true
				diff.Changes = append(diff.Changes, fmt.Sprintf("New required parameter: %s (%s)", name, newParam.Type))
			} else {
				diff.Changes = append(diff.Changes, fmt.Sprintf("New optional parameter: %s (%s)", name, newParam.Type))
			}
		}
	}

	// Find removed parameters
	for name := range oldParams {
		if _, exists := newParams[name]; !exists {
			diff.RemovedParams = append(diff.RemovedParams, name)
			diff.Changes = append(diff.Changes, fmt.Sprintf("Removed parameter: %s", name))
			hasChanges = true
			// Removing parameters can break implementations that rely on them in responses
		}
	}

	// Find modified parameters
	for name, newParam := range newParams {
		oldParam, exists := oldParams[name]
		if !exists {
			continue
		}

		// Check if parameter is now required
		if !oldParam.Required && newParam.Required {
			diff.ModifiedParams = append(diff.ModifiedParams, models.ParameterChange{
				Name:        name,
				Type:        newParam.Type,
				Required:    true,
				WasRequired: false,
			})
			diff.Changes = append(diff.Changes, fmt.Sprintf("Parameter now required: %s", name))
			diff.IsBreaking = true
			hasChanges = true
		}

		// Check if parameter is now optional
		if oldParam.Required && !newParam.Required {
			diff.ModifiedParams = append(diff.ModifiedParams, models.ParameterChange{
				Name:        name,
				Type:        newParam.Type,
				Required:    false,
				WasRequired: true,
			})
			diff.Changes = append(diff.Changes, fmt.Sprintf("Parameter now optional: %s", name))
			hasChanges = true
		}

		// Check if type changed
		if oldParam.Type != newParam.Type && oldParam.Type != "" && newParam.Type != "" {
			diff.ModifiedParams = append(diff.ModifiedParams, models.ParameterChange{
				Name:        name,
				Type:        newParam.Type,
				Required:    newParam.Required,
				Description: fmt.Sprintf("Type changed from %s to %s", oldParam.Type, newParam.Type),
			})
			diff.Changes = append(diff.Changes, fmt.Sprintf("Parameter type changed: %s (%s -> %s)", name, oldParam.Type, newParam.Type))
			diff.IsBreaking = true
			hasChanges = true
		}

		// Check if deprecated
		if !oldParam.Deprecated && newParam.Deprecated {
			diff.Changes = append(diff.Changes, fmt.Sprintf("Parameter deprecated: %s", name))
			hasChanges = true
		}
	}

	// Check if operation itself is now deprecated
	if !oldOp.Deprecated && newOp.Deprecated {
		diff.DeprecatedNow = true
		diff.Changes = append(diff.Changes, "Operation is now deprecated")
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	// Sort results
	sort.Slice(diff.NewParams, func(i, j int) bool {
		return diff.NewParams[i].Name < diff.NewParams[j].Name
	})
	sort.Strings(diff.RemovedParams)
	sort.Slice(diff.ModifiedParams, func(i, j int) bool {
		return diff.ModifiedParams[i].Name < diff.ModifiedParams[j].Name
	})
	sort.Strings(diff.Changes)

	return diff
}

// extractBreakingChanges identifies all breaking changes between versions
func (d *DiffEngine) extractBreakingChanges(serviceName string, oldModel, newModel *models.AWSService) []models.BreakingChange {
	var breaking []models.BreakingChange

	// Service removed entirely
	if oldModel != nil && newModel == nil {
		for opName := range oldModel.Operations {
			breaking = append(breaking, models.BreakingChange{
				Service:     serviceName,
				Operation:   opName,
				Reason:      "Service removed from AWS SDK",
				Severity:    models.SeverityCritical,
				Type:        models.BreakingTypeOperationRemoved,
				Remediation: "Remove implementation or maintain as deprecated",
			})
		}
		return breaking
	}

	if oldModel == nil || newModel == nil {
		return breaking
	}

	// Check for removed operations
	for opName := range oldModel.Operations {
		if _, exists := newModel.Operations[opName]; !exists {
			breaking = append(breaking, models.BreakingChange{
				Service:     serviceName,
				Operation:   opName,
				Reason:      "Operation removed from AWS API",
				Severity:    models.SeverityCritical,
				Type:        models.BreakingTypeOperationRemoved,
				Remediation: "Remove handler or maintain as deprecated stub",
			})
		}
	}

	// Check for breaking parameter changes
	for opName, newOp := range newModel.Operations {
		oldOp, exists := oldModel.Operations[opName]
		if !exists {
			continue
		}

		// Build old params map
		oldParams := make(map[string]models.Parameter)
		for _, p := range oldOp.Parameters {
			oldParams[p.Name] = p
		}

		// Check new params
		for _, newParam := range newOp.Parameters {
			oldParam, existed := oldParams[newParam.Name]

			// New required parameter
			if !existed && newParam.Required {
				breaking = append(breaking, models.BreakingChange{
					Service:     serviceName,
					Operation:   opName,
					Reason:      fmt.Sprintf("New required parameter: %s", newParam.Name),
					Severity:    models.SeverityCritical,
					Type:        models.BreakingTypeNewRequiredParam,
					Details:     fmt.Sprintf("Type: %s", newParam.Type),
					Remediation: fmt.Sprintf("Add %s parameter handling to the operation handler", newParam.Name),
				})
			}

			// Parameter became required
			if existed && !oldParam.Required && newParam.Required {
				breaking = append(breaking, models.BreakingChange{
					Service:     serviceName,
					Operation:   opName,
					Reason:      fmt.Sprintf("Parameter now required: %s", newParam.Name),
					Severity:    models.SeverityHigh,
					Type:        models.BreakingTypeParamNowRequired,
					Remediation: fmt.Sprintf("Ensure %s parameter is validated as required", newParam.Name),
				})
			}

			// Type changed
			if existed && oldParam.Type != newParam.Type && oldParam.Type != "" && newParam.Type != "" {
				breaking = append(breaking, models.BreakingChange{
					Service:     serviceName,
					Operation:   opName,
					Reason:      fmt.Sprintf("Parameter type changed: %s (%s -> %s)", newParam.Name, oldParam.Type, newParam.Type),
					Severity:    models.SeverityHigh,
					Type:        models.BreakingTypeParamTypeChanged,
					Remediation: fmt.Sprintf("Update parameter parsing for %s to handle new type", newParam.Name),
				})
			}
		}

		// Check for removed parameters (may affect response structure)
		for _, oldParam := range oldOp.Parameters {
			found := false
			for _, newParam := range newOp.Parameters {
				if newParam.Name == oldParam.Name {
					found = true
					break
				}
			}
			if !found {
				breaking = append(breaking, models.BreakingChange{
					Service:     serviceName,
					Operation:   opName,
					Reason:      fmt.Sprintf("Parameter removed: %s", oldParam.Name),
					Severity:    models.SeverityMedium,
					Type:        models.BreakingTypeParamRemoved,
					Details:     "Callers may still send this parameter",
					Remediation: "Ignore the parameter if received, or return deprecation warning",
				})
			}
		}
	}

	// Sort by severity and then by service/operation
	sort.Slice(breaking, func(i, j int) bool {
		severityOrder := map[models.Severity]int{
			models.SeverityCritical: 0,
			models.SeverityHigh:     1,
			models.SeverityMedium:   2,
			models.SeverityLow:      3,
		}
		if severityOrder[breaking[i].Severity] != severityOrder[breaking[j].Severity] {
			return severityOrder[breaking[i].Severity] < severityOrder[breaking[j].Severity]
		}
		if breaking[i].Service != breaking[j].Service {
			return breaking[i].Service < breaking[j].Service
		}
		return breaking[i].Operation < breaking[j].Operation
	})

	return breaking
}

// calculateChangeSummary calculates aggregate statistics for the report
func (d *DiffEngine) calculateChangeSummary(report *models.SDKChangeReport) models.ChangeSummary {
	summary := models.ChangeSummary{}

	for _, svc := range report.Services {
		if svc.HasChanges {
			summary.TotalServicesChanged++
		}
		summary.TotalNewOperations += len(svc.NewOperations)
		summary.TotalModifiedOps += len(svc.ModifiedOps)
		summary.TotalDeprecatedOps += len(svc.DeprecatedOps)
		summary.TotalRemovedOps += len(svc.RemovedOps)
	}

	summary.TotalBreakingChanges = len(report.BreakingChanges)
	for _, bc := range report.BreakingChanges {
		switch bc.Severity {
		case models.SeverityCritical:
			summary.CriticalBreakingCount++
		case models.SeverityHigh:
			summary.HighBreakingCount++
		case models.SeverityMedium:
			summary.MediumBreakingCount++
		case models.SeverityLow:
			summary.LowBreakingCount++
		}
	}

	return summary
}

// truncateDoc truncates documentation to a reasonable length
func truncateDoc(doc string) string {
	// Remove HTML tags
	doc = strings.ReplaceAll(doc, "<p>", "")
	doc = strings.ReplaceAll(doc, "</p>", " ")
	doc = strings.ReplaceAll(doc, "<code>", "`")
	doc = strings.ReplaceAll(doc, "</code>", "`")
	doc = strings.TrimSpace(doc)

	// Truncate if too long
	if len(doc) > 200 {
		return doc[:197] + "..."
	}
	return doc
}
