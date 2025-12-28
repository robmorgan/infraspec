// Package terraform provides Terraform test execution functionality.
package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Config holds configuration for the Terraform test runner.
type Config struct {
	TestsPath string
	Endpoint  string
	Filter    string
	Verbose   bool
}

// TestRunner executes Terraform tests against the emulator.
type TestRunner struct {
	config Config
}

// TestResults holds the results of running all tests.
type TestResults struct {
	TotalTests   int            `json:"total_tests"`
	Passed       int            `json:"passed"`
	Failed       int            `json:"failed"`
	Skipped      int            `json:"skipped"`
	Duration     time.Duration  `json:"duration"`
	Tests        []TestResult   `json:"tests"`
	ByService    map[string]int `json:"by_service"`
	ByOperation  map[string]int `json:"by_operation"`
}

// TestResult holds the result of a single test.
type TestResult struct {
	Service   string        `json:"service"`
	Operation string        `json:"operation"`
	TestFile  string        `json:"test_file"`
	Status    string        `json:"status"` // passed, failed, skipped
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Output    string        `json:"output,omitempty"`
}

// NewTestRunner creates a new Terraform test runner.
func NewTestRunner(config Config) *TestRunner {
	return &TestRunner{config: config}
}

// RunAll runs all Terraform tests.
func (r *TestRunner) RunAll(ctx context.Context) (*TestResults, error) {
	start := time.Now()
	results := &TestResults{
		Tests:       []TestResult{},
		ByService:   make(map[string]int),
		ByOperation: make(map[string]int),
	}

	// Find all test directories
	operationsDir := filepath.Join(r.config.TestsPath, "operations")
	services, err := os.ReadDir(operationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read operations directory: %w", err)
	}

	for _, service := range services {
		if !service.IsDir() {
			continue
		}

		// Apply filter if specified
		if r.config.Filter != "" && !strings.Contains(service.Name(), r.config.Filter) {
			continue
		}

		serviceDir := filepath.Join(operationsDir, service.Name())
		operations, err := os.ReadDir(serviceDir)
		if err != nil {
			continue
		}

		for _, operation := range operations {
			if !operation.IsDir() {
				continue
			}

			// Apply operation filter if specified
			if r.config.Filter != "" &&
				!strings.Contains(service.Name(), r.config.Filter) &&
				!strings.Contains(operation.Name(), r.config.Filter) {
				continue
			}

			testDir := filepath.Join(serviceDir, operation.Name())
			result := r.runTest(ctx, service.Name(), operation.Name(), testDir)

			results.Tests = append(results.Tests, result)
			results.TotalTests++

			switch result.Status {
			case "passed":
				results.Passed++
				results.ByService[service.Name()]++
				results.ByOperation[operation.Name()]++
			case "failed":
				results.Failed++
			case "skipped":
				results.Skipped++
			}
		}
	}

	results.Duration = time.Since(start)
	return results, nil
}

// runTest runs a single Terraform test.
func (r *TestRunner) runTest(ctx context.Context, service, operation, testDir string) TestResult {
	start := time.Now()
	result := TestResult{
		Service:   service,
		Operation: operation,
		TestFile:  testDir,
	}

	// Check if test.tftest.hcl exists
	testFile := filepath.Join(testDir, "test.tftest.hcl")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		result.Status = "skipped"
		result.Error = "no test.tftest.hcl file"
		result.Duration = time.Since(start)
		return result
	}

	// Initialize Terraform
	initCmd := exec.CommandContext(ctx, "terraform", "init", "-input=false", "-no-color")
	initCmd.Dir = testDir
	initCmd.Env = r.buildEnv()

	if output, err := initCmd.CombinedOutput(); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("terraform init failed: %v", err)
		result.Output = string(output)
		result.Duration = time.Since(start)
		return result
	}

	// Run Terraform test
	testCmd := exec.CommandContext(ctx, "terraform", "test", "-no-color")
	testCmd.Dir = testDir
	testCmd.Env = r.buildEnv()

	output, err := testCmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
	} else {
		result.Status = "passed"
	}

	return result
}

// buildEnv builds the environment variables for Terraform.
func (r *TestRunner) buildEnv() []string {
	env := os.Environ()

	// Add emulator endpoint
	env = append(env, fmt.Sprintf("TF_VAR_emulator_endpoint=%s", r.config.Endpoint))

	// Set AWS endpoint URLs for each service
	services := []string{"S3", "RDS", "EC2", "IAM", "STS", "DYNAMODB", "SQS"}
	for _, svc := range services {
		env = append(env, fmt.Sprintf("AWS_ENDPOINT_URL_%s=%s", svc, r.config.Endpoint))
	}

	// Set dummy AWS credentials
	env = append(env, "AWS_ACCESS_KEY_ID=test")
	env = append(env, "AWS_SECRET_ACCESS_KEY=test")
	env = append(env, "AWS_DEFAULT_REGION=us-east-1")

	return env
}

// RunSingle runs a single Terraform test by path.
func (r *TestRunner) RunSingle(ctx context.Context, testPath string) (*TestResult, error) {
	// Extract service and operation from path
	parts := strings.Split(testPath, string(filepath.Separator))
	var service, operation string

	for i, part := range parts {
		if part == "operations" && i+2 < len(parts) {
			service = parts[i+1]
			operation = parts[i+2]
			break
		}
	}

	result := r.runTest(ctx, service, operation, testPath)
	return &result, nil
}

// GenerateTestReport generates a markdown report from test results.
func (r *TestRunner) GenerateTestReport(results *TestResults) string {
	var sb strings.Builder

	sb.WriteString("# Terraform Validation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Total Tests:** %d\n", results.TotalTests))
	sb.WriteString(fmt.Sprintf("**Passed:** %d\n", results.Passed))
	sb.WriteString(fmt.Sprintf("**Failed:** %d\n", results.Failed))
	sb.WriteString(fmt.Sprintf("**Skipped:** %d\n", results.Skipped))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", results.Duration))

	if results.Failed > 0 {
		sb.WriteString("## Failed Tests\n\n")
		for _, test := range results.Tests {
			if test.Status == "failed" {
				sb.WriteString(fmt.Sprintf("### %s.%s\n\n", test.Service, test.Operation))
				sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", test.Error))
				if test.Output != "" {
					sb.WriteString("<details>\n<summary>Output</summary>\n\n```\n")
					sb.WriteString(test.Output)
					sb.WriteString("\n```\n</details>\n\n")
				}
			}
		}
	}

	sb.WriteString("## Results by Service\n\n")
	sb.WriteString("| Service | Passed |\n")
	sb.WriteString("|---------|--------|\n")
	for service, count := range results.ByService {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", service, count))
	}

	return sb.String()
}

// TerraformTestOutput represents the JSON output from terraform test.
type TerraformTestOutput struct {
	Type      string                 `json:"@type"`
	Timestamp string                 `json:"@timestamp"`
	Test      *TerraformTestResult   `json:"test,omitempty"`
	Diagnostic *TerraformDiagnostic  `json:"diagnostic,omitempty"`
}

// TerraformTestResult represents a test result from terraform test -json.
type TerraformTestResult struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Run    string `json:"run,omitempty"`
}

// TerraformDiagnostic represents a diagnostic message.
type TerraformDiagnostic struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail,omitempty"`
}

// ParseTerraformJSON parses JSON output from terraform test -json.
func ParseTerraformJSON(output string) ([]TerraformTestOutput, error) {
	var results []TerraformTestOutput

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var result TerraformTestOutput
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue // Skip non-JSON lines
		}
		results = append(results, result)
	}

	return results, nil
}
