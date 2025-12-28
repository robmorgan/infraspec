package models

import "time"

// Priority represents the implementation priority of an operation
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// CoverageReport represents the coverage analysis for a service
type CoverageReport struct {
	ServiceName     string            `json:"service_name"`
	ServiceFullName string            `json:"service_full_name"`
	APIVersion      string            `json:"api_version"`
	Protocol        string            `json:"protocol"`
	TotalOperations int               `json:"total_operations"`
	CoveragePercent float64           `json:"coverage_percent"`
	Supported       []OperationStatus `json:"supported"`
	Missing         []OperationStatus `json:"missing"`
	Deprecated      []OperationStatus `json:"deprecated"`
	ParameterGaps   []ParameterGap    `json:"parameter_gaps,omitempty"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// OperationStatus represents the implementation status of an operation
type OperationStatus struct {
	Name          string   `json:"name"`
	Documentation string   `json:"documentation,omitempty"`
	Deprecated    bool     `json:"deprecated,omitempty"`
	DeprecatedMsg string   `json:"deprecated_message,omitempty"`
	Implemented   bool     `json:"implemented"`
	Handler       string   `json:"handler,omitempty"`
	File          string   `json:"file,omitempty"`
	Line          int      `json:"line,omitempty"`
	Priority      Priority `json:"priority"`
}

// ParameterGap represents missing or incomplete parameter support
type ParameterGap struct {
	Operation  string   `json:"operation"`
	Parameters []string `json:"parameters"`
}

// VersionDiff represents changes between two AWS API versions
type VersionDiff struct {
	ServiceName       string            `json:"service_name"`
	OldVersion        string            `json:"old_version"`
	NewVersion        string            `json:"new_version"`
	NewOperations     []string          `json:"new_operations"`
	RemovedOperations []string          `json:"removed_operations"`
	ChangedOperations []OperationChange `json:"changed_operations"`
}

// OperationChange represents changes to an operation between versions
type OperationChange struct {
	Name    string   `json:"name"`
	Changes []string `json:"changes"`
}

// MultiServiceReport represents coverage across multiple services
type MultiServiceReport struct {
	Services        []*CoverageReport `json:"services"`
	TotalOperations int               `json:"total_operations"`
	TotalSupported  int               `json:"total_supported"`
	TotalMissing    int               `json:"total_missing"`
	OverallCoverage float64           `json:"overall_coverage"`
	GeneratedAt     time.Time         `json:"generated_at"`
}
