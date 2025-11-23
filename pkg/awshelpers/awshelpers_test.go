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

func TestNewAuthenticatedSession_VirtualCloudWithoutToken(t *testing.T) {
	// Save original env vars
	originalVirtualCloud := os.Getenv("USE_INFRASPEC_VIRTUAL_CLOUD")
	originalToken := os.Getenv("INFRASPEC_CLOUD_TOKEN")
	defer func() {
		if originalVirtualCloud != "" {
			os.Setenv("USE_INFRASPEC_VIRTUAL_CLOUD", originalVirtualCloud)
		} else {
			os.Unsetenv("USE_INFRASPEC_VIRTUAL_CLOUD")
		}
		if originalToken != "" {
			os.Setenv("INFRASPEC_CLOUD_TOKEN", originalToken)
		} else {
			os.Unsetenv("INFRASPEC_CLOUD_TOKEN")
		}
	}()

	// Enable virtual cloud but don't provide token
	os.Setenv("USE_INFRASPEC_VIRTUAL_CLOUD", "1")
	os.Unsetenv("INFRASPEC_CLOUD_TOKEN")

	// This should fail with clear error message
	_, err := NewAuthenticatedSession("us-east-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "virtual cloud is enabled but no token provided")
	assert.Contains(t, err.Error(), "INFRASPEC_CLOUD_TOKEN")
}

func TestNewAuthenticatedSession_VirtualCloudWithToken(t *testing.T) {
	// Save original env vars
	originalVirtualCloud := os.Getenv("USE_INFRASPEC_VIRTUAL_CLOUD")
	originalToken := os.Getenv("INFRASPEC_CLOUD_TOKEN")
	defer func() {
		if originalVirtualCloud != "" {
			os.Setenv("USE_INFRASPEC_VIRTUAL_CLOUD", originalVirtualCloud)
		} else {
			os.Unsetenv("USE_INFRASPEC_VIRTUAL_CLOUD")
		}
		if originalToken != "" {
			os.Setenv("INFRASPEC_CLOUD_TOKEN", originalToken)
		} else {
			os.Unsetenv("INFRASPEC_CLOUD_TOKEN")
		}
	}()

	// Enable virtual cloud with token
	os.Setenv("USE_INFRASPEC_VIRTUAL_CLOUD", "1")
	os.Setenv("INFRASPEC_CLOUD_TOKEN", "test-token")

	// This should succeed
	cfg, err := NewAuthenticatedSession("us-east-1")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "us-east-1", cfg.Region)
}
