package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/analyzer"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

var (
	compareOldSDKPath     string
	compareService        string
	compareOutput         string
	compareOutputFile     string
	compareOnlyBreaking   bool
	compareCheckImplement bool
	compareUseAllowlist   bool
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two SDK versions for breaking changes",
	Long: `Compare two AWS SDK versions to identify new operations, modified operations,
deprecated operations, removed operations, and breaking changes.

Examples:
  cloudmirror compare --old-sdk-path=/path/to/old-sdk --sdk-path=/path/to/new-sdk
  cloudmirror compare --old-sdk-path=/path/old --sdk-path=/path/new --only-breaking
  cloudmirror compare --old-sdk-path=/path/old --sdk-path=/path/new --service=rds
  cloudmirror compare --old-sdk-path=/path/old --sdk-path=/path/new --check-implement`,
	Run: runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.Flags().StringVar(&compareOldSDKPath, "old-sdk-path", "", "Path to old SDK version [required]")
	compareCmd.Flags().StringVar(&compareService, "service", "", "Service to compare (or 'all' for all services)")
	compareCmd.Flags().StringVar(&compareOutput, "output", "json", "Output format: json, markdown")
	compareCmd.Flags().StringVar(&compareOutputFile, "output-file", "", "Output file path (default: stdout)")
	compareCmd.Flags().BoolVar(&compareOnlyBreaking, "only-breaking", false, "Only show breaking changes")
	compareCmd.Flags().BoolVar(&compareCheckImplement, "check-implement", false, "Cross-reference changes with our implementation")
	compareCmd.Flags().BoolVar(&compareUseAllowlist, "use-allowlist", true, "Only include services from the allowlist")

	compareCmd.MarkFlagRequired("old-sdk-path")
}

func runCompare(cmd *cobra.Command, args []string) {
	requireSDKPath()

	if compareOldSDKPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --old-sdk-path is required")
		os.Exit(1)
	}

	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)

	// Build service list if filtered
	var services []string
	if compareService != "" && compareService != "all" {
		services = []string{compareService}
	}

	config := &models.SDKCompareConfig{
		OldSDKPath:     compareOldSDKPath,
		NewSDKPath:     sdkPath,
		Services:       services,
		OnlyBreaking:   compareOnlyBreaking,
		CheckImplement: compareCheckImplement,
		UseAllowlist:   compareUseAllowlist,
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Comparing SDK versions:\n")
		fmt.Fprintf(os.Stderr, "  Old: %s\n", config.OldSDKPath)
		fmt.Fprintf(os.Stderr, "  New: %s\n", config.NewSDKPath)
		if len(config.Services) > 0 {
			fmt.Fprintf(os.Stderr, "  Services: %v\n", config.Services)
		} else {
			fmt.Fprintf(os.Stderr, "  Services: all\n")
		}
		fmt.Fprintln(os.Stderr)
	}

	// Run comparison
	report, err := anal.CompareSDKVersions(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error comparing SDK versions: %v\n", err)
		os.Exit(1)
	}

	// Get implementation impact if requested
	var impacts []models.ImplementationImpact
	if compareCheckImplement {
		impacts, err = anal.GetImplementationImpact(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not analyze implementation impact: %v\n", err)
		}
	}

	// Generate output based on format
	var outputData []byte
	switch compareOutput {
	case "json":
		if len(impacts) > 0 {
			combined := struct {
				*models.SDKChangeReport
				ImplementationImpacts []models.ImplementationImpact `json:"implementation_impacts,omitempty"`
			}{
				SDKChangeReport:       report,
				ImplementationImpacts: impacts,
			}
			outputData, err = json.MarshalIndent(combined, "", "  ")
		} else {
			outputData, err = json.MarshalIndent(report, "", "  ")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

	case "markdown":
		outputData = generateSDKCompareMarkdown(report, impacts)

	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported output format: %s (use 'json' or 'markdown')\n", compareOutput)
		os.Exit(1)
	}

	writeOutput(outputData, compareOutputFile)

	// Print summary to stderr
	if !quiet {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "=== SDK Change Summary ===\n")
		fmt.Fprintf(os.Stderr, "Services with changes: %d\n", report.Summary.TotalServicesChanged)
		fmt.Fprintf(os.Stderr, "New operations: %d\n", report.Summary.TotalNewOperations)
		fmt.Fprintf(os.Stderr, "Modified operations: %d\n", report.Summary.TotalModifiedOps)
		fmt.Fprintf(os.Stderr, "Deprecated operations: %d\n", report.Summary.TotalDeprecatedOps)
		fmt.Fprintf(os.Stderr, "Removed operations: %d\n", report.Summary.TotalRemovedOps)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Breaking changes: %d total\n", report.Summary.TotalBreakingChanges)
		if report.Summary.CriticalBreakingCount > 0 {
			fmt.Fprintf(os.Stderr, "  - Critical: %d\n", report.Summary.CriticalBreakingCount)
		}
		if report.Summary.HighBreakingCount > 0 {
			fmt.Fprintf(os.Stderr, "  - High: %d\n", report.Summary.HighBreakingCount)
		}
		if report.Summary.MediumBreakingCount > 0 {
			fmt.Fprintf(os.Stderr, "  - Medium: %d\n", report.Summary.MediumBreakingCount)
		}
		if report.Summary.LowBreakingCount > 0 {
			fmt.Fprintf(os.Stderr, "  - Low: %d\n", report.Summary.LowBreakingCount)
		}

		if len(impacts) > 0 {
			needUpdate := 0
			needNew := 0
			for _, impact := range impacts {
				if impact.RequiresUpdate {
					needUpdate++
				}
				if impact.RequiresNewHandler {
					needNew++
				}
			}
			fmt.Fprintln(os.Stderr)
			fmt.Fprintf(os.Stderr, "Implementation impact:\n")
			fmt.Fprintf(os.Stderr, "  - Operations needing updates: %d\n", needUpdate)
			fmt.Fprintf(os.Stderr, "  - New handlers needed: %d\n", needNew)
		}
	}
}

func generateSDKCompareMarkdown(report *models.SDKChangeReport, impacts []models.ImplementationImpact) []byte {
	var sb strings.Builder

	sb.WriteString("# AWS SDK Version Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Old Version:** %s\n", report.OldVersion))
	sb.WriteString(fmt.Sprintf("**New Version:** %s\n\n", report.NewVersion))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Services with changes | %d |\n", report.Summary.TotalServicesChanged))
	sb.WriteString(fmt.Sprintf("| New operations | %d |\n", report.Summary.TotalNewOperations))
	sb.WriteString(fmt.Sprintf("| Modified operations | %d |\n", report.Summary.TotalModifiedOps))
	sb.WriteString(fmt.Sprintf("| Deprecated operations | %d |\n", report.Summary.TotalDeprecatedOps))
	sb.WriteString(fmt.Sprintf("| Removed operations | %d |\n", report.Summary.TotalRemovedOps))
	sb.WriteString(fmt.Sprintf("| **Breaking changes** | **%d** |\n", report.Summary.TotalBreakingChanges))
	sb.WriteString("\n")

	// Breaking changes
	if len(report.BreakingChanges) > 0 {
		sb.WriteString("## Breaking Changes\n\n")

		bySeverity := make(map[models.Severity][]models.BreakingChange)
		for _, bc := range report.BreakingChanges {
			bySeverity[bc.Severity] = append(bySeverity[bc.Severity], bc)
		}

		severities := []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow}
		severityEmoji := map[models.Severity]string{
			models.SeverityCritical: "🔴",
			models.SeverityHigh:     "🟠",
			models.SeverityMedium:   "🟡",
			models.SeverityLow:      "🟢",
		}

		for _, sev := range severities {
			changes := bySeverity[sev]
			if len(changes) == 0 {
				continue
			}

			sb.WriteString(fmt.Sprintf("### %s %s (%d)\n\n", severityEmoji[sev], strings.Title(string(sev)), len(changes)))
			sb.WriteString("| Service | Operation | Reason | Remediation |\n")
			sb.WriteString("|---------|-----------|--------|-------------|\n")

			for _, bc := range changes {
				remediation := bc.Remediation
				if len(remediation) > 50 {
					remediation = remediation[:47] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					bc.Service, bc.Operation, bc.Reason, remediation))
			}
			sb.WriteString("\n")
		}
	}

	// Service changes
	if len(report.Services) > 0 {
		sb.WriteString("## Service Changes\n\n")

		for _, svc := range report.Services {
			sb.WriteString(fmt.Sprintf("### %s\n\n", svc.Name))

			if len(svc.NewOperations) > 0 {
				sb.WriteString("#### New Operations\n\n")
				for _, op := range svc.NewOperations {
					sb.WriteString(fmt.Sprintf("- **%s** (Priority: %s)\n", op.Name, op.Priority))
					if op.Documentation != "" {
						sb.WriteString(fmt.Sprintf("  - %s\n", op.Documentation))
					}
				}
				sb.WriteString("\n")
			}

			if len(svc.ModifiedOps) > 0 {
				sb.WriteString("#### Modified Operations\n\n")
				for _, op := range svc.ModifiedOps {
					breaking := ""
					if op.IsBreaking {
						breaking = " ⚠️"
					}
					sb.WriteString(fmt.Sprintf("- **%s**%s\n", op.Name, breaking))
					for _, change := range op.Changes {
						sb.WriteString(fmt.Sprintf("  - %s\n", change))
					}
				}
				sb.WriteString("\n")
			}

			if len(svc.DeprecatedOps) > 0 {
				sb.WriteString("#### Deprecated Operations\n\n")
				for _, op := range svc.DeprecatedOps {
					sb.WriteString(fmt.Sprintf("- %s\n", op))
				}
				sb.WriteString("\n")
			}

			if len(svc.RemovedOps) > 0 {
				sb.WriteString("#### Removed Operations\n\n")
				for _, op := range svc.RemovedOps {
					sb.WriteString(fmt.Sprintf("- %s\n", op))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Implementation impacts
	if len(impacts) > 0 {
		sb.WriteString("## Implementation Impact\n\n")
		sb.WriteString("| Service | Operation | Currently Implemented | Action Required | Suggested Action |\n")
		sb.WriteString("|---------|-----------|----------------------|-----------------|------------------|\n")

		for _, impact := range impacts {
			implemented := "No"
			if impact.CurrentlyImplemented {
				implemented = "Yes"
			}

			action := ""
			if impact.RequiresNewHandler {
				action = "New handler"
			} else if impact.RequiresUpdate {
				action = "Update handler"
			}

			suggested := impact.SuggestedAction
			if len(suggested) > 40 {
				suggested = suggested[:37] + "..."
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				impact.Service, impact.Operation, implemented, action, suggested))
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String())
}
