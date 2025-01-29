package steps

import (
	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/internal/generators"
	"github.com/robmorgan/infraspec/pkg/steps/aws"
	"github.com/robmorgan/infraspec/pkg/steps/terraform"
)

// StepDefinition represents a single step implementation
type StepDefinition interface {
	// Pattern returns the Gherkin pattern this step matches
	Pattern() string

	// Execute runs the step implementation
	Execute(ctx *context.TestContext, args ...string) error
}

// RegisterSteps registers all step definitions with Godog
func RegisterSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	// Register common steps
	registerCommonSteps(ctx, sc)

	// Register Terraform steps
	terraform.RegisterSteps(ctx, sc)

	// Register provider-specific steps
	switch ctx.Config().Provider {
	case "aws":
		aws.RegisterSteps(ctx, sc)
	}
}

// registerCommonSteps registers common steps that are shared across providers
func registerCommonSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	sc.Step(`^I generate a random resource name with prefix "([^"]*)"$`, newRandomNameStep(ctx))
}

// Common step implementations
func newRandomNameStep(ctx *context.TestContext) func(string) error {
	return func(prefix string) error {
		name := generators.RandomResourceName(prefix, ctx.Config().Functions.RandomString)
		ctx.StoreValue("resource_name", name)
		return nil
	}
}
