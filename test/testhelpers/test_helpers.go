package testhelpers

import (
	"os"

	"github.com/robmorgan/infraspec/internal/config"
)

var envVarsToCleanup []string

// GetTestConfig returns a config suitable for testing
func GetTestConfig() *config.Config {
	cfg, err := config.LoadConfig("", false)
	if err != nil {
		panic(err)
	}

	// override the config for testing
	config.Logging.SetDevelopmentLogger()

	return cfg
}

// SetupAwsTestEnvironment sets up the AWS environment for testing
func SetupAwsTestEnvironment() {
	// Only set if not already present (allows external override)
	envVars := map[string]string{
		"AWS_ACCESS_KEY_ID":      "infraspec-test",
		"AWS_SECRET_ACCESS_KEY":  "securetoken",
		"AWS_DEFAULT_REGION":     "us-east-1",
		"INFRASPEC_BEARER_TOKEN": "test-token",
	}

	// Only set AWS_ENDPOINT_URL to localhost if Virtual Cloud mode is NOT enabled
	// When Virtual Cloud is enabled, we want to use the InfraSpec Cloud API
	if os.Getenv("USE_INFRASPEC_VIRTUAL_CLOUD") != "1" {
		envVars["AWS_ENDPOINT_URL"] = "http://localhost:3687"
	}

	for key, value := range envVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
			envVarsToCleanup = append(envVarsToCleanup, key)
		}
	}
}

// CleanupAwsTestEnvironment cleans up the AWS environment after testing
func CleanupAwsTestEnvironment() {
	for _, key := range envVarsToCleanup {
		os.Unsetenv(key)
	}
}

// SetupAWSTestsAndConfig sets up the AWS test environment and returns the config
func SetupAWSTestsAndConfig() *config.Config {
	SetupAwsTestEnvironment()
	return GetTestConfig()
}
