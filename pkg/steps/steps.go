package steps

import (
	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/pkg/steps/aws"
	"github.com/robmorgan/infraspec/pkg/steps/terraform"
)

// StepDefinition represents a single step implementation
// type StepDefinition interface {
// 	// Pattern returns the Gherkin pattern this step matches
// 	Pattern() string

// 	// Execute runs the step implementation
// 	Execute(ctx *context.TestContext, args ...string) error
// }

// RegisterSteps registers all step definitions with Godog
func RegisterSteps(sc *godog.ScenarioContext) {
	// Register common steps
	registerCommonSteps(sc)

	// Register Terraform steps
	terraform.RegisterSteps(sc)

	// Register provider-specific steps
	aws.RegisterSteps(sc)
}

// registerCommonSteps registers common steps that are shared across providers
func registerCommonSteps(sc *godog.ScenarioContext) {
	// TODO
}

// Common step implementations
// func newRandomNameStep(ctx context.Context) func(string) error {
// 	return func(prefix string) error {
// 		name := generators.RandomResourceName(prefix, ctx.Config().Functions.RandomString)
// 		ctx.StoreValue("resource_name", name)
// 		return nil
// 	}
// }
