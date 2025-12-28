// Package generator provides AI-assisted code generation for AWS service implementations.
package aigen

import "time"

// SDKChangeReport represents changes between two AWS SDK versions.
type SDKChangeReport struct {
	OldVersion      string           `json:"old_version"`
	NewVersion      string           `json:"new_version"`
	Services        []ServiceChange  `json:"services"`
	BreakingChanges []BreakingChange `json:"breaking_changes"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

// ServiceChange represents changes to a specific AWS service.
type ServiceChange struct {
	Name           string          `json:"name"`
	Protocol       string          `json:"protocol"`
	NewOperations  []OperationInfo `json:"new_operations"`
	ModifiedOps    []OperationDiff `json:"modified_operations"`
	DeprecatedOps  []string        `json:"deprecated_operations"`
	RemovedOps     []string        `json:"removed_operations"`
}

// OperationInfo contains information about a new operation.
type OperationInfo struct {
	Name          string      `json:"name"`
	Documentation string      `json:"documentation"`
	Priority      string      `json:"priority"` // high, medium, low
	InputType     string      `json:"input_type"`
	OutputType    string      `json:"output_type"`
	Parameters    []Parameter `json:"parameters"`
}

// Parameter represents an operation parameter.
type Parameter struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	Deprecated bool   `json:"deprecated"`
	Location   string `json:"location"` // header, querystring, uri, body
}

// OperationDiff represents changes to an existing operation.
type OperationDiff struct {
	Name    string   `json:"name"`
	Changes []string `json:"changes"`
}

// BreakingChange represents a breaking API change.
type BreakingChange struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Reason    string `json:"reason"`
	Details   string `json:"details"`
}

// GenerationResult represents the result of code generation.
type GenerationResult struct {
	ServicesProcessed int               `json:"services_processed"`
	OperationsCreated int               `json:"operations_created"`
	TestsCreated      int               `json:"tests_created"`
	LimitReached      bool              `json:"limit_reached,omitempty"`
	Errors            []GenerationError `json:"errors,omitempty"`
	Files             []GeneratedFile   `json:"files"`
	Summary           GenerationSummary `json:"summary"`
}

// GeneratedFile represents a file that was generated.
type GeneratedFile struct {
	Path       string `json:"path"`
	Type       string `json:"type"` // handler, types, test
	Service    string `json:"service"`
	Operation  string `json:"operation,omitempty"`
	LinesOfCode int   `json:"lines_of_code"`
}

// GenerationError represents an error during generation.
type GenerationError struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Phase     string `json:"phase"` // prompt, generation, validation
	Message   string `json:"message"`
}

// GenerationSummary provides a summary of the generation.
type GenerationSummary struct {
	TotalOperations   int            `json:"total_operations"`
	SuccessfullyGen   int            `json:"successfully_generated"`
	FailedGeneration  int            `json:"failed_generation"`
	SkippedOperations int            `json:"skipped_operations"`
	ByService         map[string]int `json:"by_service"`
	ByPriority        map[string]int `json:"by_priority"`
}

// Config holds configuration for the implementation generator.
type Config struct {
	SDKPath       string
	ServicesPath  string
	OutputDir     string
	ClaudeAPIKey  string
	ClaudeModel   string
	DryRun        bool
	Verbose       bool
	MaxOperations int // Maximum number of operations to generate (0 = unlimited)
}

// GeneratedCode represents generated handler code.
type GeneratedCode struct {
	Service     string
	Operation   string
	HandlerCode string
	TypesCode   string
	TestCode    string
	Validated   bool
	Errors      []string
}

// ValidationResult represents the result of code validation.
type ValidationResult struct {
	FilePath      string   `json:"file_path"`
	Valid         bool     `json:"valid"`
	CompileErrors []string `json:"compile_errors,omitempty"`
	VetErrors     []string `json:"vet_errors,omitempty"`
	PatternErrors []string `json:"pattern_errors,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

// ValidatorConfig holds configuration for the code validator.
type ValidatorConfig struct {
	Verbose bool
}
