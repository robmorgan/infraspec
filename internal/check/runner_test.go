package check

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/rules"
)

// testdataPath returns the absolute path to a testdata file in the rules/aws/testdata/plans directory.
func testdataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "..", "rules", "aws", "testdata", "plans", filename)
}

func TestRunner_EmptyPlan(t *testing.T) {
	// Use a plan file that has no resources or only passing resources
	opts := Options{
		PlanPath:    testdataPath("sg_secure.json"),
		MinSeverity: rules.Info,
	}

	runner := NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have results but all should pass
	assert.GreaterOrEqual(t, summary.TotalChecks, 0)
	assert.Equal(t, 0, summary.Failed)
	assert.Equal(t, 0, summary.ExitCode)
}

func TestRunner_IgnoreRules(t *testing.T) {
	// Use a plan that would normally fail the public SSH rule
	opts := Options{
		PlanPath:      testdataPath("sg_public_ssh.json"),
		MinSeverity:   rules.Info,
		IgnoreRuleIDs: []string{"aws-sg-no-public-ssh"},
	}

	runner := NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// The ignored rule should not appear in results
	for _, result := range summary.Results {
		assert.NotEqual(t, "aws-sg-no-public-ssh", result.RuleID)
	}

	// Should have at least one skipped rule
	assert.GreaterOrEqual(t, summary.Skipped, 1)
}

func TestRunner_SeverityFiltering(t *testing.T) {
	// Use a plan with violations
	opts := Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Critical, // Only show critical issues
	}

	runner := NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// All results should be critical severity
	for _, result := range summary.Results {
		assert.Equal(t, rules.Critical, result.Severity)
	}
}

func TestRunner_Violation(t *testing.T) {
	// Use a plan that has a security violation
	opts := Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Info,
	}

	runner := NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have at least one failure
	assert.GreaterOrEqual(t, summary.Failed, 1)
	assert.Equal(t, 1, summary.ExitCode)

	// Find the SSH rule failure
	var foundSSHFailure bool
	for _, result := range summary.Results {
		if result.RuleID == "aws-sg-no-public-ssh" && !result.Passed {
			foundSSHFailure = true
			assert.Contains(t, result.Message, "SSH")
			assert.Equal(t, rules.Critical, result.Severity)
		}
	}
	assert.True(t, foundSSHFailure, "Expected to find aws-sg-no-public-ssh failure")
}

func TestSeverityFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected rules.Severity
		wantErr  bool
	}{
		{"critical", rules.Critical, false},
		{"CRITICAL", rules.Critical, false},
		{"Critical", rules.Critical, false},
		{"warning", rules.Warning, false},
		{"WARNING", rules.Warning, false},
		{"info", rules.Info, false},
		{"INFO", rules.Info, false},
		{"invalid", rules.Info, true},
		{"", rules.Info, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			sev, err := SeverityFromString(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, sev)
			}
		})
	}
}

func TestRunner_MultipleViolations(t *testing.T) {
	// Use a plan with multiple database port violations
	opts := Options{
		PlanPath:    testdataPath("sg_public_databases.json"),
		MinSeverity: rules.Info,
	}

	runner := NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have multiple failures
	assert.GreaterOrEqual(t, summary.Failed, 1)
	assert.Equal(t, 1, summary.ExitCode)
}
