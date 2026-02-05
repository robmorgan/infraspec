package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/pkg/gatekeeper"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/config"
)

var (
	// Check command flags
	checkRulesFile string
	checkFormat    string
	checkSeverity  string
	checkVarFile   string
	checkExclude   string
	checkInclude   string
	checkVerbose   bool
	checkNoBuiltin bool
	checkListRules bool
	checkStrict    bool
)

var checkCmd = &cobra.Command{
	Use:           "check [path...]",
	Short:         "Validate Terraform configurations against security rules",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Check Terraform configurations against security rules before applying.

InfraSpec Gatekeeper performs static analysis on Terraform HCL files to detect
security misconfigurations, policy violations, and best practice deviations.

Rule Discovery:
  InfraSpec automatically discovers rules from multiple sources:
  - Built-in rules (can be disabled with --no-builtin)
  - .infraspec.hcl in the repository root (auto-discovered)
  - *.spec.hcl files alongside Terraform configurations
  - Custom rules file specified with --rules

Examples:
  # Check all Terraform files in a directory
  infraspec check ./terraform

  # Check with custom rules
  infraspec check ./terraform --rules my-rules.hcl

  # Check with JSON output for CI
  infraspec check ./terraform --format json

  # List all available rules
  infraspec check --list-rules

  # Exclude specific rules
  infraspec check ./terraform --exclude S3_004,VPC_001

Exit codes:
  0 - All checks passed
  1 - One or more violations found
  2 - Parse error or configuration error`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Allow no args if --list-rules is specified
		if checkListRules {
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 path argument")
		}
		return nil
	},
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkRulesFile, "rules", "r", "", "path to custom rules HCL file")
	checkCmd.Flags().StringVarP(&checkFormat, "format", "f", "text", "output format (text, json)")
	checkCmd.Flags().StringVarP(&checkSeverity, "severity", "s", "error", "minimum severity to report (error, warning, info)")
	checkCmd.Flags().StringVar(&checkVarFile, "var-file", "", "path to tfvars file for variable resolution")
	checkCmd.Flags().StringVar(&checkExclude, "exclude", "", "comma-separated list of rule IDs to exclude")
	checkCmd.Flags().StringVar(&checkInclude, "include", "", "comma-separated list of rule IDs to include (excludes all others)")
	checkCmd.Flags().BoolVarP(&checkVerbose, "verbose", "v", false, "enable verbose output")
	checkCmd.Flags().BoolVar(&checkNoBuiltin, "no-builtin", false, "disable built-in rules")
	checkCmd.Flags().BoolVar(&checkListRules, "list-rules", false, "list all available rules and exit")
	checkCmd.Flags().BoolVar(&checkStrict, "strict", false, "treat unknown values as violations")

	RootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Try to find and load .infraspec.hcl config
	var infraspecConfig *config.LoadedConfig
	if len(args) > 0 {
		configPath, err := config.FindConfigFile(args[0])
		if err != nil && checkVerbose {
			fmt.Fprintf(os.Stderr, "Warning: error finding config file: %v\n", err)
		}
		if configPath != "" {
			infraspecConfig, err = config.LoadConfigFile(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config file %s: %v\n", configPath, err)
				return ExitError{Code: 2}
			}
			if checkVerbose {
				fmt.Fprintf(os.Stderr, "Loaded config from %s\n", configPath)
			}
		}
	}

	// Build checker configuration, merging config file settings with CLI flags
	// CLI flags take precedence over config file settings
	cfg := gatekeeper.Config{
		RulesFile:      checkRulesFile,
		VarFile:        checkVarFile,
		Format:         checkFormat,
		MinSeverity:    parseSeverity(checkSeverity),
		ExcludeRules:   parseCommaSeparated(checkExclude),
		IncludeRules:   parseCommaSeparated(checkInclude),
		Verbose:        checkVerbose,
		NoBuiltin:      checkNoBuiltin,
		StrictUnknowns: checkStrict,
	}

	// Apply config file settings if not overridden by CLI
	if infraspecConfig != nil {
		// Only apply config file format if CLI didn't specify
		if !cmd.Flags().Changed("format") && infraspecConfig.Format != "" {
			cfg.Format = infraspecConfig.Format
		}
		// Only apply config file severity if CLI didn't specify
		if !cmd.Flags().Changed("severity") && infraspecConfig.MinSeverity != "" {
			cfg.MinSeverity = parseSeverity(infraspecConfig.MinSeverity)
		}
		// Only apply config file strict if CLI didn't specify
		if !cmd.Flags().Changed("strict") {
			cfg.StrictUnknowns = infraspecConfig.Strict
		}
		// Only apply config file no-builtin if CLI didn't specify
		if !cmd.Flags().Changed("no-builtin") {
			cfg.NoBuiltin = infraspecConfig.NoBuiltin
		}
		// Add rules from config file
		cfg.ConfigRules = infraspecConfig.Rules
	}

	// Create checker instance
	checker, err := gatekeeper.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitError{Code: 2}
	}

	// Handle --list-rules
	if checkListRules {
		return listRules(checker)
	}

	// Discover all .tf files from provided paths
	tfFiles, err := discoverTerraformFiles(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering Terraform files: %v\n", err)
		return ExitError{Code: 2}
	}

	if len(tfFiles) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no Terraform files found in specified paths\n")
		return ExitError{Code: 2}
	}

	// Discover spec files alongside Terraform files
	specFiles, err := discoverSpecFiles(args)
	if err != nil && checkVerbose {
		fmt.Fprintf(os.Stderr, "Warning: error discovering spec files: %v\n", err)
	}

	// Load rules from spec files
	if len(specFiles) > 0 {
		if checkVerbose {
			fmt.Fprintf(os.Stderr, "Found %d spec file(s)\n", len(specFiles))
		}
		if err := checker.LoadSpecFiles(specFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading spec files: %v\n", err)
			return ExitError{Code: 2}
		}
	}

	// Run the checks
	result, err := checker.Check(tfFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running checks: %v\n", err)
		return ExitError{Code: 2}
	}

	// Output the results
	if err := checker.Output(result, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		return ExitError{Code: 2}
	}

	// Return appropriate exit code
	if result.HasViolations() {
		return ExitError{Code: 1}
	}

	return nil
}

// ExitError is an error that carries an exit code
type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

// discoverTerraformFiles finds all .tf files in the given paths
func discoverTerraformFiles(paths []string) ([]string, error) {
	var tfFiles []string
	seen := make(map[string]bool)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot access path %s: %w", path, err)
		}

		if info.IsDir() {
			// Walk directory recursively
			err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip hidden directories
				if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}

				// Collect .tf files
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".tf") {
					absPath, err := filepath.Abs(filePath)
					if err != nil {
						return err
					}
					if !seen[absPath] {
						seen[absPath] = true
						tfFiles = append(tfFiles, absPath)
					}
				}

				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error walking directory %s: %w", path, err)
			}
		} else if strings.HasSuffix(path, ".tf") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("error getting absolute path for %s: %w", path, err)
			}
			if !seen[absPath] {
				seen[absPath] = true
				tfFiles = append(tfFiles, absPath)
			}
		}
	}

	return tfFiles, nil
}

// discoverSpecFiles finds all *.spec.hcl files in the same directories as Terraform files
func discoverSpecFiles(paths []string) ([]string, error) {
	var specFiles []string
	seen := make(map[string]bool)
	dirsChecked := make(map[string]bool)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot access path %s: %w", path, err)
		}

		if info.IsDir() {
			// Walk directory recursively
			err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip hidden directories
				if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}

				// Collect *.spec.hcl files
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".spec.hcl") {
					absPath, err := filepath.Abs(filePath)
					if err != nil {
						return err
					}
					if !seen[absPath] {
						seen[absPath] = true
						specFiles = append(specFiles, absPath)
					}
				}

				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error walking directory %s: %w", path, err)
			}
		} else if strings.HasSuffix(path, ".tf") {
			// Check the directory containing this .tf file for spec files
			dir := filepath.Dir(path)
			absDir, err := filepath.Abs(dir)
			if err != nil {
				continue
			}
			if dirsChecked[absDir] {
				continue
			}
			dirsChecked[absDir] = true

			entries, err := os.ReadDir(absDir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".spec.hcl") {
					specPath := filepath.Join(absDir, entry.Name())
					if !seen[specPath] {
						seen[specPath] = true
						specFiles = append(specFiles, specPath)
					}
				}
			}
		}
	}

	return specFiles, nil
}

func listRules(checker *gatekeeper.Checker) error {
	rules := checker.ListRules()

	if checkFormat == "json" {
		return checker.OutputRulesJSON(rules, os.Stdout)
	}

	// Text output
	fmt.Println("Available Rules:")
	fmt.Println()

	for _, rule := range rules {
		severityBadge := formatSeverityBadge(rule.Severity)
		fmt.Printf("  %s  %-10s  %s\n", severityBadge, rule.ID, rule.Name)
		if checkVerbose && rule.Description != "" {
			fmt.Printf("                      %s\n", rule.Description)
		}
	}

	fmt.Printf("\nTotal: %d rules\n", len(rules))
	return nil
}

func formatSeverityBadge(severity gatekeeper.Severity) string {
	switch severity {
	case gatekeeper.SeverityError:
		return "[ERROR]  "
	case gatekeeper.SeverityWarning:
		return "[WARNING]"
	case gatekeeper.SeverityInfo:
		return "[INFO]   "
	default:
		return "[UNKNOWN]"
	}
}

func parseSeverity(s string) gatekeeper.Severity {
	switch strings.ToLower(s) {
	case "error":
		return gatekeeper.SeverityError
	case "warning", "warn":
		return gatekeeper.SeverityWarning
	case "info":
		return gatekeeper.SeverityInfo
	default:
		return gatekeeper.SeverityError
	}
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
