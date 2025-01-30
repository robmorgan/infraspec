package contexthelpers

import (
	"context"
	"fmt"

	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/pkg/assertions"
)

// ConfigCtxKey is the key used to store the configuration in the context.Context.
type ConfigCtxKey struct{}

// TFOptionsCtxKey is the key used to store the Terraform options in the context.Context.
type TFOptionsCtxKey struct{}

// AssertionsCtxKey is the key used to store the available assertions in the context.Context.
type AssertionsCtxKey struct{}

// UriCtxKey is the key used to store the scenario URI in the context.Context.
type UriCtxKey struct{}

func GetAsserter(ctx context.Context) (assertions.Asserter, error) {
	asserter, exists := ctx.Value(AssertionsCtxKey{}).(assertions.Asserter)
	if !exists {
		return nil, fmt.Errorf("no asserter found in context")
	}
	return asserter, nil
}

// GetTerraformOptions returns the Terraform options from the context.
func GetTerraformOptions(ctx context.Context) *terraform.Options {
	opts, exists := ctx.Value(TFOptionsCtxKey{}).(*terraform.Options)
	if !exists {
		return nil
	}
	return opts
}

// GetUri returns the URI from the context.
func GetUri(ctx context.Context) string {
	uri, exists := ctx.Value(UriCtxKey{}).(string)
	if !exists {
		return ""
	}
	return uri
}

// func (t *TestContext) GetAsserter(provider string) (assertions.Asserter, error) {
// 	if asserter, exists := t.assertions[provider]; exists {
// 		return asserter, nil
// 	}

// 	// Create new asserter based on provider
// 	asserter, err := assertions.New(provider, t.config.DefaultRegion)
// 	if err != nil {
// 		return nil, err
// 	}

// 	t.assertions[provider] = asserter
// 	return asserter, nil
// }
