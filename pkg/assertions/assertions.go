package assertions

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/testing"
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
	AssertExists(t testing.TestingT, resourceType, resourceName string) error
	AssertTags(t testing.TestingT, resourceType, resourceName string, tags map[string]string) error
}

// Factory function to create new asserters
func New(provider, region string) (Asserter, error) {
	switch provider {
	case "aws":
		return aws.NewAWSAsserter(region)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
