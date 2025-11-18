package terraform

import (
	"context"
	"fmt"
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

	options, err := iacprovisioner.WithDefaultRetryableErrors(&iacprovisioner.Options{
		WorkingDir: absPath,
		Vars:       make(map[string]interface{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform options: %w", err)
	}

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
