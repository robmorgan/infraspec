package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/check"
	"github.com/robmorgan/infraspec/internal/rules"
)

// testdataPath returns the absolute path to a testdata file in the rules/aws/testdata/plans directory.
func testdataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "..", "internal", "rules", "aws", "testdata", "plans", filename)
}

func TestCheckCommand_PlanFileViolation(t *testing.T) {
	// Test against a plan with a known violation
	opts := check.Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Info,
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have at least one failure and exit code 1
	assert.GreaterOrEqual(t, summary.Failed, 1)
	assert.Equal(t, 1, summary.ExitCode)
}

func TestCheckCommand_SecurePlan(t *testing.T) {
	// Test against a secure plan
	opts := check.Options{
		PlanPath:    testdataPath("sg_secure.json"),
		MinSeverity: rules.Info,
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have no failures
	assert.Equal(t, 0, summary.Failed)
	assert.Equal(t, 0, summary.ExitCode)
}

func TestCheckCommand_SeverityFilter(t *testing.T) {
	// With --severity=critical, info/warning violations should be skipped
	opts := check.Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Critical,
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// All results should be critical severity
	for _, result := range summary.Results {
		assert.Equal(t, rules.Critical, result.Severity)
	}
}

func TestCheckCommand_IgnoreRules(t *testing.T) {
	// Test that --ignore suppresses specific rules
	opts := check.Options{
		PlanPath:      testdataPath("sg_public_ssh.json"),
		MinSeverity:   rules.Info,
		IgnoreRuleIDs: []string{"aws-sg-no-public-ssh"},
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// The ignored rule should not appear in results
	for _, result := range summary.Results {
		assert.NotEqual(t, "aws-sg-no-public-ssh", result.RuleID)
	}
}

func TestCheckCommand_JSONOutput(t *testing.T) {
	opts := check.Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Info,
		Format:      "json",
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Format as JSON
	formatter := check.NewFormatter("json")
	var buf bytes.Buffer
	err = formatter.Format(&buf, summary)
	require.NoError(t, err)

	// Verify JSON is valid
	var parsed check.Summary
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, summary.TotalChecks, parsed.TotalChecks)
	assert.Equal(t, summary.Failed, parsed.Failed)
	assert.Equal(t, summary.ExitCode, parsed.ExitCode)
	assert.Len(t, parsed.Results, len(summary.Results))
}

func TestCheckCommand_TextOutput(t *testing.T) {
	opts := check.Options{
		PlanPath:    testdataPath("sg_public_ssh.json"),
		MinSeverity: rules.Info,
		Format:      "text",
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Format as text
	formatter := check.NewFormatter("text")
	var buf bytes.Buffer
	err = formatter.Format(&buf, summary)
	require.NoError(t, err)

	output := buf.String()

	// Verify output contains expected elements
	assert.Contains(t, output, "InfraSpec")
	assert.Contains(t, output, "Pre-flight Check")
	assert.Contains(t, output, "passed")
	assert.Contains(t, output, "failed")
}

func TestCheckCommand_MultipleIgnoreRules(t *testing.T) {
	opts := check.Options{
		PlanPath:    testdataPath("sg_public_databases.json"),
		MinSeverity: rules.Info,
		IgnoreRuleIDs: []string{
			"aws-sg-no-public-mysql",
			"aws-sg-no-public-postgres",
		},
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// The ignored rules should not appear in results
	for _, result := range summary.Results {
		assert.NotEqual(t, "aws-sg-no-public-mysql", result.RuleID)
		assert.NotEqual(t, "aws-sg-no-public-postgres", result.RuleID)
	}
}

func TestCheckCommand_S3Violations(t *testing.T) {
	// Test S3 rules
	opts := check.Options{
		PlanPath:    testdataPath("s3_public_acl.json"),
		MinSeverity: rules.Info,
	}

	runner := check.NewRunner(opts)
	ctx := context.Background()

	summary, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have at least one S3-related failure
	hasS3Failure := false
	for _, result := range summary.Results {
		if !result.Passed && result.ResourceType == "aws_s3_bucket" {
			hasS3Failure = true
			break
		}
	}
	assert.True(t, hasS3Failure, "Expected S3 rule failures")
}
