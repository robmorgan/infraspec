package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/analyzer"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/generator"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/typegen"
)

var (
	scaffoldService  string
	scaffoldOutput   string
	scaffoldDryRun   bool
	scaffoldPriority string
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate service scaffold for a new AWS service",
	Long: `Generate a complete scaffold for implementing a new AWS service,
including service.go with handler stubs and types.go with type definitions.

Examples:
  cloudmirror scaffold --service=sns
  cloudmirror scaffold --service=sns --priority=high
  cloudmirror scaffold --service=sns --dry-run
  cloudmirror scaffold --service=sns --output=./my-output-dir`,
	Run: runScaffold,
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)

	scaffoldCmd.Flags().StringVar(&scaffoldService, "service", "", "Service to scaffold [required]")
	scaffoldCmd.Flags().StringVar(&scaffoldOutput, "output", "", "Output directory (default: internal/emulator/services/<service>)")
	scaffoldCmd.Flags().BoolVar(&scaffoldDryRun, "dry-run", false, "Preview scaffold without writing files")
	scaffoldCmd.Flags().StringVar(&scaffoldPriority, "priority", "", "Filter operations by priority: high, medium, low")

	scaffoldCmd.MarkFlagRequired("service")
}

func runScaffold(cmd *cobra.Command, args []string) {
	requireSDKPath()

	if scaffoldService == "all" {
		fmt.Fprintln(os.Stderr, "Error: scaffold requires a specific service name (not 'all')")
		os.Exit(1)
	}

	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)

	// Get AWS model for the service
	awsModel, err := anal.GetAWSModel(scaffoldService)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting AWS model for %s: %v\n", scaffoldService, err)
		os.Exit(1)
	}

	// Create stub generator
	stubGen, err := generator.NewStubGenerator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing scaffold generator: %v\n", err)
		os.Exit(1)
	}

	// Generate scaffold
	var priority models.Priority
	if scaffoldPriority != "" {
		priority = models.Priority(scaffoldPriority)
	}

	scaffold, err := stubGen.GenerateScaffold(awsModel, priority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating scaffold: %v\n", err)
		os.Exit(1)
	}

	// Determine output directory
	outputDir := scaffoldOutput
	if outputDir == "" {
		outputDir = filepath.Join(servicesPath, strings.ToLower(scaffoldService))
	}

	// Generate Smithy types using gentypes
	smithyTypesCode, smithyTypesErr := generateSmithyTypes(scaffoldService, strings.ToLower(scaffoldService))

	if scaffoldDryRun {
		// Dry run: print to stdout
		fmt.Println("=== service.go ===")
		fmt.Println(scaffold.ServiceCode)
		if smithyTypesErr == nil && smithyTypesCode != "" {
			fmt.Println("\n=== smithy_types.go ===")
			fmt.Println(smithyTypesCode)
		} else if smithyTypesErr != nil && !quiet {
			fmt.Fprintf(os.Stderr, "\nNote: Could not generate smithy_types.go: %v\n", smithyTypesErr)
			fmt.Fprintf(os.Stderr, "You can generate it later with: cloudmirror gentypes --service=%s\n", scaffoldService)
		}

		if !quiet {
			fmt.Fprintf(os.Stderr, "\nDry run complete. Would write to: %s\n", outputDir)
		}
		return
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write service.go
	servicePath := filepath.Join(outputDir, "service.go")
	if err := os.WriteFile(servicePath, []byte(scaffold.ServiceCode), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing service.go: %v\n", err)
		os.Exit(1)
	}

	// Write smithy_types.go if generated successfully
	var smithyTypesPath string
	if smithyTypesErr == nil && smithyTypesCode != "" {
		smithyTypesPath = filepath.Join(outputDir, "smithy_types.go")
		if err := os.WriteFile(smithyTypesPath, []byte(smithyTypesCode), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error writing smithy_types.go: %v\n", err)
			smithyTypesPath = ""
		}
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Generated scaffold for %s:\n", scaffoldService)
		fmt.Fprintf(os.Stderr, "  - %s\n", servicePath)
		if smithyTypesPath != "" {
			fmt.Fprintf(os.Stderr, "  - %s (types from Smithy models)\n", smithyTypesPath)
		} else if smithyTypesErr != nil {
			fmt.Fprintf(os.Stderr, "\nNote: Could not generate smithy_types.go: %v\n", smithyTypesErr)
			fmt.Fprintf(os.Stderr, "You can generate it later with: cloudmirror gentypes --service=%s\n", scaffoldService)
		}
		fmt.Fprintf(os.Stderr, "\nNext steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review and implement the handler stubs in service.go\n")
		fmt.Fprintf(os.Stderr, "  2. Use types from smithy_types.go for response data\n")
		fmt.Fprintf(os.Stderr, "  3. Register the service in cmd/emulator/main.go\n")
	}
}

// generateSmithyTypes generates entity types from Smithy models using gentypes
func generateSmithyTypes(serviceName, packageName string) (string, error) {
	// Find models path (auto-downloads if needed)
	modelsPath := findModelsPath()
	if modelsPath == "" {
		return "", fmt.Errorf("AWS API Models not available")
	}

	// Find model file for the service
	modelPath, err := findModelFile(modelsPath, serviceName)
	if err != nil {
		return "", fmt.Errorf("model not found: %w", err)
	}

	// Create generator config
	config := &typegen.Config{
		ServiceName:  serviceName,
		PackageName:  packageName,
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	// Generate types
	gen := typegen.NewGenerator(config)
	return gen.Generate()
}
