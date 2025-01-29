package steps

import (
	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/internal/generators"
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
	registerTerraformSteps(ctx, sc)

	// Register provider-specific steps
	switch ctx.Config().Provider {
	case "aws":
		registerAWSSteps(ctx, sc)
	}
}

// registerCommonSteps registers common steps that are shared across providers
func registerCommonSteps(ctx *context.TestContext, sc *godog.ScenarioContext) {
	sc.Step(`^I generate a random resource name with prefix "([^"]*)"$`, (&RandomNameStep{}).Execute)
}

// Common step implementations
type RandomNameStep struct{}

func (s *RandomNameStep) Pattern() string {
	return `^I generate a random resource name with prefix "([^"]*)"$`
}

func (s *RandomNameStep) Execute(ctx *context.TestContext, args ...string) error {
	prefix := args[0]
	name := generators.RandomResourceName(prefix, ctx.Config().Functions.RandomString)
	ctx.StoreValue("resource_name", name)
	return nil
}
