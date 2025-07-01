package assertions

import (
	"fmt"

	"github.com/robmorgan/infraspec/pkg/assertions/aws"
)

const (
	// AWS is the name of the AWS asserter
	AWS = "aws"
)

// Asserter defines the interface for all cloud resource assertions
// Provider-specific assertions must be implemented by concrete types
type Asserter interface {
	// Common assertions
	AssertExists(resourceType, resourceName string) error
	AssertTags(resourceType, resourceName string, tags map[string]string) error
}

// Factory function to create new asserters
func New(provider string) (Asserter, error) {
	switch provider {
	case "aws":
		return aws.NewAWSAsserter(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
