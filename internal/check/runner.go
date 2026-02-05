package check

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
	_ "github.com/robmorgan/infraspec/internal/rules/aws" // Register AWS rules
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// Runner executes rule checks against Terraform plans.
type Runner struct {
	registry *rules.Registry
	options  *Options
}

// NewRunner creates a new check runner with the given options.
func NewRunner(opts Options) *Runner { //nolint:gocritic // Options passed by value for caller convenience
	return &Runner{
		registry: rules.DefaultRegistry(),
		options:  &opts,
	}
}

// Run executes all applicable rules against the Terraform plan and returns a summary.
func (r *Runner) Run(ctx context.Context) (*Summary, error) {
	p, err := r.getPlan()
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	ignoreSet := r.buildIgnoreSet()
	results, skipped, err := r.evaluateResources(p, ignoreSet)
	if err != nil {
		return nil, err
	}

	return r.buildSummary(results, skipped), nil
}

// buildIgnoreSet creates a set of rule IDs to ignore.
func (r *Runner) buildIgnoreSet() map[string]struct{} {
	ignoreSet := make(map[string]struct{})
	for _, id := range r.options.IgnoreRuleIDs {
		ignoreSet[id] = struct{}{}
	}
	return ignoreSet
}

// evaluateResources runs all applicable rules against each resource in the plan.
func (r *Runner) evaluateResources(p *plan.Plan, ignoreSet map[string]struct{}) ([]Result, int, error) {
	var results []Result
	var skipped int

	for _, rc := range p.ResourceChanges {
		if r.shouldSkipResource(rc) {
			continue
		}

		rcResults, rcSkipped, err := r.evaluateResourceRules(rc, ignoreSet)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, rcResults...)
		skipped += rcSkipped
	}

	return results, skipped, nil
}

// shouldSkipResource returns true if the resource should be skipped.
func (r *Runner) shouldSkipResource(rc *plan.ResourceChange) bool {
	return rc.IsDataSource() || rc.IsNoOp()
}

// evaluateResourceRules runs all applicable rules against a single resource.
func (r *Runner) evaluateResourceRules(rc *plan.ResourceChange, ignoreSet map[string]struct{}) ([]Result, int, error) {
	var results []Result
	var skipped int

	applicableRules := r.registry.RulesForResource(rc.Type)
	for _, rule := range applicableRules {
		if _, ignored := ignoreSet[rule.ID()]; ignored {
			skipped++
			continue
		}

		if rule.Severity() < r.options.MinSeverity {
			skipped++
			continue
		}

		result, err := rule.Check(rc)
		if err != nil {
			return nil, 0, fmt.Errorf("rule %s failed: %w", rule.ID(), err)
		}

		results = append(results, Result{
			RuleID:          rule.ID(),
			RuleDescription: rule.Description(),
			ResourceAddress: rc.Address,
			ResourceType:    rc.Type,
			Passed:          result.Passed,
			Message:         result.Message,
			Severity:        result.Severity,
			SeverityString:  result.Severity.String(),
		})
	}

	return results, skipped, nil
}

// getPlan retrieves the Terraform plan, either from a file or by running terraform plan.
func (r *Runner) getPlan() (*plan.Plan, error) {
	if r.options.PlanPath != "" {
		return plan.ParsePlanFile(r.options.PlanPath)
	}

	// Use plan runner to execute terraform plan
	dir := r.options.Dir
	if dir == "" {
		dir = "."
	}

	opts := &iacprovisioner.Options{
		WorkingDir: dir,
	}
	return plan.NewRunner(opts).Run()
}

// buildSummary creates a Summary from the check results.
func (r *Runner) buildSummary(results []Result, skipped int) *Summary {
	summary := &Summary{
		Results:     results,
		TotalChecks: len(results),
		Skipped:     skipped,
	}

	for _, result := range results {
		if result.Passed {
			summary.Passed++
		} else {
			summary.Failed++
			switch result.Severity {
			case rules.Critical:
				summary.CriticalFailed++
			case rules.Warning:
				summary.WarningFailed++
			case rules.Info:
				summary.InfoFailed++
			}
		}
	}

	// Exit code: 1 if any failures, 0 otherwise
	if summary.Failed > 0 {
		summary.ExitCode = 1
	}

	return summary
}
