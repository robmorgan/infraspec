// Package gatekeeper provides static analysis of Terraform configurations
// against security rules before applying infrastructure changes.
package gatekeeper

import (
	"fmt"
	"io"
	"time"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/engine"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/output"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/parser"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules/builtin"
)

// Severity represents the severity level of a rule or violation
type Severity = rules.Severity

// Severity constants
const (
	SeverityError   = rules.SeverityError
	SeverityWarning = rules.SeverityWarning
	SeverityInfo    = rules.SeverityInfo
)

// Config holds the configuration for the gatekeeper checker
type Config struct {
	RulesFile      string       // Path to custom rules HCL file
	VarFile        string       // Path to tfvars file for variable resolution
	Format         string       // Output format (text, json)
	MinSeverity    Severity     // Minimum severity to report
	ExcludeRules   []string     // Rule IDs to exclude
	IncludeRules   []string     // Rule IDs to include (excludes all others)
	Verbose        bool         // Enable verbose output
	NoBuiltin      bool         // Disable built-in rules
	StrictUnknowns bool         // Treat unknown values as violations
	ConfigRules    []rules.Rule // Rules from .infraspec.hcl config file
}

// Checker performs static analysis on Terraform configurations
type Checker struct {
	config    Config
	allRules  []rules.Rule // All loaded rules (for listing)
	rules     []rules.Rule // Filtered rules (for checking)
	parser    *parser.Parser
	engine    *engine.Engine
	formatter output.Formatter
}

// CheckResult contains the results of a check operation
type CheckResult struct {
	Violations     []engine.Violation
	FilesScanned   int
	ResourcesFound int
	RulesEvaluated int
	Duration       time.Duration
}

// HasViolations returns true if there are any violations at or above the minimum severity
func (r *CheckResult) HasViolations() bool {
	return len(r.Violations) > 0
}

// RuleSummary contains summary information about a rule for listing
type RuleSummary struct {
	ID          string
	Name        string
	Description string
	Severity    Severity
	Tags        []string
}

// New creates a new Checker with the given configuration
func New(cfg Config) (*Checker, error) {
	c := &Checker{
		config: cfg,
		parser: parser.New(parser.Config{
			VarFile: cfg.VarFile,
		}),
	}

	// Load rules
	if err := c.loadRules(); err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Create engine
	c.engine = engine.New(engine.Config{
		StrictUnknowns: cfg.StrictUnknowns,
	})

	// Create formatter
	var err error
	c.formatter, err = output.NewFormatter(cfg.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create formatter: %w", err)
	}

	return c, nil
}

// loadRules loads both built-in and custom rules
func (c *Checker) loadRules() error {
	var allRules []rules.Rule
	seenIDs := make(map[string]bool)

	// Load built-in rules unless disabled
	if !c.config.NoBuiltin {
		builtinRules, err := builtin.LoadBuiltinRules()
		if err != nil {
			return fmt.Errorf("failed to load built-in rules: %w", err)
		}
		for _, r := range builtinRules {
			seenIDs[r.ID] = true
		}
		allRules = append(allRules, builtinRules...)
	}

	// Load rules from .infraspec.hcl config file
	for _, r := range c.config.ConfigRules {
		if seenIDs[r.ID] {
			// Config rules override built-in rules with same ID
			for i, existing := range allRules {
				if existing.ID == r.ID {
					allRules[i] = r
					break
				}
			}
		} else {
			seenIDs[r.ID] = true
			allRules = append(allRules, r)
		}
	}

	// Load custom rules if specified
	if c.config.RulesFile != "" {
		customRules, err := rules.LoadFromFile(c.config.RulesFile)
		if err != nil {
			return fmt.Errorf("failed to load custom rules from %s: %w", c.config.RulesFile, err)
		}
		for _, r := range customRules {
			if seenIDs[r.ID] {
				// Custom rules override previous rules with same ID
				for i, existing := range allRules {
					if existing.ID == r.ID {
						allRules[i] = r
						break
					}
				}
			} else {
				seenIDs[r.ID] = true
				allRules = append(allRules, r)
			}
		}
	}

	// Store all rules (for listing)
	c.allRules = allRules

	// Apply include/exclude filters (for checking)
	c.rules = c.filterRules(allRules)

	return nil
}

// LoadSpecFiles loads additional rules from spec files
func (c *Checker) LoadSpecFiles(specFiles []string) error {
	seenIDs := make(map[string]bool)
	for _, r := range c.allRules {
		seenIDs[r.ID] = true
	}

	for _, path := range specFiles {
		specRules, err := rules.LoadFromHCLFile(path)
		if err != nil {
			return fmt.Errorf("failed to load spec file %s: %w", path, err)
		}

		for _, r := range specRules {
			if seenIDs[r.ID] {
				// Spec rules override previous rules with same ID
				for i, existing := range c.allRules {
					if existing.ID == r.ID {
						c.allRules[i] = r
						break
					}
				}
			} else {
				seenIDs[r.ID] = true
				c.allRules = append(c.allRules, r)
			}
		}
	}

	// Re-apply filters
	c.rules = c.filterRules(c.allRules)

	return nil
}

// filterRules applies include/exclude filters to the rule set
func (c *Checker) filterRules(allRules []rules.Rule) []rules.Rule {
	// Build exclude set
	excludeSet := make(map[string]bool)
	for _, id := range c.config.ExcludeRules {
		excludeSet[id] = true
	}

	// Build include set (if specified)
	var includeSet map[string]bool
	if len(c.config.IncludeRules) > 0 {
		includeSet = make(map[string]bool)
		for _, id := range c.config.IncludeRules {
			includeSet[id] = true
		}
	}

	var filtered []rules.Rule
	for _, rule := range allRules {
		// Skip if excluded
		if excludeSet[rule.ID] {
			continue
		}

		// Skip if include list is specified and rule is not in it
		if includeSet != nil && !includeSet[rule.ID] {
			continue
		}

		// Skip if below minimum severity
		if rule.Severity > c.config.MinSeverity {
			continue
		}

		filtered = append(filtered, rule)
	}

	return filtered
}

// Check runs the gatekeeper checks on the given Terraform files
func (c *Checker) Check(tfFiles []string) (*CheckResult, error) {
	startTime := time.Now()

	result := &CheckResult{
		FilesScanned:   len(tfFiles),
		RulesEvaluated: len(c.rules),
	}

	// Parse all Terraform files
	var allResources []parser.Resource
	for _, file := range tfFiles {
		resources, err := c.parser.ParseFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		allResources = append(allResources, resources...)
	}

	result.ResourcesFound = len(allResources)

	// Evaluate rules against resources
	violations := c.engine.Evaluate(c.rules, allResources)

	// Filter violations by severity
	for _, v := range violations {
		if v.Severity <= c.config.MinSeverity {
			result.Violations = append(result.Violations, v)
		}
	}

	result.Duration = time.Since(startTime)

	return result, nil
}

// Output writes the check results to the given writer
func (c *Checker) Output(result *CheckResult, w io.Writer) error {
	return c.formatter.Format(output.Result{
		Violations:     result.Violations,
		FilesScanned:   result.FilesScanned,
		ResourcesFound: result.ResourcesFound,
		RulesEvaluated: result.RulesEvaluated,
		Duration:       result.Duration,
		Passed:         !result.HasViolations(),
	}, w)
}

// ListRules returns a summary of all loaded rules (before filtering)
func (c *Checker) ListRules() []RuleSummary {
	summaries := make([]RuleSummary, len(c.allRules))
	for i, rule := range c.allRules {
		summaries[i] = RuleSummary{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Tags:        rule.Tags,
		}
	}
	return summaries
}

// OutputRulesJSON outputs the rules in JSON format
func (c *Checker) OutputRulesJSON(summaries []RuleSummary, w io.Writer) error {
	return c.formatter.FormatRules(summaries, w)
}
