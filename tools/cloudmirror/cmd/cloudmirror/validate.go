package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/terraform"
)

var (
	validateEmulatorEndpoint string
	validateTestsPath        string
	validateFilter           string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run Terraform validation tests",
	Long: `Run Terraform validation tests against the emulator to verify
that generated implementations work correctly.

Examples:
  cloudmirror validate
  cloudmirror validate --emulator-endpoint=http://localhost:3687
  cloudmirror validate --filter=rds
  cloudmirror validate --tests-path=terraform/tests`,
	Run: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateEmulatorEndpoint, "emulator-endpoint", "http://localhost:3687", "Emulator endpoint URL")
	validateCmd.Flags().StringVar(&validateTestsPath, "tests-path", "terraform/tests", "Path to Terraform tests")
	validateCmd.Flags().StringVar(&validateFilter, "filter", "", "Filter tests by service or operation")
}

func runValidate(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Create runner
	runner := terraform.NewTestRunner(terraform.Config{
		Endpoint:  validateEmulatorEndpoint,
		TestsPath: validateTestsPath,
		Filter:    validateFilter,
		Verbose:   verbose,
	})

	if !quiet {
		fmt.Fprintf(os.Stderr, "Running Terraform validation tests...\n")
		fmt.Fprintf(os.Stderr, "  Emulator: %s\n", validateEmulatorEndpoint)
		fmt.Fprintf(os.Stderr, "  Tests path: %s\n", validateTestsPath)
		if validateFilter != "" {
			fmt.Fprintf(os.Stderr, "  Filter: %s\n", validateFilter)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Run tests
	results, err := runner.RunAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running tests: %v\n", err)
		os.Exit(1)
	}

	// Output results
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling results: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	// Print summary
	if !quiet {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "=== Validation Summary ===\n")
		fmt.Fprintf(os.Stderr, "Total: %d tests\n", results.TotalTests)
		fmt.Fprintf(os.Stderr, "  Passed: %d\n", results.Passed)
		fmt.Fprintf(os.Stderr, "  Failed: %d\n", results.Failed)
		fmt.Fprintf(os.Stderr, "  Skipped: %d\n", results.Skipped)
	}

	// Exit with error if any tests failed
	if results.Failed > 0 {
		os.Exit(1)
	}
}
