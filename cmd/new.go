package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [feature-file-name]",
	Short: "Create a new feature file",
	Long: `Create a new feature file in the ./features directory.

The feature file will be created with a basic template structure.
If the name doesn't end with .feature, the extension will be added automatically.
All feature files are created under the ./features directory.

Examples:
  infraspec new my-test.feature
  infraspec new api-test.feature
  infraspec new database-test`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	RootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fileName := args[0]

	// Ensure the file has .feature extension
	if !strings.HasSuffix(fileName, ".feature") {
		fileName += ".feature"
	}

	// Always create under ./features directory
	filePath := filepath.Join("features", fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	// Create features directory if it doesn't exist
	if err := os.MkdirAll("features", 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create features directory: %w", err)
	}

	// Create the feature file with template content
	template := `Feature: ` + getFeatureName(fileName) + `
  As a DevOps engineer
  I want to create cloud infrastructure
  So that I can deliver software reliability

  Scenario: Basic infrastructure check
    Given the system is running
    When I check the service status
    Then the service should be healthy
`

	if err := os.WriteFile(filePath, []byte(template), 0o600); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create feature file: %w", err)
	}

	cmd.Printf("✅ Successfully created feature file: %s\n", filePath)
	cmd.Printf("📝 Edit the file to add your specific test scenarios\n")
	cmd.Printf("🚀 Run your test with: infraspec %s\n", filePath)

	return nil
}

func getFeatureName(filePath string) string {
	// Extract filename without extension and convert to title case
	filename := filepath.Base(filePath)
	name := strings.TrimSuffix(filename, ".feature")

	// Replace hyphens and underscores with spaces and title case
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Simple title case conversion
	words := strings.Fields(name)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
