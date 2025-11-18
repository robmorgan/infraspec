package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/awshelpers"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// RegisterSteps registers all Terraform-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I run [Tt]erraform apply$`, newTerraformApplyStep)
	sc.Step(`^the Terraform module at "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I set the variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep)
	sc.Step(`^I set the variable "([^"]*)" to$`, newTerraformSetMapVariableStep)
	sc.Step(`^I set the variable "([^"]*)" to a random stable AWS region$`, newTerraformSetRandomStableAWSRegion)
	sc.Step(`^the "([^"]*)" output is "([^"]*)"$`, newTerraformOutputEqualsStep)
	sc.Step(`^the output "([^"]*)" should equal "([^"]*)"$`, newTerraformOutputEqualsStep)
	sc.Step(`^the output "([^"]*)" should contain "([^"]*)"$`, newTerraformOutputContainsStep)
}

func newTerraformConfigStep(ctx context.Context, path string) (context.Context, error) {
	// construct an absolute path to the terraform configuration. If the path is relative, we need to prepend the base
	// uri of the scenario otherwise, we can just use the path as is.
	var (
		absPath string
		err     error
	)
	if !filepath.IsAbs(path) {
		base := filepath.Dir(contexthelpers.GetUri(ctx))
		absPath, err = filepath.Abs(filepath.Join(base, path))
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
	}

	// Generate providers.tf for testing if AWS_ENDPOINT_URL is set
	if err := generateTestProvidersFile(absPath); err != nil {
		return nil, fmt.Errorf("failed to generate test providers file: %w", err)
	}

	options, err := iacprovisioner.WithDefaultRetryableErrors(&iacprovisioner.Options{
		WorkingDir: absPath,
		Vars:       make(map[string]interface{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform options: %w", err)
	}

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

// generateTestProvidersFile creates a providers_override.tf file in the given directory
// if AWS_ENDPOINT_URL is set, indicating we're running in test mode.
// Using _override.tf ensures it merges with existing provider configurations.
func generateTestProvidersFile(workingDir string) error {
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	// Only generate if we're in test mode (AWS_ENDPOINT_URL is set)
	if endpointURL == "" {
		return nil
	}

	// Use providers_override.tf to merge with existing provider blocks
	providersPath := filepath.Join(workingDir, "providers_override.tf")
	_ = os.Remove(providersPath) // Remove existing override file

	// Get region from environment or default to us-east-1
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	providersContent := fmt.Sprintf(`# Auto-generated file for testing - DO NOT COMMIT
# This file is generated automatically when running tests with InfraSpec API
# It merges/overrides existing AWS provider configuration for testing

provider "aws" {
  region = %q

  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true

  endpoints {
    dynamodb = %q
    sts      = %q
    s3       = %q
    rds      = %q
  }
}
`, region, endpointURL, endpointURL, endpointURL, endpointURL)

	return os.WriteFile(providersPath, []byte(providersContent), 0644)
}

func newTerraformApplyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	out, err := iacprovisioner.InitAndApply(options)
	if err != nil {
		// Include both the error message and output for better debugging
		if out != "" {
			return ctx, fmt.Errorf("there was an error running terraform apply: %s\n%s", err.Error(), out)
		}
		return ctx, fmt.Errorf("there was an error running terraform apply: %s", err.Error())
	}
	return contexthelpers.SetTerraformHasApplied(ctx, true), nil
}

func NewTerraformDestroyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	out, err := iacprovisioner.Destroy(options)
	if err != nil {
		return ctx, fmt.Errorf("there was an error running terraform destroy: %s", out)
	}
	return contexthelpers.SetTerraformHasApplied(ctx, false), nil
}

func newTerraformSetVariableStep(ctx context.Context, name, value string) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	options.Vars[name] = value
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformSetMapVariableStep(ctx context.Context, name string, table *godog.Table) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)

	// convert the table to a map[string]string
	varMap := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		varMap[row.Cells[0].Value] = row.Cells[1].Value
	}

	options.Vars[name] = varMap

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformSetRandomStableAWSRegion(ctx context.Context, name string) (context.Context, error) {
	awsRegion, err := awshelpers.GetRandomStableRegion(nil, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to get random stable AWS region: %w", err)
	}
	ctx = contexthelpers.SetAwsRegion(ctx, awsRegion)
	return newTerraformSetVariableStep(ctx, name, awsRegion)
}

func newTerraformOutputEqualsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	actualValue, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s, got %s: %w", outputName, actualValue, err)
	}

	if actualValue != expectedValue {
		return fmt.Errorf("expected output %s to be %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}

func newTerraformOutputContainsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	actualValue, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s, got %s: %w", outputName, actualValue, err)
	}

	// check if the expected value is a substring of the actual value
	if !strings.Contains(actualValue, expectedValue) {
		return fmt.Errorf("expected output %s to contain %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}
