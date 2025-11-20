package awshelpers

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/robmorgan/infraspec/internal/config"
)

func TestGetVirtualCloudEndpoint(t *testing.T) {
	tests := []struct {
		name                  string
		service               string
		virtualCloudEnabled   bool
		awsEndpointURL        string
		awsServiceEndpointURL string
		expectedEndpoint      string
		expectedOk            bool
	}{
		{
			name:                "virtual cloud disabled returns empty",
			service:             "dynamodb",
			virtualCloudEnabled: false,
			expectedEndpoint:    "",
			expectedOk:          false,
		},
		{
			name:                  "service-specific env var takes precedence",
			service:               "dynamodb",
			virtualCloudEnabled:   true,
			awsEndpointURL:        "http://localhost:8000",
			awsServiceEndpointURL: "http://custom-dynamodb:9000",
			expectedEndpoint:      "http://custom-dynamodb:9000",
			expectedOk:            true,
		},
		{
			name:                "builds subdomain from AWS_ENDPOINT_URL when set",
			service:             "rds",
			virtualCloudEnabled: true,
			awsEndpointURL:      "http://localhost:8000",
			expectedEndpoint:    "http://rds.localhost:8000",
			expectedOk:          true,
		},
		{
			name:                "builds subdomain from default endpoint when no env vars",
			service:             "s3",
			virtualCloudEnabled: true,
			expectedEndpoint:    "https://s3.infraspec.sh",
			expectedOk:          true,
		},
		{
			name:                "returns base endpoint when service is empty with AWS_ENDPOINT_URL",
			service:             "",
			virtualCloudEnabled: true,
			awsEndpointURL:      "http://localhost:8000",
			expectedEndpoint:    "http://localhost:8000",
			expectedOk:          true,
		},
		{
			name:                "returns base endpoint when service is empty with default",
			service:             "",
			virtualCloudEnabled: true,
			expectedEndpoint:    "https://infraspec.sh",
			expectedOk:          true,
		},
		{
			name:                "builds subdomain for ec2 from default endpoint",
			service:             "ec2",
			virtualCloudEnabled: true,
			expectedEndpoint:    "https://ec2.infraspec.sh",
			expectedOk:          true,
		},
		{
			name:                "builds subdomain for sts from localhost",
			service:             "sts",
			virtualCloudEnabled: true,
			awsEndpointURL:      "http://localhost:8000",
			expectedEndpoint:    "http://sts.localhost:8000",
			expectedOk:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origVirtualCloud := os.Getenv(config.UseInfraspecVirtualCloudEnvVar)
			origEndpointURL := os.Getenv("AWS_ENDPOINT_URL")
			var origServiceEndpointURL string
			if tt.service != "" {
				serviceEnvVar := "AWS_ENDPOINT_URL_" + strings.ToUpper(tt.service)
				origServiceEndpointURL = os.Getenv(serviceEnvVar)
				defer os.Setenv(serviceEnvVar, origServiceEndpointURL)
			}
			defer os.Setenv(config.UseInfraspecVirtualCloudEnvVar, origVirtualCloud)
			defer os.Setenv("AWS_ENDPOINT_URL", origEndpointURL)

			// Setup test environment
			if tt.virtualCloudEnabled {
				os.Setenv(config.UseInfraspecVirtualCloudEnvVar, "true")
			} else {
				os.Unsetenv(config.UseInfraspecVirtualCloudEnvVar)
			}

			if tt.awsEndpointURL != "" {
				os.Setenv("AWS_ENDPOINT_URL", tt.awsEndpointURL)
			} else {
				os.Unsetenv("AWS_ENDPOINT_URL")
			}

			if tt.awsServiceEndpointURL != "" && tt.service != "" {
				serviceEnvVar := "AWS_ENDPOINT_URL_" + strings.ToUpper(tt.service)
				os.Setenv(serviceEnvVar, tt.awsServiceEndpointURL)
			}

			// Reset config to pick up new env vars
			config.LoadConfig("", tt.virtualCloudEnabled)

			// Execute test
			endpoint, ok := GetVirtualCloudEndpoint(tt.service)

			// Verify results
			assert.Equal(t, tt.expectedOk, ok, "ok value mismatch")
			assert.Equal(t, tt.expectedEndpoint, endpoint, "endpoint mismatch")
		})
	}
}

func TestBuildServiceEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		baseEndpoint string
		subdomain    string
		expected     string
	}{
		{
			name:         "builds infraspec.sh subdomain",
			baseEndpoint: "https://infraspec.sh",
			subdomain:    "dynamodb",
			expected:     "https://dynamodb.infraspec.sh",
		},
		{
			name:         "builds localhost subdomain with port",
			baseEndpoint: "http://localhost:8000",
			subdomain:    "rds",
			expected:     "http://rds.localhost:8000",
		},
		{
			name:         "builds s3 subdomain",
			baseEndpoint: "https://infraspec.sh",
			subdomain:    "s3",
			expected:     "https://s3.infraspec.sh",
		},
		{
			name:         "builds sts subdomain for localhost",
			baseEndpoint: "http://localhost:8000",
			subdomain:    "sts",
			expected:     "http://sts.localhost:8000",
		},
		{
			name:         "handles custom domain",
			baseEndpoint: "https://api.example.com",
			subdomain:    "ec2",
			expected:     "https://ec2.api.example.com",
		},
		{
			name:         "handles custom port",
			baseEndpoint: "http://api.example.com:9000",
			subdomain:    "ssm",
			expected:     "http://ssm.api.example.com:9000",
		},
		{
			name:         "returns base on parse error",
			baseEndpoint: "://invalid-url",
			subdomain:    "dynamodb",
			expected:     "://invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildServiceEndpoint(tt.baseEndpoint, tt.subdomain)
			assert.Equal(t, tt.expected, result)
		})
	}
}
