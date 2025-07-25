package assertions

import (
	"fmt"

	"github.com/robmorgan/infraspec/pkg/assertions/aws"
	"github.com/robmorgan/infraspec/pkg/assertions/http"
)

const (
	// AWS is the name of the AWS asserter
	AWS = "aws"
	// HTTP is the name of the HTTP asserter
	HTTP = "http"
)

// Asserter defines the interface for all provider assertions
// Provider-specific assertions must be implemented by concrete types
type Asserter interface {
	GetName() string
}

// Factory function to create new asserters
func New(provider string) (Asserter, error) {
	switch provider {
	case "aws":
		return aws.NewAWSAsserter(), nil
	case "http":
		return http.NewHTTPAsserter(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
