package models

import (
	"testing"
)

func TestGetAWSModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rds", "rds"},
		{"s3", "s3"},
		{"dynamodb", "dynamodb"},
		{"sts", "sts"},
		{"applicationautoscaling", "application-auto-scaling"},
		{"application-auto-scaling", "application-auto-scaling"},
		{"ec2", "ec2"},
		{"lambda", "lambda"},
		{"sqs", "sqs"},
		{"sns", "sns"},
		{"iam", "iam"},
		// Unknown service should return as-is
		{"unknownservice", "unknownservice"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetAWSModelName(tt.input)
			if got != tt.expected {
				t.Errorf("GetAWSModelName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetInfraSpecServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rds", "rds"},
		{"s3", "s3"},
		{"application-auto-scaling", "applicationautoscaling"},
		// Unknown service should return as-is
		{"unknownservice", "unknownservice"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetInfraSpecServiceName(tt.input)
			if got != tt.expected {
				t.Errorf("GetInfraSpecServiceName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
