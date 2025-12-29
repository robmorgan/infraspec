package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// InfraSpecScanner scans InfraSpec assertion files to extract supported operations
type InfraSpecScanner struct {
	assertionsPath string
}

// NewInfraSpecScanner creates a new InfraSpec scanner
func NewInfraSpecScanner(assertionsPath string) *InfraSpecScanner {
	return &InfraSpecScanner{
		assertionsPath: assertionsPath,
	}
}

// ScanAssertions scans InfraSpec assertion files and returns operations by service
func (s *InfraSpecScanner) ScanAssertions() (map[string][]models.InfraSpecOperation, error) {
	result := make(map[string][]models.InfraSpecOperation)

	// Find all Go files in the assertions/aws directory
	awsPath := filepath.Join(s.assertionsPath, "aws")
	if _, err := os.Stat(awsPath); os.IsNotExist(err) {
		// Try alternative path
		awsPath = s.assertionsPath
	}

	entries, err := os.ReadDir(awsPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Skip test files and the base aws.go file
		if strings.HasSuffix(entry.Name(), "_test.go") || entry.Name() == "aws.go" {
			continue
		}

		filePath := filepath.Join(awsPath, entry.Name())
		serviceName := strings.TrimSuffix(entry.Name(), ".go")

		operations, err := s.scanFile(filePath, serviceName)
		if err != nil {
			continue // Skip files we can't parse
		}

		if len(operations) > 0 {
			result[serviceName] = operations
		}
	}

	return result, nil
}

// scanFile scans a single Go file for assertion interface methods
func (s *InfraSpecScanner) scanFile(filePath, serviceName string) ([]models.InfraSpecOperation, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var operations []models.InfraSpecOperation

	// Look for interface declarations ending in "Asserter"
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for type declarations
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Check if it's an interface ending with "Asserter"
		if !strings.HasSuffix(typeSpec.Name.Name, "Asserter") {
			return true
		}

		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		// Extract methods from the interface
		for _, method := range interfaceType.Methods.List {
			if len(method.Names) == 0 {
				continue
			}

			methodName := method.Names[0].Name
			if !strings.HasPrefix(methodName, "Assert") {
				continue
			}

			// Get description from comments
			description := extractDescription(method.Doc, methodName)

			operations = append(operations, models.InfraSpecOperation{
				Name:        methodName,
				Implemented: true,
				Description: description,
			})
		}

		return true
	})

	return operations, nil
}

// extractDescription extracts a description from doc comments or generates one from the method name
func extractDescription(doc *ast.CommentGroup, methodName string) string {
	if doc != nil && len(doc.List) > 0 {
		// Get the first line of the doc comment
		text := doc.Text()
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			return strings.TrimSpace(lines[0])
		}
	}

	// Generate description from method name
	return generateDescription(methodName)
}

// generateDescription generates a human-readable description from an assertion method name
func generateDescription(methodName string) string {
	// Remove "Assert" prefix
	name := strings.TrimPrefix(methodName, "Assert")

	// Split camelCase into words using a Go-compatible approach
	// Insert space before each uppercase letter, then split
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}

	words := result.String()
	if words == "" {
		return name
	}

	return "Check " + strings.ToLower(words)
}

// GetAssertionsPath returns the path to InfraSpec assertions
// It tries to find the infraspec repository relative to the current location
func GetAssertionsPath() string {
	// Try common relative paths
	candidates := []string{
		"../infraspec/pkg/assertions",
		"../../infraspec/pkg/assertions",
		"../../../infraspec/pkg/assertions",
	}

	// Also check environment variable
	if envPath := os.Getenv("INFRASPEC_PATH"); envPath != "" {
		candidates = append([]string{filepath.Join(envPath, "pkg/assertions")}, candidates...)
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}

	// Return default path
	return "pkg/assertions"
}
