package terraform

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robmorgan/infraspec/internal/context"
)

// SetVariableStep handles setting variables in the test context
type SetVariableStep struct {
	ctx *context.TestContext
}

// NewSetVariableStep creates a new SetVariableStep
func NewSetVariableStep(ctx *context.TestContext) *SetVariableStep {
	return &SetVariableStep{
		ctx: ctx,
	}
}

// Pattern returns the Gherkin pattern for this step
func (s *SetVariableStep) Pattern() string {
	return `^I set variable "([^"]*)" to "([^"]*)"$`
}

// Execute runs the step implementation
func (s *SetVariableStep) Execute(args ...string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name := args[0]
	value := args[1]

	// Interpolate any variables in the value
	interpolatedValue, err := s.interpolateValue(value)
	if err != nil {
		return fmt.Errorf("failed to interpolate value: %w", err)
	}

	// Store in the IaC provisioner options
	if s.ctx.GetIacProvisionerOptions() != nil {
		s.ctx.GetIacProvisionerOptions().Vars[name] = interpolatedValue
	}

	// Also store in context values for future reference
	s.ctx.StoreValue(name, interpolatedValue)

	return nil
}

// interpolateValue replaces variables and environment variables in the value
func (s *SetVariableStep) interpolateValue(value string) (string, error) {
	// First, handle stored variables ${variable}
	varRegex := regexp.MustCompile(`\${([^}]+)}`)
	result := varRegex.ReplaceAllStringFunc(value, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		if storedValue, exists := s.ctx.GetStoredValues()[varName]; exists {
			return storedValue
		}
		// If not found, leave as is
		return match
	})

	// Then, handle environment variables %{ENV_VAR}
	envRegex := regexp.MustCompile(`%{([^}]+)}`)
	result = envRegex.ReplaceAllStringFunc(result, func(match string) string {
		envName := match[2 : len(match)-1] // Remove %{ and }
		if envValue, exists := os.LookupEnv(envName); exists {
			return envValue
		}
		// If not found, leave as is
		return match
	})

	// Check if there are any unresolved variables
	if strings.Contains(result, "${") || strings.Contains(result, "%{") {
		// Find all unresolved variables
		var unresolvedVars []string

		varMatches := varRegex.FindAllString(result, -1)
		for _, match := range varMatches {
			varName := match[2 : len(match)-1]
			unresolvedVars = append(unresolvedVars, fmt.Sprintf("${%s}", varName))
		}

		envMatches := envRegex.FindAllString(result, -1)
		for _, match := range envMatches {
			envName := match[2 : len(match)-1]
			unresolvedVars = append(unresolvedVars, fmt.Sprintf("%%{%s}", envName))
		}

		if len(unresolvedVars) > 0 {
			return "", fmt.Errorf("unresolved variables: %s", strings.Join(unresolvedVars, ", "))
		}
	}

	return result, nil
}
