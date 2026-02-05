// Package testfile provides types and functions for parsing .infraspec.hcl test files.
package testfile

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// TestFile represents the top-level structure of an .infraspec.hcl file.
type TestFile struct {
	SourceFile string                 // Path to the parsed file
	Variables  map[string]interface{} // Test-scoped variables
	Runs       []*Run                 // Ordered list of run blocks
}

// Run represents a single test run block.
type Run struct {
	Name    string    // Unique identifier (from block label)
	Module  string    // Path to Terraform module (required)
	State   string    // Optional path to state fixture
	Command string    // "plan" or "apply" (default: "plan")
	Asserts []*Assert // List of assertions
	Range   hcl.Range // Source range for error reporting
}

// Assert represents a single assertion within a run block.
type Assert struct {
	Condition    hcl.Expression // Parsed HCL expression
	ConditionRaw string         // Original text for display
	ErrorMessage string         // Message when assertion fails
	Range        hcl.Range      // Source range for error reporting
}

// Validate checks the TestFile for semantic errors such as duplicate run names.
func (tf *TestFile) Validate() error {
	seen := make(map[string]bool)
	for _, run := range tf.Runs {
		if seen[run.Name] {
			return fmt.Errorf("duplicate run name: %q", run.Name)
		}
		seen[run.Name] = true
	}
	return nil
}

// RunByName returns the Run with the given name, or nil if not found.
func (tf *TestFile) RunByName(name string) *Run {
	for _, run := range tf.Runs {
		if run.Name == name {
			return run
		}
	}
	return nil
}

// RunNames returns the names of all runs in order.
func (tf *TestFile) RunNames() []string {
	names := make([]string, len(tf.Runs))
	for i, run := range tf.Runs {
		names[i] = run.Name
	}
	return names
}
