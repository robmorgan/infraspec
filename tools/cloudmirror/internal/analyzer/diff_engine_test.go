package analyzer

import (
	"testing"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

func TestDiffEngine_Compare_FullImplementation(t *testing.T) {
	engine := NewDiffEngine()

	aws := &models.AWSService{
		Name:     "test-service",
		FullName: "Test Service",
		Operations: map[string]*models.Operation{
			"CreateResource": {Name: "CreateResource"},
			"DeleteResource": {Name: "DeleteResource"},
			"ListResources":  {Name: "ListResources"},
		},
	}

	impl := &models.ServiceImplementation{
		Name: "test-service",
		Operations: map[string]*models.ImplementedOperation{
			"CreateResource": {Name: "CreateResource", Handler: "handleCreate"},
			"DeleteResource": {Name: "DeleteResource", Handler: "handleDelete"},
			"ListResources":  {Name: "ListResources", Handler: "handleList"},
		},
	}

	report := engine.Compare(aws, impl)

	if report.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", report.ServiceName)
	}

	if report.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", report.TotalOperations)
	}

	if len(report.Supported) != 3 {
		t.Errorf("Expected 3 supported operations, got %d", len(report.Supported))
	}

	if len(report.Missing) != 0 {
		t.Errorf("Expected 0 missing operations, got %d", len(report.Missing))
	}

	if report.CoveragePercent != 100.0 {
		t.Errorf("Expected 100%% coverage, got %.2f%%", report.CoveragePercent)
	}
}

func TestDiffEngine_Compare_PartialImplementation(t *testing.T) {
	engine := NewDiffEngine()

	aws := &models.AWSService{
		Name: "test-service",
		Operations: map[string]*models.Operation{
			"CreateResource":   {Name: "CreateResource"},
			"DeleteResource":   {Name: "DeleteResource"},
			"ListResources":    {Name: "ListResources"},
			"DescribeResource": {Name: "DescribeResource"},
		},
	}

	impl := &models.ServiceImplementation{
		Name: "test-service",
		Operations: map[string]*models.ImplementedOperation{
			"CreateResource": {Name: "CreateResource", Handler: "handleCreate"},
			"ListResources":  {Name: "ListResources", Handler: "handleList"},
		},
	}

	report := engine.Compare(aws, impl)

	if len(report.Supported) != 2 {
		t.Errorf("Expected 2 supported operations, got %d", len(report.Supported))
	}

	if len(report.Missing) != 2 {
		t.Errorf("Expected 2 missing operations, got %d", len(report.Missing))
	}

	// Coverage should be 50%
	if report.CoveragePercent != 50.0 {
		t.Errorf("Expected 50%% coverage, got %.2f%%", report.CoveragePercent)
	}
}

func TestDiffEngine_Compare_NoImplementation(t *testing.T) {
	engine := NewDiffEngine()

	aws := &models.AWSService{
		Name: "test-service",
		Operations: map[string]*models.Operation{
			"CreateResource": {Name: "CreateResource"},
			"DeleteResource": {Name: "DeleteResource"},
		},
	}

	report := engine.Compare(aws, nil)

	if len(report.Supported) != 0 {
		t.Errorf("Expected 0 supported operations, got %d", len(report.Supported))
	}

	if len(report.Missing) != 2 {
		t.Errorf("Expected 2 missing operations, got %d", len(report.Missing))
	}

	if report.CoveragePercent != 0 {
		t.Errorf("Expected 0%% coverage, got %.2f%%", report.CoveragePercent)
	}
}

func TestDiffEngine_Compare_DeprecatedOperations(t *testing.T) {
	engine := NewDiffEngine()

	aws := &models.AWSService{
		Name: "test-service",
		Operations: map[string]*models.Operation{
			"CreateResource":    {Name: "CreateResource"},
			"OldDeprecatedCall": {Name: "OldDeprecatedCall", Deprecated: true},
		},
	}

	impl := &models.ServiceImplementation{
		Name: "test-service",
		Operations: map[string]*models.ImplementedOperation{
			"CreateResource": {Name: "CreateResource", Handler: "handleCreate"},
		},
	}

	report := engine.Compare(aws, impl)

	if len(report.Deprecated) != 1 {
		t.Errorf("Expected 1 deprecated operation, got %d", len(report.Deprecated))
	}

	// Coverage should exclude deprecated operations
	// Total: 2, Deprecated: 1, Non-deprecated: 1, Supported: 1 = 100%
	if report.CoveragePercent != 100.0 {
		t.Errorf("Expected 100%% coverage (excluding deprecated), got %.2f%%", report.CoveragePercent)
	}
}

func TestDiffEngine_CalculatePriority(t *testing.T) {
	engine := NewDiffEngine()

	tests := []struct {
		opName   string
		expected models.Priority
	}{
		// High priority operations
		{"CreateDBInstance", models.PriorityHigh},
		{"DescribeTables", models.PriorityHigh},
		{"DeleteBucket", models.PriorityHigh},
		{"ListObjects", models.PriorityHigh},
		{"GetObject", models.PriorityHigh},
		{"PutItem", models.PriorityHigh},

		// Medium priority operations
		{"ModifyDBInstance", models.PriorityMedium},
		{"UpdateTable", models.PriorityMedium},
		{"StartInstances", models.PriorityMedium},
		{"StopInstances", models.PriorityMedium},
		{"RebootDBInstance", models.PriorityMedium},
		{"EnableAlarmActions", models.PriorityMedium},
		{"DisableAlarmActions", models.PriorityMedium},
		{"AddTagsToResource", models.PriorityMedium},
		{"RemoveTagsFromResource", models.PriorityMedium},
		{"AttachVolume", models.PriorityMedium},
		{"DetachVolume", models.PriorityMedium},
		{"AssociateAddress", models.PriorityMedium},
		{"DisassociateAddress", models.PriorityMedium},

		// Low priority operations
		{"RestoreDBInstanceFromSnapshot", models.PriorityLow},
		{"CopySnapshot", models.PriorityLow},
		{"BatchWriteItem", models.PriorityLow},
		{"AdminCreateUser", models.PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.opName, func(t *testing.T) {
			got := engine.calculatePriority(tt.opName)
			if got != tt.expected {
				t.Errorf("calculatePriority(%q) = %q, want %q", tt.opName, got, tt.expected)
			}
		})
	}
}

func TestDiffEngine_CompareTwoVersions(t *testing.T) {
	engine := NewDiffEngine()

	oldModel := &models.AWSService{
		Name:       "test-service",
		APIVersion: "v1",
		Operations: map[string]*models.Operation{
			"CreateResource": {Name: "CreateResource"},
			"DeleteResource": {Name: "DeleteResource"},
			"OldOperation":   {Name: "OldOperation"},
		},
	}

	newModel := &models.AWSService{
		Name:       "test-service",
		APIVersion: "v2",
		Operations: map[string]*models.Operation{
			"CreateResource": {Name: "CreateResource"},
			"DeleteResource": {Name: "DeleteResource"},
			"NewOperation":   {Name: "NewOperation"},
		},
	}

	diff := engine.CompareTwoVersions(oldModel, newModel)

	if len(diff.NewOperations) != 1 || diff.NewOperations[0] != "NewOperation" {
		t.Errorf("Expected 1 new operation 'NewOperation', got %v", diff.NewOperations)
	}

	if len(diff.RemovedOperations) != 1 || diff.RemovedOperations[0] != "OldOperation" {
		t.Errorf("Expected 1 removed operation 'OldOperation', got %v", diff.RemovedOperations)
	}
}

func TestDiffEngine_CompareReports(t *testing.T) {
	engine := NewDiffEngine()

	baseline := &models.CoverageReport{
		ServiceName: "test-service",
		APIVersion:  "v1",
		Supported: []models.OperationStatus{
			{Name: "CreateResource"},
		},
		Missing: []models.OperationStatus{
			{Name: "DeleteResource"},
		},
	}

	current := &models.CoverageReport{
		ServiceName: "test-service",
		APIVersion:  "v2",
		Supported: []models.OperationStatus{
			{Name: "CreateResource"},
		},
		Missing: []models.OperationStatus{
			{Name: "DeleteResource"},
			{Name: "NewOperation"},
		},
	}

	diff := engine.CompareReports(baseline, current)

	if len(diff.NewOperations) != 1 || diff.NewOperations[0] != "NewOperation" {
		t.Errorf("Expected 1 new operation 'NewOperation', got %v", diff.NewOperations)
	}
}

func TestDiffEngine_CreateMultiServiceReport(t *testing.T) {
	engine := NewDiffEngine()

	reports := []*models.CoverageReport{
		{
			ServiceName:     "service1",
			TotalOperations: 10,
			Supported:       make([]models.OperationStatus, 5),
			Missing:         make([]models.OperationStatus, 5),
		},
		{
			ServiceName:     "service2",
			TotalOperations: 20,
			Supported:       make([]models.OperationStatus, 10),
			Missing:         make([]models.OperationStatus, 10),
		},
	}

	multi := engine.CreateMultiServiceReport(reports)

	if multi.TotalOperations != 30 {
		t.Errorf("Expected 30 total operations, got %d", multi.TotalOperations)
	}

	if multi.TotalSupported != 15 {
		t.Errorf("Expected 15 supported operations, got %d", multi.TotalSupported)
	}

	if multi.TotalMissing != 15 {
		t.Errorf("Expected 15 missing operations, got %d", multi.TotalMissing)
	}

	if multi.OverallCoverage != 50.0 {
		t.Errorf("Expected 50%% overall coverage, got %.2f%%", multi.OverallCoverage)
	}
}
