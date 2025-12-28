package models

// ServiceImplementation represents a scanned InfraSpec service implementation
type ServiceImplementation struct {
	Name       string                           `json:"name"`
	Path       string                           `json:"path"`
	Operations map[string]*ImplementedOperation `json:"operations"`
}

// ImplementedOperation represents an operation found in the source code
type ImplementedOperation struct {
	Name        string `json:"name"`
	Handler     string `json:"handler"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Implemented bool   `json:"implemented"`
}
