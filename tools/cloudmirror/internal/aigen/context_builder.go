package aigen

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ContextBuilder builds context for AI prompts by extracting patterns from existing code.
type ContextBuilder struct {
	servicesPath string
	sdkPath      string
}

// NewContextBuilder creates a new context builder.
func NewContextBuilder(servicesPath, sdkPath string) *ContextBuilder {
	return &ContextBuilder{
		servicesPath: servicesPath,
		sdkPath:      sdkPath,
	}
}

// BuildContext builds the prompt context for a given service and operation.
func (cb *ContextBuilder) BuildContext(ctx context.Context, service, operation string) (*PromptContext, error) {
	pc := &PromptContext{
		ExampleHandlers: make(map[string]string),
		SDKTypes:        make(map[string]string),
	}

	// Extract example handlers from IAM service (reference implementation)
	if err := cb.extractExampleHandlers(pc); err != nil {
		// Non-fatal, continue without examples
	}

	// Extract response builder patterns
	if err := cb.extractResponsePatterns(pc); err != nil {
		// Non-fatal
	}

	// Extract state patterns
	if err := cb.extractStatePatterns(pc); err != nil {
		// Non-fatal
	}

	// Extract test patterns
	if err := cb.extractTestPatterns(pc); err != nil {
		// Non-fatal
	}

	// Extract Terraform examples for Create operations
	if err := cb.extractTerraformExamples(pc); err != nil {
		// Non-fatal
	}

	return pc, nil
}

// extractExampleHandlers extracts example handler implementations from the IAM service.
func (cb *ContextBuilder) extractExampleHandlers(pc *PromptContext) error {
	iamPath := filepath.Join(cb.servicesPath, "iam", "service.go")
	content, err := os.ReadFile(iamPath)
	if err != nil {
		return err
	}

	// Parse the file to find handler functions
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, iamPath, content, parser.ParseComments)
	if err != nil {
		return err
	}

	// Look for specific handler patterns
	handlers := []string{"createRole", "putRolePolicy", "deleteRole"}
	for _, handler := range handlers {
		if code := cb.extractFunction(file, fset, content, handler); code != "" {
			pc.ExampleHandlers[handler] = code
		}
	}

	return nil
}

// extractFunction extracts a function's source code from an AST.
func (cb *ContextBuilder) extractFunction(file *ast.File, fset *token.FileSet, content []byte, funcName string) string {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Name.Name == funcName {
			start := fset.Position(fn.Pos()).Offset
			end := fset.Position(fn.End()).Offset
			if start < len(content) && end <= len(content) {
				return string(content[start:end])
			}
		}
	}
	return ""
}

// extractResponsePatterns extracts response builder usage patterns.
func (cb *ContextBuilder) extractResponsePatterns(pc *PromptContext) error {
	// Read the IAM service to find response pattern examples
	iamPath := filepath.Join(cb.servicesPath, "iam", "service.go")
	content, err := os.ReadFile(iamPath)
	if err != nil {
		return err
	}

	// Find successResponse and errorResponse patterns
	lines := strings.Split(string(content), "\n")
	var patterns []string

	for i, line := range lines {
		if strings.Contains(line, "func (s *IAMService) successResponse") ||
			strings.Contains(line, "func (s *IAMService) errorResponse") {
			// Extract the next 10 lines for context
			end := i + 10
			if end > len(lines) {
				end = len(lines)
			}
			patterns = append(patterns, strings.Join(lines[i:end], "\n"))
		}
	}

	pc.ResponseBuilderPatterns = strings.Join(patterns, "\n\n")
	return nil
}

// extractStatePatterns extracts state management patterns.
func (cb *ContextBuilder) extractStatePatterns(pc *PromptContext) error {
	// Common state patterns
	pc.StatePatterns = `// State key pattern: <service>:<resource-type>:<id>
stateKey := fmt.Sprintf("iam:roles:%s", roleName)

// Store resource
if err := s.state.Set(stateKey, role); err != nil {
    return s.errorResponse(500, "InternalFailure", "Failed to store role"), nil
}

// Retrieve resource
var role Role
if err := s.state.Get(stateKey, &role); err != nil {
    return s.errorResponse(404, "NoSuchEntity", "Role not found"), nil
}

// Check existence
if s.state.Exists(stateKey) {
    return s.errorResponse(409, "EntityAlreadyExists", "Role already exists"), nil
}

// List resources
keys, err := s.state.List("iam:roles:")
if err != nil {
    return s.errorResponse(500, "InternalFailure", "Failed to list roles"), nil
}

// Delete resource
if err := s.state.Delete(stateKey); err != nil {
    return s.errorResponse(500, "InternalFailure", "Failed to delete role"), nil
}`

	return nil
}

// extractTestPatterns extracts test patterns from existing tests.
func (cb *ContextBuilder) extractTestPatterns(pc *PromptContext) error {
	// Look for IAM service tests
	testPath := filepath.Join(cb.servicesPath, "iam", "service_test.go")
	content, err := os.ReadFile(testPath)
	if err != nil {
		// Try other services if IAM tests don't exist
		testPath = filepath.Join(cb.servicesPath, "rds", "service_test.go")
		content, err = os.ReadFile(testPath)
		if err != nil {
			return err
		}
	}

	// Extract a sample test function
	lines := strings.Split(string(content), "\n")
	var testCode []string
	inTest := false
	braceCount := 0

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "func Test") && !inTest {
			inTest = true
			braceCount = 0
		}

		if inTest {
			testCode = append(testCode, line)
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount == 0 && len(testCode) > 1 {
				break
			}
		}
	}

	pc.TestPatterns = strings.Join(testCode, "\n")
	return nil
}

// ExtractSDKTypes extracts type definitions from the AWS SDK.
func (cb *ContextBuilder) ExtractSDKTypes(service, typeName string) (string, error) {
	// Map service names to SDK paths
	sdkServiceMap := map[string]string{
		"iam":      "service/iam",
		"rds":      "service/rds",
		"ec2":      "service/ec2",
		"s3":       "service/s3",
		"dynamodb": "service/dynamodb",
		"sqs":      "service/sqs",
		"sts":      "service/sts",
	}

	sdkService, ok := sdkServiceMap[strings.ToLower(service)]
	if !ok {
		return "", nil
	}

	typesPath := filepath.Join(cb.sdkPath, sdkService, "types", "types.go")
	content, err := os.ReadFile(typesPath)
	if err != nil {
		return "", err
	}

	// Parse and find the type
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, typesPath, content, parser.ParseComments)
	if err != nil {
		return "", err
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name == typeName {
				start := fset.Position(genDecl.Pos()).Offset
				end := fset.Position(genDecl.End()).Offset
				if start < len(content) && end <= len(content) {
					return string(content[start:end]), nil
				}
			}
		}
	}

	return "", nil
}

// extractTerraformExamples extracts Terraform test examples from existing tests.
func (cb *ContextBuilder) extractTerraformExamples(pc *PromptContext) error {
	// Try to find the terraform tests directory relative to services path
	// services path is typically: internal/emulator/services
	// terraform tests are at: terraform/tests/operations
	basePath := filepath.Dir(filepath.Dir(filepath.Dir(cb.servicesPath))) // Go up from internal/emulator/services to repo root
	terraformBasePath := filepath.Join(basePath, "terraform", "tests", "operations")

	// Read IAM create_role as the reference example
	mainTfPath := filepath.Join(terraformBasePath, "iam", "create_role", "main.tf")
	if content, err := os.ReadFile(mainTfPath); err == nil {
		pc.TerraformMainExample = string(content)
	}

	testHclPath := filepath.Join(terraformBasePath, "iam", "create_role", "test.tftest.hcl")
	if content, err := os.ReadFile(testHclPath); err == nil {
		pc.TerraformTestExample = string(content)
	}

	// If IAM examples not found, try EC2 create_vpc as fallback
	if pc.TerraformMainExample == "" {
		mainTfPath = filepath.Join(terraformBasePath, "ec2", "create_vpc", "main.tf")
		if content, err := os.ReadFile(mainTfPath); err == nil {
			pc.TerraformMainExample = string(content)
		}

		testHclPath = filepath.Join(terraformBasePath, "ec2", "create_vpc", "test.tftest.hcl")
		if content, err := os.ReadFile(testHclPath); err == nil {
			pc.TerraformTestExample = string(content)
		}
	}

	return nil
}
