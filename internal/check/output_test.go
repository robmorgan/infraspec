package check

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/rules"
)

func TestTextFormatter_Format(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{
				RuleID:          "aws-sg-no-public-ssh",
				RuleDescription: "Security groups should not allow SSH",
				ResourceAddress: "aws_security_group.test",
				ResourceType:    "aws_security_group",
				Passed:          false,
				Message:         "Security group allows SSH (port 22) from public internet",
				Severity:        rules.Critical,
				SeverityString:  "critical",
			},
			{
				RuleID:          "aws-s3-encryption-enabled",
				RuleDescription: "S3 buckets should have encryption enabled",
				ResourceAddress: "aws_s3_bucket.test",
				ResourceType:    "aws_s3_bucket",
				Passed:          true,
				Message:         "S3 bucket has encryption enabled",
				Severity:        rules.Warning,
				SeverityString:  "warning",
			},
		},
		TotalChecks:    2,
		Passed:         1,
		Failed:         1,
		CriticalFailed: 1,
		ExitCode:       1,
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(&buf, summary)

	require.NoError(t, err)
	output := buf.String()

	// Check that output contains expected elements
	assert.Contains(t, output, "InfraSpec")
	assert.Contains(t, output, "Pre-flight Check")
	assert.Contains(t, output, "aws-sg-no-public-ssh")
	assert.Contains(t, output, "aws_security_group.test")
	assert.Contains(t, output, "1 passed")
	assert.Contains(t, output, "1 failed")
}

func TestJSONFormatter_Format(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{
				RuleID:          "aws-sg-no-public-ssh",
				RuleDescription: "Security groups should not allow SSH",
				ResourceAddress: "aws_security_group.test",
				ResourceType:    "aws_security_group",
				Passed:          false,
				Message:         "Security group allows SSH (port 22) from public internet",
				Severity:        rules.Critical,
				SeverityString:  "critical",
			},
		},
		TotalChecks:    1,
		Passed:         0,
		Failed:         1,
		CriticalFailed: 1,
		ExitCode:       1,
	}

	formatter := &JSONFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(&buf, summary)

	require.NoError(t, err)

	// Parse the output as JSON
	var parsed Summary
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	assert.Equal(t, 1, parsed.TotalChecks)
	assert.Equal(t, 1, parsed.Failed)
	assert.Equal(t, 1, parsed.ExitCode)
	assert.Len(t, parsed.Results, 1)
	assert.Equal(t, "aws-sg-no-public-ssh", parsed.Results[0].RuleID)
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   string
		expected interface{}
	}{
		{"text", &TextFormatter{}},
		{"json", &JSONFormatter{}},
		{"", &TextFormatter{}},        // Default
		{"unknown", &TextFormatter{}}, // Unknown defaults to text
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			formatter := NewFormatter(tc.format)
			assert.IsType(t, tc.expected, formatter)
		})
	}
}

func TestTextFormatter_EmptySummary(t *testing.T) {
	summary := &Summary{
		Results:     []Result{},
		TotalChecks: 0,
		Passed:      0,
		Failed:      0,
		ExitCode:    0,
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(&buf, summary)

	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "0 passed")
	assert.Contains(t, output, "0 failed")
}

func TestTextFormatter_SeverityBreakdown(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{Passed: false, Severity: rules.Critical},
			{Passed: false, Severity: rules.Warning},
			{Passed: false, Severity: rules.Info},
		},
		TotalChecks:    3,
		Passed:         0,
		Failed:         3,
		CriticalFailed: 1,
		WarningFailed:  1,
		InfoFailed:     1,
		ExitCode:       1,
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(&buf, summary)

	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "1 critical")
	assert.Contains(t, output, "1 warning")
	assert.Contains(t, output, "1 info")
}

func TestTextFormatter_SkippedCount(t *testing.T) {
	summary := &Summary{
		Results:     []Result{},
		TotalChecks: 0,
		Passed:      0,
		Failed:      0,
		Skipped:     5,
		ExitCode:    0,
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(&buf, summary)

	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "5 skipped")
}
