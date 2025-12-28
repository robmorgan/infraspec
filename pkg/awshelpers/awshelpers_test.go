package awshelpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestMain(m *testing.M) {
	testhelpers.SetupAWSTestsAndConfig()
	code := m.Run()
	testhelpers.CleanupAwsTestEnvironment()
	os.Exit(code)
}

func TestNewAuthenticatedSession_LocalhostEndpoint(t *testing.T) {
	// Save original env vars
	originalEndpoint := os.Getenv("AWS_ENDPOINT_URL")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("AWS_ENDPOINT_URL", originalEndpoint)
		} else {
			os.Unsetenv("AWS_ENDPOINT_URL")
		}
	}()

	// Set localhost endpoint (embedded emulator mode)
	os.Setenv("AWS_ENDPOINT_URL", "http://localhost:8000")

	// Should succeed with dummy credentials
	cfg, err := NewAuthenticatedSession("us-east-1")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "us-east-1", cfg.Region)
}

func TestNewAuthenticatedSession_127001Endpoint(t *testing.T) {
	// Save original env vars
	originalEndpoint := os.Getenv("AWS_ENDPOINT_URL")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("AWS_ENDPOINT_URL", originalEndpoint)
		} else {
			os.Unsetenv("AWS_ENDPOINT_URL")
		}
	}()

	// Set 127.0.0.1 endpoint (embedded emulator mode)
	os.Setenv("AWS_ENDPOINT_URL", "http://127.0.0.1:9000")

	// Should succeed with dummy credentials
	cfg, err := NewAuthenticatedSession("us-west-2")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "us-west-2", cfg.Region)
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		endpoint string
		expected bool
	}{
		{"http://localhost:8000", true},
		{"http://127.0.0.1:8000", true},
		{"http://[::1]:8000", true},
		{"https://localhost", true},
		{"https://example.com", false},
		{"http://api.example.com:8000", false},
		{"", false},
		{"://invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			result := isLocalhost(tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}
