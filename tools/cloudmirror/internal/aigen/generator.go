package aigen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/allowlist"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/generator"
)

// ImplementationGenerator generates AWS service implementations using AI.
type ImplementationGenerator struct {
	config         Config
	claudeClient   *ClaudeClient
	contextBuilder *ContextBuilder
	validator      *CodeValidator
}

// NewImplementationGenerator creates a new implementation generator.
func NewImplementationGenerator(config Config) *ImplementationGenerator {
	return &ImplementationGenerator{
		config:         config,
		claudeClient:   NewClaudeClient(config.ClaudeAPIKey, config.ClaudeModel),
		contextBuilder: NewContextBuilder(config.ServicesPath, config.SDKPath),
		validator:      NewCodeValidator(ValidatorConfig{Verbose: config.Verbose}),
	}
}

// Generate generates implementations for all operations in the change report.
// Only services in the allowlist will be processed.
func (g *ImplementationGenerator) Generate(ctx context.Context, report *SDKChangeReport) (*GenerationResult, error) {
	result := &GenerationResult{
		Files: []GeneratedFile{},
		Summary: GenerationSummary{
			ByService:  make(map[string]int),
			ByPriority: make(map[string]int),
		},
	}

	operationsGenerated := 0

	for _, svc := range report.Services {
		// Check if we've hit the operation limit
		if g.config.MaxOperations > 0 && operationsGenerated >= g.config.MaxOperations {
			result.LimitReached = true
			if g.config.Verbose {
				fmt.Printf("Operation limit reached (%d), stopping generation\n", g.config.MaxOperations)
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

		result.ServicesProcessed++

		for _, op := range svc.NewOperations {
			// Check limit before each operation
			if g.config.MaxOperations > 0 && operationsGenerated >= g.config.MaxOperations {
				result.LimitReached = true
				break
			}

			result.Summary.TotalOperations++
			result.Summary.ByPriority[op.Priority]++

			if g.config.Verbose {
				fmt.Printf("Generating %s.%s (priority: %s)...\n", svc.Name, op.Name, op.Priority)
			}

			generated, err := g.generateOperation(ctx, svc.Name, svc.Protocol, &op)
			if err != nil {
				result.Errors = append(result.Errors, GenerationError{
					Service:   svc.Name,
					Operation: op.Name,
					Phase:     "generation",
					Message:   err.Error(),
				})
				result.Summary.FailedGeneration++
				continue
			}

			// Validate generated code
			if !g.config.DryRun {
				validationErrors := g.validateGenerated(ctx, generated)
				if len(validationErrors) > 0 {
					for _, ve := range validationErrors {
						result.Errors = append(result.Errors, GenerationError{
							Service:   svc.Name,
							Operation: op.Name,
							Phase:     "validation",
							Message:   ve,
						})
					}
					result.Summary.FailedGeneration++
					continue
				}
			}

			// Write generated files
			files, err := g.writeGenerated(generated)
			if err != nil {
				result.Errors = append(result.Errors, GenerationError{
					Service:   svc.Name,
					Operation: op.Name,
					Phase:     "write",
					Message:   err.Error(),
				})
				result.Summary.FailedGeneration++
				continue
			}

			result.Files = append(result.Files, files...)
			result.OperationsCreated++
			operationsGenerated++
			if generated.TestCode != "" {
				result.TestsCreated++
			}
			result.Summary.SuccessfullyGen++
			result.Summary.ByService[svc.Name]++
		}
	}

	return result, nil
}

func (g *ImplementationGenerator) generateOperation(ctx context.Context, service, protocol string, op *OperationInfo) (*GeneratedCode, error) {
	// Build context for the prompt
	promptContext, err := g.contextBuilder.BuildContext(ctx, service, op.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Build the prompt
	prompt := BuildHandlerPrompt(service, protocol, op, promptContext)

	// Call Claude API
	handlerCode, err := g.claudeClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Claude API error: %w", err)
	}

	// Build test prompt
	testPrompt := BuildTestPrompt(service, op.Name, handlerCode, promptContext)
	testCode, err := g.claudeClient.Generate(ctx, testPrompt)
	if err != nil {
		// Tests are optional, log but don't fail
		if g.config.Verbose {
			fmt.Printf("Warning: failed to generate tests for %s.%s: %v\n", service, op.Name, err)
		}
	}

	return &GeneratedCode{
		Service:     service,
		Operation:   op.Name,
		HandlerCode: handlerCode,
		TestCode:    testCode,
	}, nil
}

func (g *ImplementationGenerator) validateGenerated(ctx context.Context, code *GeneratedCode) []string {
	var errors []string

	// Create temp file for validation
	tmpDir, err := os.MkdirTemp("", "autotrack-validate-*")
	if err != nil {
		return []string{fmt.Sprintf("failed to create temp dir: %v", err)}
	}
	defer os.RemoveAll(tmpDir)

	// Write handler code to temp file
	handlerPath := filepath.Join(tmpDir, "handler.go")
	if err := os.WriteFile(handlerPath, []byte(code.HandlerCode), 0644); err != nil {
		return []string{fmt.Sprintf("failed to write temp file: %v", err)}
	}

	// Validate
	result, err := g.validator.ValidateFile(ctx, handlerPath)
	if err != nil {
		return []string{fmt.Sprintf("validation error: %v", err)}
	}

	if !result.Valid {
		errors = append(errors, result.CompileErrors...)
		errors = append(errors, result.VetErrors...)
		errors = append(errors, result.PatternErrors...)
	}

	return errors
}

func (g *ImplementationGenerator) writeGenerated(code *GeneratedCode) ([]GeneratedFile, error) {
	// Convert operation name to snake_case for file names
	// e.g., "DeleteScheduledAction" -> "delete_scheduled_action"
	snakeCaseOp := generator.ToSnakeCase(code.Operation)

	if g.config.DryRun {
		return []GeneratedFile{
			{
				Path:      filepath.Join(g.config.OutputDir, "services", code.Service, fmt.Sprintf("%s_handler.go", snakeCaseOp)),
				Type:      "handler",
				Service:   code.Service,
				Operation: code.Operation,
			},
		}, nil
	}

	var files []GeneratedFile

	// Create output directory
	serviceDir := filepath.Join(g.config.OutputDir, "services", code.Service)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write handler code
	handlerPath := filepath.Join(serviceDir, fmt.Sprintf("%s_handler.go", snakeCaseOp))
	if err := os.WriteFile(handlerPath, []byte(code.HandlerCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write handler: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:        handlerPath,
		Type:        "handler",
		Service:     code.Service,
		Operation:   code.Operation,
		LinesOfCode: countLines(code.HandlerCode),
	})

	// Write test code if present
	if code.TestCode != "" {
		testPath := filepath.Join(serviceDir, fmt.Sprintf("%s_handler_test.go", snakeCaseOp))
		if err := os.WriteFile(testPath, []byte(code.TestCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test: %w", err)
		}
		files = append(files, GeneratedFile{
			Path:        testPath,
			Type:        "test",
			Service:     code.Service,
			Operation:   code.Operation,
			LinesOfCode: countLines(code.TestCode),
		})
	}

	return files, nil
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
