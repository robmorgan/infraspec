package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/config"
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

	options, err := iacprovisioner.WithDefaultRetryableErrors(&iacprovisioner.Options{
		WorkingDir: absPath,
		Vars:       make(map[string]interface{}),
		EnvVars:    make(map[string]string),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform options: %w", err)
	}

	// Set AWS endpoint environment variables when virtual cloud is enabled
	configureVirtualCloudEndpoints(options)

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
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

// configureVirtualCloudEndpoints sets AWS endpoint environment variables when
// InfraSpec Virtual Cloud is enabled. This configures Terraform/OpenTofu to use
// the InfraSpec Cloud API endpoints instead of real AWS.
//
// The function sets service-specific AWS_ENDPOINT_URL_* environment variables
// that are automatically recognized by the AWS provider in Terraform/OpenTofu.
// See: https://search.opentofu.org/provider/opentofu/aws/v6.1.0/docs/guides/custom-service-endpoints
func configureVirtualCloudEndpoints(options *iacprovisioner.Options) {
	if !config.UseInfraspecVirtualCloud() {
		return
	}

	// Get the base endpoint URL (defaults to InfraSpec Cloud API if not set)
	endpoint, ok := awshelpers.GetVirtualCloudEndpoint("")
	if !ok {
		return
	}

	// Check if AWS_ENDPOINT_URL is already set in the environment
	// If it is, we don't need to set the service-specific ones
	if existingEndpoint := os.Getenv("AWS_ENDPOINT_URL"); existingEndpoint != "" {
		config.Logging.Logger.Infof("AWS_ENDPOINT_URL already set to: %s", existingEndpoint)
		return
	}

	// List of AWS services that InfraSpec supports
	// These will be set as AWS_ENDPOINT_URL_<SERVICE> environment variables
	awsServices := []string{
		"DYNAMODB",
		"EC2",
		"RDS",
		"S3",
		"STS",
		"SSM",
	}

	config.Logging.Logger.Infof("Configuring virtual cloud endpoints for Terraform/OpenTofu to use: %s", endpoint)

	// Set service-specific endpoint environment variables
	for _, service := range awsServices {
		envVar := fmt.Sprintf("AWS_ENDPOINT_URL_%s", service)

		// Check if a service-specific endpoint is already set in the environment
		if existingServiceEndpoint := os.Getenv(envVar); existingServiceEndpoint != "" {
			config.Logging.Logger.Debugf("%s already set to: %s", envVar, existingServiceEndpoint)
			options.EnvVars[envVar] = existingServiceEndpoint
			continue
		}

		options.EnvVars[envVar] = endpoint
		config.Logging.Logger.Debugf("Setting %s=%s", envVar, endpoint)
	}

	// Also set credentials configuration to skip AWS credential validation
	// These are required when using custom endpoints
	options.EnvVars["AWS_ACCESS_KEY_ID"] = awshelpers.InfraspecCloudAccessKeyID

	// Get the InfraSpec Cloud token
	token, err := config.GetInfraspecCloudToken()
	if err == nil && token != "" {
		options.EnvVars["AWS_SECRET_ACCESS_KEY"] = token
	}
}
