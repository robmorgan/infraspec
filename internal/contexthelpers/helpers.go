package contexthelpers

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// ConfigCtxKey is the key used to store the configuration in the context.Context.
type ConfigCtxKey struct{}

// TFOptionsCtxKey is the key used to store the Terraform options in the context.Context.
type TFOptionsCtxKey struct{}

// AwsRegionCtxKey is the key used to store the AWS region in the context.Context.
type AwsRegionCtxKey struct{}

// RDSDBInstanceIDCtxKey is the key used to store the RDS DB instance ID in the context.Context.
type RDSDBInstanceIDCtxKey struct{}

// TerraformHasAppliedCtxKey is the key used to store the Terraform has applied flag in the context.Context.
type TerraformHasAppliedCtxKey struct{}

// AssertionsCtxKey is the key used to store the available assertions in the context.Context.
type AssertionsCtxKey struct{}

// UriCtxKey is the key used to store the scenario URI in the context.Context.
type UriCtxKey struct{}

// GetAsserter returns the asserter for the given provider.
func GetAsserter(ctx context.Context, provider string) (assertions.Asserter, error) {
	var a map[string]assertions.Asserter

	// get the assertions from the context
	a, exists := ctx.Value(AssertionsCtxKey{}).(map[string]assertions.Asserter)

	// if the asserter is not available, create a new map
	if !exists {
		a = make(map[string]assertions.Asserter)
	}

	// check if the asserter for the given provider already exists
	asserter, exists := a[provider]
	if exists {
		return asserter, nil
	}

	// Create new asserter based on provider
	cfg := GetConfig(ctx)
	if cfg == nil {
		return nil, fmt.Errorf("no assertions available for provider: %s", provider)
	}

	asserter, err := assertions.New(provider)
	if err != nil {
		return nil, err
	}

	return asserter, nil
}

// GetConfig returns the configuration from the context.
func GetConfig(ctx context.Context) *config.Config {
	cfg, exists := ctx.Value(ConfigCtxKey{}).(*config.Config)
	if !exists {
		return nil
	}
	return cfg
}

// GetIacProvisionerOptions returns the IaC provisioner options from the context.
func GetIacProvisionerOptions(ctx context.Context) *iacprovisioner.Options {
	opts, exists := ctx.Value(TFOptionsCtxKey{}).(*iacprovisioner.Options)
	if !exists {
		return nil
	}
	return opts
}

// SetAwsRegion sets the AWS region in the context.
func SetAwsRegion(ctx context.Context, region string) context.Context {
	return context.WithValue(ctx, AwsRegionCtxKey{}, region)
}

// GetAwsRegion returns the AWS region from the context.
func GetAwsRegion(ctx context.Context) string {
	region, exists := ctx.Value(AwsRegionCtxKey{}).(string)
	if !exists {
		return ""
	}
	return region
}

// GetRDSDBInstanceID returns the RDS DB instance ID from the context.
func GetRDSDBInstanceID(ctx context.Context) string {
	instanceID, exists := ctx.Value(RDSDBInstanceIDCtxKey{}).(string)
	if !exists {
		return ""
	}
	return instanceID
}

// SetRDSDBInstanceID sets the RDS DB instance ID in the context.
func SetRDSDBInstanceID(ctx context.Context, instanceID string) context.Context {
	return context.WithValue(ctx, RDSDBInstanceIDCtxKey{}, instanceID)
}

// GetTerraformHasApplied returns the Terraform has applied flag from the context.
func GetTerraformHasApplied(ctx context.Context) bool {
	hasApplied, exists := ctx.Value(TerraformHasAppliedCtxKey{}).(bool)
	if !exists {
		return false
	}
	return hasApplied
}

// SetTerraformHasApplied sets the Terraform has applied flag in the context.
func SetTerraformHasApplied(ctx context.Context, hasApplied bool) context.Context {
	return context.WithValue(ctx, TerraformHasAppliedCtxKey{}, hasApplied)
}

// GetUri returns the URI from the context.
func GetUri(ctx context.Context) string {
	uri, exists := ctx.Value(UriCtxKey{}).(string)
	if !exists {
		return ""
	}
	return uri
}
