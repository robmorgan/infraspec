package steps

import (
	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/pkg/steps/aws"
	"github.com/robmorgan/infraspec/pkg/steps/terraform"
)

// RegisterSteps registers all step definitions.
func RegisterSteps(sc *godog.ScenarioContext) {
	// Register Terraform steps
	terraform.RegisterSteps(sc)

	// Register provider-specific steps
	aws.RegisterSteps(sc)
}
