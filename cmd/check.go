package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/check"
)

var (
	checkPlanPath string
	checkDir      string
	checkSeverity string
	checkIgnore   []string
	checkFormat   string
)

// checkCmd represents the check command.
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Evaluate Terraform plans against security rules",
	Long: `Evaluate a Terraform plan against security and best practice rules.

The check command runs your Terraform plan through InfraSpec's rule engine
to identify security issues, misconfigurations, and policy violations before
applying changes to your infrastructure.

Examples:
  # Check a saved plan JSON file
  infraspec check --plan tfplan.json

  # Check the current Terraform directory
  infraspec check --dir ./terraform

  # Only show critical issues
  infraspec check --plan tfplan.json --severity critical

  # Ignore specific rules
  infraspec check --plan tfplan.json --ignore aws-sg-no-public-ssh

  # Output as JSON
  infraspec check --plan tfplan.json --format json`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&checkPlanPath, "plan", "", "path to Terraform plan JSON file")
	checkCmd.Flags().StringVar(&checkDir, "dir", ".", "path to Terraform directory (used when --plan is not specified)")
	checkCmd.Flags().StringVar(&checkSeverity, "severity", "info", "minimum severity level: critical, warning, info")
	checkCmd.Flags().StringSliceVar(&checkIgnore, "ignore", nil, "rule IDs to skip (can be specified multiple times)")
	checkCmd.Flags().StringVar(&checkFormat, "format", "text", "output format: text, json")

	RootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Parse severity
	minSeverity, err := check.SeverityFromString(checkSeverity)
	if err != nil {
		return fmt.Errorf("invalid severity: %w", err)
	}

	// Build options
	opts := check.Options{
		PlanPath:      checkPlanPath,
		Dir:           checkDir,
		MinSeverity:   minSeverity,
		IgnoreRuleIDs: checkIgnore,
		Format:        checkFormat,
	}

	// Create and run the check runner
	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	// Format and output results
	formatter := check.NewFormatter(opts.Format)
	if err := formatter.Format(os.Stdout, summary); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Exit with appropriate code
	if summary.ExitCode != 0 {
		os.Exit(summary.ExitCode)
	}

	return nil
}
