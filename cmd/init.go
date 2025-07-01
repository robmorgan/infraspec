package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize InfraSpec in the current directory",
	Long: `Initialize InfraSpec in the current directory by creating a ./features directory
if it doesn't already exist. This directory will contain your infrastructure test files.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	featuresDir := "./features"

	// Check if features directory already exists
	if _, err := os.Stat(featuresDir); err == nil {
		fmt.Printf("✅ Features directory already exists at %s\n", featuresDir)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking features directory: %w", err)
	}

	// Create the features directory
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		return fmt.Errorf("failed to create features directory: %w", err)
	}

	// Get absolute path for display
	absPath, err := filepath.Abs(featuresDir)
	if err != nil {
		absPath = featuresDir
	}

	fmt.Printf("🎉 Successfully initialized InfraSpec!\n")
	fmt.Printf("📁 Created features directory at: %s\n", absPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Create your first test: infraspec new my-test.feature\n")
	fmt.Printf("  2. Run your tests: infraspec features/my-test.feature\n")

	return nil
}
