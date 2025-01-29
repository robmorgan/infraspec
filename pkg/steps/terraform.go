package steps

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/robmorgan/infraspec/internal/context"
)

func registerTerraformSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	sc.Step(`^I run Terraform apply$`, newTerraformApplyStep(ctx))
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep(ctx))
	sc.Step(`^I generate a random resource name with prefix "([^"]*)"$`, (&RandomNameStep{}).Execute)
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)"$`, (&SetVariableStep{}).Execute)
	sc.Step(`^the output "([^"]*)" should contain "([^"]*)"$`, (&TerraformOutput{}).Execute)
	//sc.Step(`^I expect the
}

func newTerraformConfigStep(ctx *context.TestContext) func(string) error {
	return func(path string) error {
		opts := terraform.WithDefaultRetryableErrors(GetT(), &terraform.Options{
			TerraformDir: path,
			Vars:         make(map[string]interface{}),
		})
		ctx.SetTerraformOptions(opts)
		return nil
	}
}

func newTerraformApplyStep(ctx *context.TestContext) func() error {
	return func() error {
		out, err := terraform.InitAndApplyE(GetT(), ctx.GetTerraformOptions())
		if err != nil {
			return fmt.Errorf("there was an error running terraform apply: %s", out)
		}
		//s.hasApplied = true
		// ctx.Terraform.Applied = true
		return nil
	}
}

type TerraformOutput struct{}

func (s *TerraformOutput) Pattern() string {
	return `^the output "([^"]*)" should contain "([^"]*)"$`
}

func (s *TerraformOutput) Execute(ctx *context.TestContext, args ...string) error {
	// TODO - implement
	return nil
}
