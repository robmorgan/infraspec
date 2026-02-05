package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/analyzer"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/generator"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/reporter"
)

var (
	analyzeService       string
	analyzeOutput        string
	analyzeOutputFile    string
	analyzeBaseline      string
	analyzePriority      string
	analyzeGenerateStubs bool
	analyzeStubsFile     string
	analyzeWebsiteReport bool
	analyzeInfraspecPath string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze AWS API coverage for implemented services",
	Long: `Analyze AWS API coverage by comparing implemented services against
the AWS SDK Go V2 models. Generates coverage reports in various formats.

Examples:
  cloudmirror analyze --service=rds --output=json
  cloudmirror analyze --service=all --output=markdown --output-file=PARITY.md
  cloudmirror analyze --service=rds --baseline=baseline.json
  cloudmirror analyze --service=rds --generate-stubs`,
	Run: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeService, "service", "", "Service to analyze (e.g., 'rds', 's3', or 'all') [required]")
	analyzeCmd.Flags().StringVar(&analyzeOutput, "output", "json", "Output format: json, markdown, badge")
	analyzeCmd.Flags().StringVar(&analyzeOutputFile, "output-file", "", "Output file path (default: stdout)")
	analyzeCmd.Flags().StringVar(&analyzeBaseline, "baseline", "", "Baseline report JSON for diff comparison")
	analyzeCmd.Flags().StringVar(&analyzePriority, "priority", "", "Filter by priority: high, medium, low")
	analyzeCmd.Flags().BoolVar(&analyzeGenerateStubs, "generate-stubs", false, "Generate stub implementations for missing operations")
	analyzeCmd.Flags().StringVar(&analyzeStubsFile, "stubs-file", "", "Output file for generated stubs")
	analyzeCmd.Flags().BoolVar(&analyzeWebsiteReport, "website-report", false, "Generate combined report for InfraSpec website")
	analyzeCmd.Flags().StringVar(&analyzeInfraspecPath, "infraspec-path", "", "Path to InfraSpec repository (for scanning assertions)")

	analyzeCmd.MarkFlagRequired("service")
}

func runAnalyze(cmd *cobra.Command, args []string) {
	requireSDKPath()

	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)

	// Handle website report generation
	if analyzeWebsiteReport {
		generateWebsiteReport(anal, analyzeInfraspecPath, analyzeOutputFile)
		return
	}

	// Run analysis
	var reports []*models.CoverageReport
	var err error

	if analyzeService == "all" {
		reports, err = anal.AnalyzeAllServices()
	} else {
		var report *models.CoverageReport
		report, err = anal.AnalyzeService(analyzeService)
		if report != nil {
			reports = append(reports, report)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing service(s): %v\n", err)
		os.Exit(1)
	}

	if len(reports) == 0 {
		fmt.Fprintln(os.Stderr, "No reports generated")
		os.Exit(1)
	}

	// Handle baseline comparison
	if analyzeBaseline != "" {
		if len(reports) != 1 {
			fmt.Fprintln(os.Stderr, "Error: baseline comparison requires a single service (use --service=<service-name>, not --service=all)")
			os.Exit(1)
		}

		baseline, err := loadBaseline(analyzeBaseline)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading baseline: %v\n", err)
			os.Exit(1)
		}

		diffEngine := analyzer.NewDiffEngine()
		diff := diffEngine.CompareReports(baseline, reports[0])

		rep := reporter.NewReporter()
		diffOutput, err := rep.GenerateDiffReport(diff, reporter.OutputFormat(analyzeOutput))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating diff report: %v\n", err)
			os.Exit(1)
		}

		writeOutput(diffOutput, analyzeOutputFile)
		return
	}

	// Generate report
	rep := reporter.NewReporter()
	var outputData []byte

	if len(reports) == 1 {
		outputData, err = rep.GenerateReport(reports[0], reporter.OutputFormat(analyzeOutput))
	} else {
		outputData, err = rep.GenerateMultiReport(reports, reporter.OutputFormat(analyzeOutput))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(1)
	}

	writeOutput(outputData, analyzeOutputFile)

	// Generate stubs if requested
	if analyzeGenerateStubs && len(reports) == 1 {
		stubGen, err := generator.NewStubGenerator()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing stub generator: %v\n", err)
			os.Exit(1)
		}

		var stubs string
		if analyzePriority != "" {
			stubs, err = stubGen.GenerateStubsForPriority(reports[0], models.Priority(analyzePriority))
		} else {
			stubs, err = stubGen.GenerateStubs(reports[0])
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating stubs: %v\n", err)
			os.Exit(1)
		}

		stubOutput := analyzeStubsFile
		if stubOutput == "" {
			stubOutput = filepath.Join(servicesPath, analyzeService, "stubs_generated.go")
		}

		if err := os.WriteFile(stubOutput, []byte(stubs), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing stubs: %v\n", err)
			os.Exit(1)
		}

		if !quiet {
			fmt.Fprintf(os.Stderr, "Generated stubs: %s\n", stubOutput)
		}
	}

	// Print summary to stderr
	if !quiet && analyzeOutputFile != "" {
		for _, r := range reports {
			fmt.Fprintf(os.Stderr, "Service: %s - Coverage: %.1f%% (%d/%d operations)\n",
				r.ServiceName, r.CoveragePercent, len(r.Supported), r.TotalOperations-len(r.Deprecated))
		}
	}
}

func loadBaseline(path string) (*models.CoverageReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var report models.CoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

func writeOutput(data []byte, outputFile string) {
	if outputFile != "" {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(outputFile, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(string(data))
	}
}

func generateWebsiteReport(anal *analyzer.Analyzer, infraspecPath, outputFile string) {
	vcReports, err := anal.AnalyzeAllServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing Virtual Cloud services: %v\n", err)
		os.Exit(1)
	}

	if infraspecPath == "" {
		infraspecPath = analyzer.GetAssertionsPath()
	}
	scanner := analyzer.NewInfraSpecScanner(infraspecPath)
	infraspecOps, err := scanner.ScanAssertions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not scan InfraSpec assertions: %v\n", err)
		infraspecOps = make(map[string][]models.InfraSpecOperation)
	}

	websiteReporter := reporter.NewWebsiteReporter()
	report, err := websiteReporter.GenerateWebsiteReport(vcReports, infraspecOps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating website report: %v\n", err)
		os.Exit(1)
	}

	outputData, err := websiteReporter.GenerateJSON(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	writeOutput(outputData, outputFile)

	if !quiet {
		fmt.Fprintf(os.Stderr, "Generated website report with %d services\n", len(report.Services))
	}
}
