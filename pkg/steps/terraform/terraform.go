package terraform

import (
	"fmt"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/internal/context"
	t "github.com/robmorgan/infraspec/internal/testing"
)

// RegisterSteps registers all Terraform-specific step definitions
func RegisterSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	sc.Step(`^I run Terraform apply$`, newTerraformApplyStep(ctx))
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep(ctx))
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep(ctx))
	sc.Step(`^the output "([^"]*)" should contain "([^"]*)"$`, newTerraformOutputContainsStep(ctx))
}

func newTerraformConfigStep(ctx *context.TestContext) func(string) error {
	return func(path string) error {
		// construct an absolute path to the terraform configuration.
		// if the path is relative, we need to prepend the base uri of the scenario otherwise, we can just use the path
		// as is.
		if !filepath.IsAbs(path) {
			uri := ctx.GetScenarioUri()
			path, err := filepath.Abs(filepath.Join(uri, path))
			if err != nil {
				return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
			}
		}

		opts := terraform.WithDefaultRetryableErrors(t.GetT(), &terraform.Options{
			TerraformDir: path,
			Vars:         make(map[string]interface{}),
		})
		ctx.SetTerraformOptions(opts)
		return nil
	}
}

func newTerraformApplyStep(ctx *context.TestContext) func() error {
	return func() error {
		out, err := terraform.InitAndApplyE(t.GetT(), ctx.GetTerraformOptions())
		if err != nil {
			return fmt.Errorf("there was an error running terraform apply: %s", out)
		}
		//s.hasApplied = true
		// ctx.Terraform.Applied = true
		return nil
	}
}

func newTerraformSetVariableStep(ctx *context.TestContext) func(string, string) error {
	return func(name, value string) error {
		ctx.GetTerraformOptions().Vars[name] = value
		return nil
	}
}

func newTerraformOutputContainsStep(ctx *context.TestContext) func(string, string) error {
	return func(outputName, expectedValue string) error {
		actualValue, err := terraform.OutputE(t.GetT(), ctx.GetTerraformOptions(), outputName)
		if err != nil {
			return fmt.Errorf("failed to get output %s: %w", outputName, err)
		}

		if actualValue != expectedValue {
			return fmt.Errorf("expected output %s to be %s, got %s", outputName, expectedValue, actualValue)
		}
		return nil
	}
}
