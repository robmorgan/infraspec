package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/aigen"
)

var checkTarget string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Analyze code for pattern compliance",
	Long: `Analyze generated or existing code for compliance with project patterns,
including banned patterns (manual XML construction), recommended patterns
(response builders), and Go syntax/vet issues.

Examples:
  cloudmirror check --target=generated/services/sns/
  cloudmirror check --target=internal/emulator/services/rds/service.go`,
	Run: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVar(&checkTarget, "target", "", "Path to code to analyze [required]")

	checkCmd.MarkFlagRequired("target")
}

func runCheck(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Create validator
	validator := aigen.NewCodeValidator(aigen.ValidatorConfig{
		Verbose: verbose,
	})

	if !quiet {
		fmt.Fprintf(os.Stderr, "Analyzing code in %s...\n", checkTarget)
	}

	// Check if target is a file or directory
	info, err := os.Stat(checkTarget)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing target: %v\n", err)
		os.Exit(1)
	}

	var results []aigen.ValidationResult

	if info.IsDir() {
		// Walk directory and validate all .go files
		err = filepath.Walk(checkTarget, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}

			result, err := validator.ValidateFile(ctx, path)
			if err != nil {
				return fmt.Errorf("error validating %s: %w", path, err)
			}
			results = append(results, *result)
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Single file
		result, err := validator.ValidateFile(ctx, checkTarget)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
			os.Exit(1)
		}
		results = append(results, *result)
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
		valid := 0
		invalid := 0
		totalErrors := 0
		totalWarnings := 0

		for _, r := range results {
			if r.Valid {
				valid++
			} else {
				invalid++
			}
			totalErrors += len(r.CompileErrors) + len(r.VetErrors) + len(r.PatternErrors)
			totalWarnings += len(r.Warnings)
		}

		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "=== Code Analysis Summary ===\n")
		fmt.Fprintf(os.Stderr, "Files analyzed: %d\n", len(results))
		fmt.Fprintf(os.Stderr, "  Valid: %d\n", valid)
		fmt.Fprintf(os.Stderr, "  Invalid: %d\n", invalid)
		fmt.Fprintf(os.Stderr, "Errors: %d\n", totalErrors)
		fmt.Fprintf(os.Stderr, "Warnings: %d\n", totalWarnings)

		if invalid > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Files with issues:")
			for _, r := range results {
				if !r.Valid {
					fmt.Fprintf(os.Stderr, "  %s:\n", r.FilePath)
					for _, e := range r.CompileErrors {
						fmt.Fprintf(os.Stderr, "    [compile] %s\n", e)
					}
					for _, e := range r.VetErrors {
						fmt.Fprintf(os.Stderr, "    [vet] %s\n", e)
					}
					for _, e := range r.PatternErrors {
						fmt.Fprintf(os.Stderr, "    [pattern] %s\n", e)
					}
				}
			}
		}
	}

	// Exit with error if any files are invalid
	for _, r := range results {
		if !r.Valid {
			os.Exit(1)
		}
	}
}
