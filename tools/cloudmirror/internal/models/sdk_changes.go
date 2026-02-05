package models

import "time"

// SDKChangeReport represents a comprehensive report of changes between two SDK versions
type SDKChangeReport struct {
	OldVersion      string           `json:"old_version"`
	NewVersion      string           `json:"new_version"`
	Services        []ServiceChange  `json:"services"`
	BreakingChanges []BreakingChange `json:"breaking_changes"`
	Summary         ChangeSummary    `json:"summary"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

// ServiceChange represents changes to a single AWS service between versions
type ServiceChange struct {
	Name          string          `json:"name"`
	FullName      string          `json:"full_name,omitempty"`
	NewOperations []OperationInfo `json:"new_operations,omitempty"`
	ModifiedOps   []OperationDiff `json:"modified_operations,omitempty"`
	DeprecatedOps []string        `json:"deprecated_operations,omitempty"`
	RemovedOps    []string        `json:"removed_operations,omitempty"`
	HasChanges    bool            `json:"has_changes"`
}

// OperationInfo contains information about a new operation
type OperationInfo struct {
	Name          string      `json:"name"`
	HTTPMethod    string      `json:"http_method,omitempty"`
	HTTPPath      string      `json:"http_path,omitempty"`
	Documentation string      `json:"documentation,omitempty"`
	Priority      Priority    `json:"priority"`
	Parameters    []Parameter `json:"parameters,omitempty"`
	InputShape    string      `json:"input_shape,omitempty"`
	OutputShape   string      `json:"output_shape,omitempty"`
}

// OperationDiff describes modifications to an existing operation
type OperationDiff struct {
	Name           string            `json:"name"`
	NewParams      []ParameterChange `json:"new_parameters,omitempty"`
	RemovedParams  []string          `json:"removed_parameters,omitempty"`
	ModifiedParams []ParameterChange `json:"modified_parameters,omitempty"`
	DeprecatedNow  bool              `json:"deprecated_now,omitempty"`
	Changes        []string          `json:"changes,omitempty"` // Human-readable change descriptions
	IsBreaking     bool              `json:"is_breaking"`
}

// ParameterChange describes a change to a parameter
type ParameterChange struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required"`
	WasRequired bool   `json:"was_required,omitempty"` // For modified params
	Description string `json:"description,omitempty"`
}

// BreakingChange represents a change that could break existing implementations
type BreakingChange struct {
	Service     string       `json:"service"`
	Operation   string       `json:"operation"`
	Reason      string       `json:"reason"`
	Severity    Severity     `json:"severity"`
	Details     string       `json:"details,omitempty"`
	Remediation string       `json:"remediation,omitempty"`
	Type        BreakingType `json:"type"`
}

// Severity represents the severity of a breaking change
type Severity string

const (
	SeverityCritical Severity = "critical" // Will definitely break: removed operation, new required param
	SeverityHigh     Severity = "high"     // Likely to break: type changes, validation changes
	SeverityMedium   Severity = "medium"   // May break: behavioral changes
	SeverityLow      Severity = "low"      // Unlikely to break: deprecation warnings
)

// BreakingType categorizes the type of breaking change
type BreakingType string

const (
	BreakingTypeOperationRemoved    BreakingType = "operation_removed"
	BreakingTypeNewRequiredParam    BreakingType = "new_required_parameter"
	BreakingTypeParamTypeChanged    BreakingType = "parameter_type_changed"
	BreakingTypeParamNowRequired    BreakingType = "parameter_now_required"
	BreakingTypeParamRemoved        BreakingType = "parameter_removed"
	BreakingTypeResponseTypeChanged BreakingType = "response_type_changed"
	BreakingTypeEnumValueRemoved    BreakingType = "enum_value_removed"
	BreakingTypeValidationTightened BreakingType = "validation_tightened"
	BreakingTypeEndpointChanged     BreakingType = "endpoint_changed"
	BreakingTypeBehaviorChanged     BreakingType = "behavior_changed"
)

// ChangeSummary provides aggregate statistics about the changes
type ChangeSummary struct {
	TotalServicesChanged  int `json:"total_services_changed"`
	TotalNewOperations    int `json:"total_new_operations"`
	TotalModifiedOps      int `json:"total_modified_operations"`
	TotalDeprecatedOps    int `json:"total_deprecated_operations"`
	TotalRemovedOps       int `json:"total_removed_operations"`
	TotalBreakingChanges  int `json:"total_breaking_changes"`
	CriticalBreakingCount int `json:"critical_breaking_count"`
	HighBreakingCount     int `json:"high_breaking_count"`
	MediumBreakingCount   int `json:"medium_breaking_count"`
	LowBreakingCount      int `json:"low_breaking_count"`
}

// ImplementationImpact describes how changes affect our emulator implementation
type ImplementationImpact struct {
	Service              string   `json:"service"`
	Operation            string   `json:"operation"`
	CurrentlyImplemented bool     `json:"currently_implemented"`
	RequiresUpdate       bool     `json:"requires_update"`
	RequiresNewHandler   bool     `json:"requires_new_handler"`
	AffectedFiles        []string `json:"affected_files,omitempty"`
	EstimatedComplexity  string   `json:"estimated_complexity,omitempty"` // "low", "medium", "high"
	SuggestedAction      string   `json:"suggested_action,omitempty"`
}

// SDKCompareConfig holds configuration for SDK comparison
type SDKCompareConfig struct {
	OldSDKPath     string   `json:"old_sdk_path"`
	NewSDKPath     string   `json:"new_sdk_path"`
	Services       []string `json:"services,omitempty"` // Empty means all services
	IncludeMinor   bool     `json:"include_minor"`      // Include minor/patch changes
	OnlyBreaking   bool     `json:"only_breaking"`      // Only report breaking changes
	CheckImplement bool     `json:"check_implement"`    // Cross-reference with implementation
	UseAllowlist   bool     `json:"use_allowlist"`      // Only include services from the allowlist
}
