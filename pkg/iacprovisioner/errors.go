package iacprovisioner

import "fmt"

// OutputKeyNotFound occurs when terraform output does not contain a value for the key
// specified in the function call
type OutputKeyNotFound string

func (err OutputKeyNotFound) Error() string {
	return fmt.Sprintf("output doesn't contain a value for the key %q", string(err))
}

// EmptyOutput is an error that occurs when an output is empty.
type EmptyOutput string

func (outputName EmptyOutput) Error() string {
	return fmt.Sprintf("Required output %s was empty", string(outputName))
}

// UnexpectedOutputType is an error that occurs when the output is not of the type we expect
type UnexpectedOutputType struct {
	Key          string
	ExpectedType string
	ActualType   string
}

func (err UnexpectedOutputType) Error() string {
	return fmt.Sprintf("Expected output '%s' to be of type '%s' but got '%s'", err.Key, err.ExpectedType, err.ActualType)
}
