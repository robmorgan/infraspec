package aigen

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CodeValidator validates generated Go code.
type CodeValidator struct {
	config ValidatorConfig
}

// NewCodeValidator creates a new code validator.
func NewCodeValidator(config ValidatorConfig) *CodeValidator {
	return &CodeValidator{config: config}
}

// ValidateFile validates a Go source file.
func (v *CodeValidator) ValidateFile(ctx context.Context, path string) (*ValidationResult, error) {
	result := &ValidationResult{
		FilePath: path,
		Valid:    true,
	}

	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		result.Valid = false
		result.CompileErrors = append(result.CompileErrors, fmt.Sprintf("failed to read file: %v", err))
		return result, nil
	}

	// Check for banned patterns
	patternErrors := v.checkBannedPatterns(string(content))
	if len(patternErrors) > 0 {
		result.Valid = false
		result.PatternErrors = patternErrors
	}

	// Parse syntax
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, path, content, parser.AllErrors)
	if err != nil {
		result.Valid = false
		result.CompileErrors = append(result.CompileErrors, fmt.Sprintf("syntax error: %v", err))
	}

	// Run go vet if syntax is valid
	if len(result.CompileErrors) == 0 {
		vetErrors := v.runGoVet(ctx, path)
		if len(vetErrors) > 0 {
			result.VetErrors = vetErrors
			// vet errors are warnings, not fatal
		}
	}

	// Check for recommended patterns
	warnings := v.checkRecommendedPatterns(string(content))
	if len(warnings) > 0 {
		result.Warnings = warnings
	}

	return result, nil
}

// ValidateCode validates Go source code string.
func (v *CodeValidator) ValidateCode(ctx context.Context, code string) (*ValidationResult, error) {
	// Create temp file
	tmpDir, err := os.MkdirTemp("", "autotrack-validate-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "code.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	return v.ValidateFile(ctx, tmpFile)
}

// checkBannedPatterns checks for patterns that should not appear in generated code.
func (v *CodeValidator) checkBannedPatterns(code string) []string {
	var errors []string

	// Banned patterns from CLAUDE.md
	bannedPatterns := []struct {
		pattern string
		message string
	}{
		{
			pattern: `fmt\.Sprintf\s*\(\s*` + "`" + `<\?xml`,
			message: "Manual XML construction detected. Use response builders instead.",
		},
		{
			pattern: `fmt\.Sprintf\s*\(\s*` + "`" + `<.*Response>`,
			message: "Manual XML response construction detected. Use successResponse() helper.",
		},
		{
			pattern: `xml\.MarshalIndent\s*\(`,
			message: "Direct xml.MarshalIndent usage detected. Use response builders.",
		},
		{
			pattern: `responseXML\s*:=\s*fmt\.Sprintf`,
			message: "Manual XML response construction detected. Use response builders.",
		},
		{
			pattern: `errorXML\s*:=\s*fmt\.Sprintf`,
			message: "Manual XML error construction detected. Use errorResponse() helper.",
		},
	}

	for _, bp := range bannedPatterns {
		re := regexp.MustCompile(bp.pattern)
		if re.MatchString(code) {
			errors = append(errors, bp.message)
		}
	}

	return errors
}

// checkRecommendedPatterns checks for recommended patterns that should appear.
func (v *CodeValidator) checkRecommendedPatterns(code string) []string {
	var warnings []string

	// Check for successResponse usage in handlers
	if strings.Contains(code, "func (s *") && strings.Contains(code, "HandleRequest") {
		if !strings.Contains(code, "successResponse") && !strings.Contains(code, "errorResponse") {
			warnings = append(warnings, "Handler doesn't use successResponse/errorResponse helpers")
		}
	}

	// Check for state key pattern
	if strings.Contains(code, "s.state.Set") || strings.Contains(code, "s.state.Get") {
		if !strings.Contains(code, `fmt.Sprintf("`) {
			warnings = append(warnings, "State operations should use fmt.Sprintf for key construction")
		}
	}

	// Check for parameter validation
	if strings.Contains(code, "params map[string]interface{}") {
		if !strings.Contains(code, "getStringValue") && !strings.Contains(code, `params["`) {
			warnings = append(warnings, "Handler should extract and validate parameters")
		}
	}

	return warnings
}

// runGoVet runs go vet on the file.
func (v *CodeValidator) runGoVet(ctx context.Context, path string) []string {
	cmd := exec.CommandContext(ctx, "go", "vet", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Parse vet output for specific issues
		lines := strings.Split(string(output), "\n")
		var errors []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				errors = append(errors, line)
			}
		}
		return errors
	}
	return nil
}

// ValidateServiceFile validates a complete service file.
func (v *CodeValidator) ValidateServiceFile(ctx context.Context, path string) (*ValidationResult, error) {
	result, err := v.ValidateFile(ctx, path)
	if err != nil {
		return nil, err
	}

	// Additional service-specific checks
	content, err := os.ReadFile(path)
	if err != nil {
		return result, nil
	}

	code := string(content)

	// Check for required interface implementations
	if !strings.Contains(code, "ServiceName()") {
		result.Warnings = append(result.Warnings, "Service should implement ServiceName() method")
	}

	if !strings.Contains(code, "HandleRequest(") {
		result.Warnings = append(result.Warnings, "Service should implement HandleRequest() method")
	}

	// Check for SupportedActions (recommended)
	if !strings.Contains(code, "SupportedActions()") {
		result.Warnings = append(result.Warnings, "Service should implement SupportedActions() for routing")
	}

	return result, nil
}

// ValidateTestFile validates a test file.
func (v *CodeValidator) ValidateTestFile(ctx context.Context, path string) (*ValidationResult, error) {
	result, err := v.ValidateFile(ctx, path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return result, nil
	}

	code := string(content)

	// Check for required test patterns
	if !strings.Contains(code, "testing.T") {
		result.Warnings = append(result.Warnings, "Test file should use *testing.T")
	}

	if !strings.Contains(code, "emulator.NewMemoryStateManager()") {
		result.Warnings = append(result.Warnings, "Tests should use emulator.NewMemoryStateManager()")
	}

	// Check for at least one test function
	if !strings.Contains(code, "func Test") {
		result.Warnings = append(result.Warnings, "Test file should contain at least one Test function")
	}

	return result, nil
}

// CompileCheck attempts to compile the code in a temporary module.
func (v *CodeValidator) CompileCheck(ctx context.Context, code string, imports []string) error {
	tmpDir, err := os.MkdirTemp("", "autotrack-compile-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal go.mod
	goMod := `module tempcheck

go 1.24

require github.com/robmorgan/infraspec v0.0.0

replace github.com/robmorgan/infraspec => ../../..
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Write the code
	if err := os.WriteFile(filepath.Join(tmpDir, "check.go"), []byte(code), 0o644); err != nil {
		return fmt.Errorf("failed to write code: %w", err)
	}

	// Run go build
	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %s", string(output))
	}

	return nil
}
