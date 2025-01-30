package terraform

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	t "github.com/robmorgan/infraspec/internal/testing"
)

// RegisterSteps registers all Terraform-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I run Terraform apply$`, newTerraformApplyStep)
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep)
	sc.Step(`^the output "([^"]*)" should contain "([^"]*)"$`, newTerraformOutputContainsStep)
}

func newTerraformConfigStep(ctx context.Context, path string) (context.Context, error) {
	// construct an absolute path to the terraform configuration.
	// if the path is relative, we need to prepend the base uri of the scenario otherwise, we can just use the path
	// as is.
	if !filepath.IsAbs(path) {
		uri := contexthelpers.GetUri(ctx)
		path, err := filepath.Abs(filepath.Join(uri, path))
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
	}

	options := terraform.WithDefaultRetryableErrors(t.GetT(), &terraform.Options{
		TerraformDir: path,
		Vars:         make(map[string]interface{}),
	})
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformApplyStep(ctx context.Context) error {
	options := contexthelpers.GetTerraformOptions(ctx)
	out, err := terraform.InitAndApplyE(t.GetT(), options)
	if err != nil {
		return fmt.Errorf("there was an error running terraform apply: %s", out)
	}
	//s.hasApplied = true
	// ctx.Terraform.Applied = true
	return nil
}

func newTerraformSetVariableStep(ctx context.Context, name, value string) (context.Context, error) {
	options := contexthelpers.GetTerraformOptions(ctx)
	options.Vars[name] = value
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformOutputContainsStep(ctx context.Context, outputName, expectedValue string) error {
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
