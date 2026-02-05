package aigen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/allowlist"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/generator"
)

// PromptGeneratorConfig holds configuration for the prompt generator.
type PromptGeneratorConfig struct {
	SDKPath       string
	ServicesPath  string
	OutputDir     string
	MaxOperations int
	Verbose       bool
}

// PromptGenerator prepares prompts for use with Claude Code Action.
type PromptGenerator struct {
	config         PromptGeneratorConfig
	contextBuilder *ContextBuilder
}

// PromptResult represents the result of prompt preparation.
type PromptResult struct {
	Prompts      []PromptInfo `json:"prompts"`
	LimitReached bool         `json:"limit_reached,omitempty"`
	ManifestPath string       `json:"manifest_path"`
}

// PromptInfo contains information about a generated prompt.
type PromptInfo struct {
	Service       string `json:"service"`
	Operation     string `json:"operation"`
	Protocol      string `json:"protocol"`
	Priority      string `json:"priority"`
	PromptFile    string `json:"prompt_file"`
	HandlerPath   string `json:"handler_path"`
	TestPath      string `json:"test_path"`
	TerraformPath string `json:"terraform_path,omitempty"` // Path to main.tf for Create operations
	TfTestPath    string `json:"tftest_path,omitempty"`    // Path to test.tftest.hcl for Create operations
}

// isCreateOperation checks if the operation name is a Create operation
func isCreateOperation(opName string) bool {
	return strings.HasPrefix(opName, "Create")
}

// PromptManifest is the manifest file for Claude Code Action.
type PromptManifest struct {
	Version    string       `json:"version"`
	Operations []PromptInfo `json:"operations"`
}

// NewPromptGenerator creates a new prompt generator.
func NewPromptGenerator(config PromptGeneratorConfig) *PromptGenerator {
	return &PromptGenerator{
		config:         config,
		contextBuilder: NewContextBuilder(config.ServicesPath, config.SDKPath),
	}
}

// PreparePrompts creates prompt files for all operations in the change report.
func (g *PromptGenerator) PreparePrompts(ctx context.Context, report *SDKChangeReport) (*PromptResult, error) {
	result := &PromptResult{
		Prompts: []PromptInfo{},
	}

	// Create output directory
	if err := os.MkdirAll(g.config.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	operationsProcessed := 0

	for _, svc := range report.Services {
		// Check if we've hit the operation limit
		if g.config.MaxOperations > 0 && operationsProcessed >= g.config.MaxOperations {
			result.LimitReached = true
			if g.config.Verbose {
				fmt.Printf("Operation limit reached (%d), stopping\n", g.config.MaxOperations)
			}
			break
		}

		// Check if service is in the allowlist
		if !allowlist.IsServiceAllowed(svc.Name) {
			if g.config.Verbose {
				fmt.Printf("Skipping service %s (not in allowlist)\n", svc.Name)
			}
			continue
		}

		if len(svc.NewOperations) == 0 {
			continue
		}

		for _, op := range svc.NewOperations {
			// Check limit before each operation
			if g.config.MaxOperations > 0 && operationsProcessed >= g.config.MaxOperations {
				result.LimitReached = true
				break
			}

			if g.config.Verbose {
				fmt.Printf("Preparing prompt for %s.%s...\n", svc.Name, op.Name)
			}

			// Build the prompt
			promptContent, err := g.buildPromptFile(ctx, svc.Name, svc.Protocol, &op)
			if err != nil {
				if g.config.Verbose {
					fmt.Printf("Warning: failed to build prompt for %s.%s: %v\n", svc.Name, op.Name, err)
				}
				continue
			}

			// Write prompt file
			promptFileName := fmt.Sprintf("%s_%s.md", svc.Name, op.Name)
			promptPath := filepath.Join(g.config.OutputDir, promptFileName)
			if err := os.WriteFile(promptPath, []byte(promptContent), 0o644); err != nil {
				return nil, fmt.Errorf("failed to write prompt file: %w", err)
			}

			// Calculate output paths (using snake_case for file names)
			snakeCaseOp := generator.ToSnakeCase(op.Name)
			handlerPath := fmt.Sprintf("internal/emulator/services/%s/%s_handler.go", svc.Name, snakeCaseOp)
			testPath := fmt.Sprintf("internal/emulator/services/%s/%s_handler_test.go", svc.Name, snakeCaseOp)

			promptInfo := PromptInfo{
				Service:     svc.Name,
				Operation:   op.Name,
				Protocol:    svc.Protocol,
				Priority:    op.Priority,
				PromptFile:  promptFileName,
				HandlerPath: handlerPath,
				TestPath:    testPath,
			}

			// Add Terraform paths for Create operations
			if isCreateOperation(op.Name) {
				promptInfo.TerraformPath = fmt.Sprintf("terraform/tests/operations/%s/%s/main.tf", svc.Name, snakeCaseOp)
				promptInfo.TfTestPath = fmt.Sprintf("terraform/tests/operations/%s/%s/test.tftest.hcl", svc.Name, snakeCaseOp)
			}

			result.Prompts = append(result.Prompts, promptInfo)

			operationsProcessed++
		}
	}

	// Write manifest file
	manifest := PromptManifest{
		Version:    "1.0",
		Operations: result.Prompts,
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(g.config.OutputDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}
	result.ManifestPath = manifestPath

	return result, nil
}

// buildPromptFile creates the full prompt content for an operation.
func (g *PromptGenerator) buildPromptFile(ctx context.Context, service, protocol string, op *OperationInfo) (string, error) {
	var sb strings.Builder

	// Build context for the prompt
	promptContext, err := g.contextBuilder.BuildContext(ctx, service, op.Name)
	if err != nil {
		// Non-fatal: proceed without context
		promptContext = &PromptContext{
			ExampleHandlers: make(map[string]string),
			SDKTypes:        make(map[string]string),
		}
	}

	// System prompt section
	sb.WriteString(`# AWS Handler Implementation Task

You are implementing an AWS service handler for the InfraSpec API emulator.
This emulator provides AWS-compatible endpoints for testing Terraform configurations.

## Important Guidelines

1. **Use Response Builders**: Always use successResponse() and errorResponse() helpers
2. **Create Result Types**: For Query protocol, create Result types with XMLName
3. **State Management**: Use state key pattern: <service>:<resource-type>:<id>
4. **Validate Parameters**: Check required parameters first
5. **AWS Error Codes**: Use appropriate AWS error codes (InvalidParameterValue, ResourceNotFound, etc.)
6. **Graph Registration**: Register resources in the graph for dependency tracking

`)

	// Add the handler prompt content (reuse existing function)
	handlerPrompt := BuildHandlerPrompt(service, protocol, op, promptContext)
	sb.WriteString(handlerPrompt)

	// Add output instructions (using snake_case for file names)
	snakeCaseOp := generator.ToSnakeCase(op.Name)
	sb.WriteString(fmt.Sprintf(`
## Output Instructions

After generating the code:

1. **Handler File**: Write the handler code to:
   internal/emulator/services/%s/%s_handler.go

2. **Test File**: Write comprehensive tests to:
   internal/emulator/services/%s/%s_handler_test.go

3. **Integration**: If this is a new handler function, add it to the service's HandleRequest switch statement in service.go

Make sure to:
- Include proper package declaration
- Include all necessary imports
- Follow existing patterns in the service
- Write tests that cover success, validation errors, and not-found cases
`, service, snakeCaseOp, service, snakeCaseOp))

	// Add Terraform test instructions for Create operations
	if isCreateOperation(op.Name) && (promptContext.TerraformMainExample != "" || promptContext.TerraformTestExample != "") {
		sb.WriteString(fmt.Sprintf(`
## Terraform Test Files

For this Create operation, also generate Terraform test files to validate the API works correctly with Terraform.

4. **Terraform Config**: Write to:
   terraform/tests/operations/%s/%s/main.tf

5. **Terraform Test**: Write to:
   terraform/tests/operations/%s/%s/test.tftest.hcl

### Terraform Requirements

The main.tf should:
- Configure the AWS provider with the emulator endpoint variable
- Create a minimal test resource using the Terraform AWS provider resource type
- Output key attributes that should be verified

The test.tftest.hcl should:
- Run terraform apply
- Assert that key attributes are set correctly
- Use simple assertions (non-empty values, expected values)

`, service, snakeCaseOp, service, snakeCaseOp))

		// Add example main.tf if available
		if promptContext.TerraformMainExample != "" {
			sb.WriteString("### Example main.tf (adapt for this service/resource):\n\n```hcl\n")
			sb.WriteString(promptContext.TerraformMainExample)
			sb.WriteString("\n```\n\n")
		}

		// Add example test.tftest.hcl if available
		if promptContext.TerraformTestExample != "" {
			sb.WriteString("### Example test.tftest.hcl (adapt for this service/resource):\n\n```hcl\n")
			sb.WriteString(promptContext.TerraformTestExample)
			sb.WriteString("\n```\n\n")
		}

		sb.WriteString(`### Key Points for Terraform Tests:
- Use the correct provider endpoint for the service (e.g., iam, ec2, rds, etc.)
- Use random_id for unique resource names to avoid conflicts
- Include appropriate tags
- Export outputs that can be verified in the test
- Keep the test assertions simple and focused on verifying the API response
`)
	}

	return sb.String(), nil
}
