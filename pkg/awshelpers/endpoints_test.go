package awshelpers

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVirtualCloudEndpoint(t *testing.T) {
	tests := []struct {
		name                  string
		service               string
		awsEndpointURL        string
		awsServiceEndpointURL string
		expectedEndpoint      string
		expectedOk            bool
	}{
		{
			name:             "no endpoint URL returns empty",
			service:          "dynamodb",
			awsEndpointURL:   "",
			expectedEndpoint: "",
			expectedOk:       false,
		},
		{
			name:                  "service-specific env var takes precedence",
			service:               "dynamodb",
			awsEndpointURL:        "http://localhost:3687",
			awsServiceEndpointURL: "http://custom-dynamodb:9000",
			expectedEndpoint:      "http://custom-dynamodb:9000",
			expectedOk:            true,
		},
		{
			name:             "builds nip.io endpoint for localhost",
			service:          "rds",
			awsEndpointURL:   "http://localhost:3687",
			expectedEndpoint: "http://rds.127.0.0.1.nip.io:3687",
			expectedOk:       true,
		},
		{
			name:             "builds nip.io endpoint for 127.0.0.1",
			service:          "s3",
			awsEndpointURL:   "http://127.0.0.1:9000",
			expectedEndpoint: "http://s3.127.0.0.1.nip.io:9000",
			expectedOk:       true,
		},
		{
			name:             "returns base endpoint when service is empty",
			service:          "",
			awsEndpointURL:   "http://localhost:3687",
			expectedEndpoint: "http://localhost:3687",
			expectedOk:       true,
		},
		{
			name:             "builds subdomain for non-localhost endpoint",
			service:          "ec2",
			awsEndpointURL:   "https://example.com",
			expectedEndpoint: "https://ec2.example.com",
			expectedOk:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origEndpointURL := os.Getenv("AWS_ENDPOINT_URL")
			var origServiceEndpointURL string
			if tt.service != "" {
				serviceEnvVar := "AWS_ENDPOINT_URL_" + strings.ToUpper(tt.service)
				origServiceEndpointURL = os.Getenv(serviceEnvVar)
				defer os.Setenv(serviceEnvVar, origServiceEndpointURL)
			}
			defer os.Setenv("AWS_ENDPOINT_URL", origEndpointURL)

			// Setup test environment
			if tt.awsEndpointURL != "" {
				os.Setenv("AWS_ENDPOINT_URL", tt.awsEndpointURL)
			} else {
				os.Unsetenv("AWS_ENDPOINT_URL")
			}

			if tt.awsServiceEndpointURL != "" && tt.service != "" {
				serviceEnvVar := "AWS_ENDPOINT_URL_" + strings.ToUpper(tt.service)
				os.Setenv(serviceEnvVar, tt.awsServiceEndpointURL)
			}

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
			name:         "builds subdomain for non-localhost",
			baseEndpoint: "https://example.com",
			subdomain:    "dynamodb",
			expected:     "https://dynamodb.example.com",
		},
		{
			name:         "builds nip.io for localhost",
			baseEndpoint: "http://localhost:3687",
			subdomain:    "rds",
			expected:     "http://rds.127.0.0.1.nip.io:3687",
		},
		{
			name:         "builds nip.io for 127.0.0.1",
			baseEndpoint: "http://127.0.0.1:3687",
			subdomain:    "s3",
			expected:     "http://s3.127.0.0.1.nip.io:3687",
		},
		{
			name:         "builds nip.io for ::1",
			baseEndpoint: "http://[::1]:3687",
			subdomain:    "sts",
			expected:     "http://sts.127.0.0.1.nip.io:3687",
		},
		{
			name:         "builds nip.io for localhost without port",
			baseEndpoint: "http://localhost",
			subdomain:    "ec2",
			expected:     "http://ec2.127.0.0.1.nip.io",
		},
		{
			name:         "handles custom domain with port",
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
