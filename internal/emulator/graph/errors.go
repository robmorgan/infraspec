package graph

import (
	"fmt"
	"strings"
)

// DependencyError indicates that an operation failed due to resource dependencies.
type DependencyError struct {
	Resource   ResourceID
	Dependents []ResourceID
	Message    string
}

// Error implements the error interface.
func (e *DependencyError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	depStrs := make([]string, len(e.Dependents))
	for i, d := range e.Dependents {
		depStrs[i] = d.String()
	}
	return fmt.Sprintf("cannot delete %s: %d dependent(s) exist [%s]",
		e.Resource.String(), len(e.Dependents), strings.Join(depStrs, ", "))
}

// IsDependencyError returns true if the error is a DependencyError.
func IsDependencyError(err error) bool {
	_, ok := err.(*DependencyError)
	return ok
}

// CycleError indicates that adding an edge would create a cycle in the graph.
type CycleError struct {
	From    ResourceID
	To      ResourceID
	Message string
}

// Error implements the error interface.
func (e *CycleError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("adding edge %s -> %s would create a cycle",
		e.From.String(), e.To.String())
}

// IsCycleError returns true if the error is a CycleError.
func IsCycleError(err error) bool {
	_, ok := err.(*CycleError)
	return ok
}

// NodeNotFoundError indicates that a node does not exist in the graph.
type NodeNotFoundError struct {
	Resource ResourceID
}

// Error implements the error interface.
func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("node %s not found", e.Resource.String())
}

// IsNodeNotFoundError returns true if the error is a NodeNotFoundError.
func IsNodeNotFoundError(err error) bool {
	_, ok := err.(*NodeNotFoundError)
	return ok
}

// NodeExistsError indicates that a node already exists in the graph.
type NodeExistsError struct {
	Resource ResourceID
}

// Error implements the error interface.
func (e *NodeExistsError) Error() string {
	return fmt.Sprintf("node %s already exists", e.Resource.String())
}

// IsNodeExistsError returns true if the error is a NodeExistsError.
func IsNodeExistsError(err error) bool {
	_, ok := err.(*NodeExistsError)
	return ok
}

// SchemaValidationError indicates that an operation violated schema constraints.
type SchemaValidationError struct {
	Relationship string
	Message      string
}

// Error implements the error interface.
func (e *SchemaValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("schema validation failed for relationship: %s", e.Relationship)
}

// IsSchemaValidationError returns true if the error is a SchemaValidationError.
func IsSchemaValidationError(err error) bool {
	_, ok := err.(*SchemaValidationError)
	return ok
}

// CardinalityError indicates that an operation violated cardinality constraints.
type CardinalityError struct {
	Relationship string
	Expected     Cardinality
	Message      string
}

// Error implements the error interface.
func (e *CardinalityError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("cardinality violation for %s: expected %s",
		e.Relationship, e.Expected.String())
}

// IsCardinalityError returns true if the error is a CardinalityError.
func IsCardinalityError(err error) bool {
	_, ok := err.(*CardinalityError)
	return ok
}
