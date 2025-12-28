package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/modelscache"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/typegen"
	"github.com/spf13/cobra"
)

var (
	gentypesService       string
	gentypesOutput        string
	gentypesModelsPath    string
	gentypesProtocol      string
	gentypesDryRun        bool
	gentypesOperations    string
	gentypesTypeSuffix    string
	gentypesIncludeInputs bool
)

// modelsCacheInstance is the global models cache instance
var modelsCacheInstance *modelscache.ModelsCache

var gentypesCmd = &cobra.Command{
	Use:   "gentypes",
	Short: "Generate Go types from AWS Smithy models with correct XML tags",
	Long: `Generate Go types from AWS Smithy models (api-models-aws repository) with
correct XML serialization tags based on the smithy.api#xmlName trait.

This tool parses the Smithy JSON AST from github.com/aws/api-models-aws and
generates Go structs with proper xml:"..." tags for use in AWS API emulation.

The generated types use camelCase XML element names (e.g., instanceId, vpcId)
matching real AWS API responses, rather than PascalCase from Go SDK types.

By default, only response (output) types are generated. Use --include-inputs
to also generate request (input) types for type-safe request parsing.

Examples:
  cloudmirror gentypes --service=ec2
  cloudmirror gentypes --service=rds --output=./internal/emulator/services/rds/smithy_types.go
  cloudmirror gentypes --service=iam --dry-run
  cloudmirror gentypes --service=ec2 --operations=DescribeInstances,DescribeVpcs
  cloudmirror gentypes --service=iam --include-inputs`,
	Run: runGentypes,
}

func init() {
	rootCmd.AddCommand(gentypesCmd)

	gentypesCmd.Flags().StringVar(&gentypesService, "service", "", "AWS service name to generate types for [required]")
	gentypesCmd.Flags().StringVar(&gentypesOutput, "output", "", "Output file path (default: ./internal/emulator/services/<service>/smithy_types.go)")
	gentypesCmd.Flags().StringVar(&gentypesModelsPath, "models-path", "", "Path to api-models-aws repo (auto-downloaded if not specified)")
	gentypesCmd.Flags().StringVar(&gentypesProtocol, "protocol", "", "Override protocol detection (ec2, query, rest-xml, json)")
	gentypesCmd.Flags().BoolVar(&gentypesDryRun, "dry-run", false, "Print generated code to stdout without writing")
	gentypesCmd.Flags().StringVar(&gentypesOperations, "operations", "", "Comma-separated list of operations to generate types for (default: all)")
	gentypesCmd.Flags().StringVar(&gentypesTypeSuffix, "suffix", "", "Suffix to add to generated type names (e.g., 'XML' -> VpcXML)")
	gentypesCmd.Flags().BoolVar(&gentypesIncludeInputs, "include-inputs", false, "Also generate input types for request parsing")

	gentypesCmd.MarkFlagRequired("service")
}

func runGentypes(cmd *cobra.Command, args []string) {
	// Get models path
	modelsPath := gentypesModelsPath
	if modelsPath == "" {
		modelsPath = findModelsPath()
	}

	if modelsPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not find or download AWS API Models.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Please clone the models and specify --models-path:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  git clone --depth 1 https://github.com/aws/api-models-aws.git ~/.cloudmirror/api-models-aws")
		fmt.Fprintln(os.Stderr, "  cloudmirror gentypes --models-path=~/.cloudmirror/api-models-aws --service=ec2")
		os.Exit(1)
	}

	// Find the model file for the service
	modelPath, err := findModelFile(modelsPath, gentypesService)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Using model: %s\n", modelPath)
	}

	// Determine output path
	// Convert service name to valid Go package name (remove hyphens)
	goPackageName := strings.ReplaceAll(strings.ToLower(gentypesService), "-", "")
	outputPath := gentypesOutput
	if outputPath == "" {
		outputPath = filepath.Join(servicesPath, goPackageName, "smithy_types.go")
	}

	// Parse operations
	var operations []string
	if gentypesOperations != "" {
		operations = strings.Split(gentypesOperations, ",")
		for i := range operations {
			operations[i] = strings.TrimSpace(operations[i])
		}
	}

	// Create generator config
	config := &typegen.Config{
		ServiceName:   gentypesService,
		PackageName:   goPackageName,
		Protocol:      gentypesProtocol,
		OutputPath:    outputPath,
		ModelPath:     modelPath,
		ResponseOnly:  !gentypesIncludeInputs, // If including inputs, don't limit to response-only
		IncludeInputs: gentypesIncludeInputs,
		Operations:    operations,
		TypeSuffix:    gentypesTypeSuffix,
	}

	// Create and run generator
	generator := typegen.NewGenerator(config)

	if gentypesDryRun {
		code, err := generator.Generate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating types: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(code)
		if !quiet {
			fmt.Fprintf(os.Stderr, "\nDry run complete. Would write to: %s\n", outputPath)
		}
		return
	}

	// Generate and write file
	if err := generator.GenerateToFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Generated types written to: %s\n", outputPath)
	}
}

// findModelsPath attempts to locate the AWS API Models directory
func findModelsPath() string {
	// Check explicit path
	if gentypesModelsPath != "" {
		if hasModelsSubdir(gentypesModelsPath) {
			return gentypesModelsPath
		}
	}

	// Check common locations
	candidates := []string{
		os.Getenv("AWS_API_MODELS_PATH"),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if hasModelsSubdir(candidate) {
			return candidate
		}
	}

	// Check if models are already cached in ~/.cloudmirror
	cache, err := modelscache.NewModelsCache(verbose, quiet)
	if err == nil {
		modelsCacheInstance = cache
		if cache.HasCache() {
			return cache.GetModelsDir()
		}
	}

	// Auto-download
	if cache == nil {
		cache, err = modelscache.NewModelsCache(verbose, quiet)
		if err != nil {
			return ""
		}
		modelsCacheInstance = cache
	}

	// Download models
	modelsDir, err := cache.GetModelsPath()
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Warning: Failed to download API Models: %v\n", err)
		}
		return ""
	}
	return modelsDir
}

// hasModelsSubdir checks if the path contains the models subdirectory
func hasModelsSubdir(path string) bool {
	modelsPath := filepath.Join(path, "models")
	info, err := os.Stat(modelsPath)
	return err == nil && info.IsDir()
}

// findModelFile finds the model file for a given service
func findModelFile(modelsPath, serviceName string) (string, error) {
	serviceName = strings.ToLower(serviceName)

	// Map common service name variations
	serviceNameMap := map[string]string{
		"applicationautoscaling": "application-auto-scaling",
		"autoscaling":            "auto-scaling",
	}

	if mapped, ok := serviceNameMap[serviceName]; ok {
		serviceName = mapped
	}

	// Look for the service directory
	serviceDir := filepath.Join(modelsPath, "models", serviceName, "service")
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		// Try with hyphens
		serviceName = strings.ReplaceAll(serviceName, "_", "-")
		serviceDir = filepath.Join(modelsPath, "models", serviceName, "service")
		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			return "", fmt.Errorf("service %s not found in models at %s", serviceName, modelsPath)
		}
	}

	// Find the latest version directory
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return "", fmt.Errorf("failed to read service directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for service %s", serviceName)
	}

	// Use the latest version (sorted alphabetically, API versions are date-based)
	latestVersion := versions[len(versions)-1]
	for _, v := range versions {
		if v > latestVersion {
			latestVersion = v
		}
	}

	// Find the JSON model file
	versionDir := filepath.Join(serviceDir, latestVersion)
	jsonFiles, err := filepath.Glob(filepath.Join(versionDir, "*.json"))
	if err != nil || len(jsonFiles) == 0 {
		return "", fmt.Errorf("no model file found for service %s version %s", serviceName, latestVersion)
	}

	return jsonFiles[0], nil
}

// getModelsCache returns the global models cache instance
func getModelsCache() (*modelscache.ModelsCache, error) {
	if modelsCacheInstance != nil {
		return modelsCacheInstance, nil
	}

	cache, err := modelscache.NewModelsCache(verbose, quiet)
	if err != nil {
		return nil, err
	}
	modelsCacheInstance = cache
	return cache, nil
}
