package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/aigen"
)

var (
	generateChangeReport   string
	generateOutputDir      string
	generateClaudeAPIKey   string
	generateClaudeModel    string
	generateDryRun         bool
	generateMaxOperations  int
	generatePreparePrompts bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "AI-powered implementation generation",
	Long: `Generate AWS service handler implementations using AI based on a
change report from SDK comparison or analysis.

Modes:
  Default:          Call Claude API directly to generate code
  --prepare-prompts: Write prompt files for use with Claude Code Action

Examples:
  cloudmirror generate --change-report=changes.json --sdk-path=/tmp/aws-sdk-go-v2
  cloudmirror generate --change-report=changes.json --dry-run
  cloudmirror generate --change-report=changes.json --prepare-prompts --output-dir=prompts`,
	Run: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateChangeReport, "change-report", "", "Path to SDK change report JSON [required]")
	generateCmd.Flags().StringVar(&generateOutputDir, "output-dir", "generated", "Output directory for generated code/prompts")
	generateCmd.Flags().StringVar(&generateClaudeAPIKey, "claude-api-key", "", "Anthropic API key (or ANTHROPIC_API_KEY env)")
	generateCmd.Flags().StringVar(&generateClaudeModel, "claude-model", "claude-sonnet-4-5", "Claude model to use")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview generation without writing files")
	generateCmd.Flags().IntVar(&generateMaxOperations, "max-operations", 0, "Maximum operations to generate (0 = unlimited)")
	generateCmd.Flags().BoolVar(&generatePreparePrompts, "prepare-prompts", false, "Write prompt files instead of calling Claude API")

	generateCmd.MarkFlagRequired("change-report")
}

func runGenerate(cmd *cobra.Command, args []string) {
	requireSDKPath()

	ctx := context.Background()

	// Load change report
	data, err := os.ReadFile(generateChangeReport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading change report: %v\n", err)
		os.Exit(1)
	}

	var report aigen.SDKChangeReport
	if err := json.Unmarshal(data, &report); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing change report: %v\n", err)
		os.Exit(1)
	}

	// Handle --prepare-prompts mode
	if generatePreparePrompts {
		runPreparePrompts(ctx, &report)
		return
	}

	// API mode: Get API key from environment if not provided
	apiKey := generateClaudeAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY is required (set via --claude-api-key or environment variable)")
		fmt.Fprintln(os.Stderr, "Tip: Use --prepare-prompts to generate prompt files for Claude Code Action instead")
		os.Exit(1)
	}

	// Create generator config
	config := aigen.Config{
		SDKPath:       sdkPath,
		ServicesPath:  servicesPath,
		OutputDir:     generateOutputDir,
		ClaudeAPIKey:  apiKey,
		ClaudeModel:   generateClaudeModel,
		DryRun:        generateDryRun,
		Verbose:       verbose,
		MaxOperations: generateMaxOperations,
	}

	// Create generator
	generator := aigen.NewImplementationGenerator(config)

	// Count services and operations to process
	var servicesToProcess, totalOperations int
	for _, svc := range report.Services {
		if len(svc.NewOperations) > 0 {
			servicesToProcess++
			totalOperations += len(svc.NewOperations)
		}
	}

	// Generate implementations
	if !quiet {
		fmt.Fprintf(os.Stderr, "Generating implementations from %s...\n", generateChangeReport)
		fmt.Fprintf(os.Stderr, "Processing %d services with %d operations to implement\n", servicesToProcess, totalOperations)
		if generateMaxOperations > 0 {
			fmt.Fprintf(os.Stderr, "Limited to %d operations per run\n", generateMaxOperations)
		}
		if generateDryRun {
			fmt.Fprintln(os.Stderr, "(dry-run mode - no files will be written)")
		}
		fmt.Fprintln(os.Stderr)
	}

	result, err := generator.Generate(ctx, &report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating implementations: %v\n", err)
		os.Exit(1)
	}

	// Output result
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	// Print summary
	if !quiet {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "=== Generation Summary ===\n")
		fmt.Fprintf(os.Stderr, "Services processed: %d\n", result.ServicesProcessed)
		fmt.Fprintf(os.Stderr, "Operations created: %d\n", result.OperationsCreated)
		fmt.Fprintf(os.Stderr, "Tests created: %d\n", result.TestsCreated)
		if result.LimitReached {
			fmt.Fprintf(os.Stderr, "Note: Operation limit reached (%d)\n", generateMaxOperations)
		}
		if len(result.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "Errors: %d\n", len(result.Errors))
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  - %s.%s (%s): %s\n", e.Service, e.Operation, e.Phase, e.Message)
			}
		}
	}
}

// runPreparePrompts generates prompt files for use with Claude Code Action
func runPreparePrompts(ctx context.Context, report *aigen.SDKChangeReport) {
	// Create prompt generator
	promptGen := aigen.NewPromptGenerator(aigen.PromptGeneratorConfig{
		SDKPath:       sdkPath,
		ServicesPath:  servicesPath,
		OutputDir:     generateOutputDir,
		MaxOperations: generateMaxOperations,
		Verbose:       verbose,
	})

	if !quiet {
		fmt.Fprintf(os.Stderr, "Preparing prompts from %s...\n", generateChangeReport)
		if generateMaxOperations > 0 {
			fmt.Fprintf(os.Stderr, "Limited to %d operations\n", generateMaxOperations)
		}
		fmt.Fprintln(os.Stderr)
	}

	result, err := promptGen.PreparePrompts(ctx, report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing prompts: %v\n", err)
		os.Exit(1)
	}

	// Output result as JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	// Print summary
	if !quiet {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "=== Prompt Preparation Summary ===\n")
		fmt.Fprintf(os.Stderr, "Prompts created: %d\n", len(result.Prompts))
		fmt.Fprintf(os.Stderr, "Output directory: %s\n", generateOutputDir)
		fmt.Fprintf(os.Stderr, "Manifest file: %s/manifest.json\n", generateOutputDir)
		if result.LimitReached {
			fmt.Fprintf(os.Stderr, "Note: Operation limit reached (%d)\n", generateMaxOperations)
		}
	}
}
