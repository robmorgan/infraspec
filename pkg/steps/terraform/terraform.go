package terraform

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	t "github.com/robmorgan/infraspec/internal/testing"
)

// RegisterSteps registers all Terraform-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I run [Tt]erraform apply$`, newTerraformApplyStep)
	sc.Step(`^the Terraform module at "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep)
	sc.Step(`^I set variable "([^"]*)" to$`, newTerraformSetMapVariableStep)
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

	options := terraform.WithDefaultRetryableErrors(t.GetT(), &terraform.Options{
		TerraformDir: absPath,
		Vars:         make(map[string]interface{}),
	})
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformApplyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetTerraformOptions(ctx)
	out, err := terraform.InitAndApplyE(t.GetT(), options)
	if err != nil {
		return ctx, fmt.Errorf("there was an error running terraform apply: %s", out)
	}
	return contexthelpers.SetTerraformHasApplied(ctx, true), nil
}

func NewTerraformDestroyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetTerraformOptions(ctx)
	out, err := terraform.DestroyE(t.GetT(), options)
	if err != nil {
		return ctx, fmt.Errorf("there was an error running terraform destroy: %s", out)
	}
	return contexthelpers.SetTerraformHasApplied(ctx, false), nil
}

func newTerraformSetVariableStep(ctx context.Context, name, value string) (context.Context, error) {
	options := contexthelpers.GetTerraformOptions(ctx)
	options.Vars[name] = value
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformSetMapVariableStep(ctx context.Context, name string, table *godog.Table) (context.Context, error) {
	options := contexthelpers.GetTerraformOptions(ctx)

	// convert the table to a map[string]string
	varMap := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		varMap[row.Cells[0].Value] = row.Cells[1].Value
	}

	options.Vars[name] = varMap

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformOutputEqualsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetTerraformOptions(ctx)
	actualValue, err := terraform.OutputE(t.GetT(), options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s: %w", outputName, err)
	}

	if actualValue != expectedValue {
		return fmt.Errorf("expected output %s to be %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}

func newTerraformOutputContainsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetTerraformOptions(ctx)
	actualValue, err := terraform.OutputE(t.GetT(), options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s: %w", outputName, err)
	}

	// check if the expected value is a substring of the actual value
	if !strings.Contains(actualValue, expectedValue) {
		return fmt.Errorf("expected output %s to contain %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}
