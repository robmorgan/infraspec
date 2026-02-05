// Package parser provides HCL parsing for Terraform configurations.
package parser

// Resource represents a parsed Terraform resource
type Resource struct {
	// Type is the resource type (e.g., "aws_s3_bucket")
	Type string

	// Name is the resource name (e.g., "my_bucket")
	Name string

	// Attributes contains the resource attributes as a nested map
	Attributes map[string]interface{}

	// Location contains the file and line information
	Location Location
}

// Location represents the source location of a resource
type Location struct {
	File   string
	Line   int
	Column int
}

// String returns a human-readable location string
func (l Location) String() string {
	return l.File + ":" + string(rune('0'+l.Line/100)) + string(rune('0'+(l.Line/10)%10)) + string(rune('0'+l.Line%10))
}

// Variable represents a Terraform variable
type Variable struct {
	Name    string
	Default interface{}
	Type    string
}

// Local represents a Terraform local value
type Local struct {
	Name  string
	Value interface{}
}

// UnknownValue is a sentinel value indicating an unresolvable variable
type UnknownValue struct{}

func (UnknownValue) String() string {
	return "<unknown>"
}

// ComputedValue is a sentinel value indicating a computed/dynamic value
type ComputedValue struct{}

func (ComputedValue) String() string {
	return "<computed>"
}

// IsUnknown checks if a value is an unknown value
func IsUnknown(v interface{}) bool {
	_, ok := v.(UnknownValue)
	return ok
}

// IsComputed checks if a value is a computed value
func IsComputed(v interface{}) bool {
	_, ok := v.(ComputedValue)
	return ok
}
