package testhelpers

import (
	"os"

	"github.com/robmorgan/infraspec/internal/config"
)

var envVarsToCleanup []string

// GetTestConfig returns a config suitable for testing
func GetTestConfig() *config.Config {
	cfg, err := config.DefaultConfig()
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
		"AWS_ACCESS_KEY_ID":     "test",
		"AWS_SECRET_ACCESS_KEY": "test",
		"AWS_DEFAULT_REGION":    "us-east-1",
		"AWS_ENDPOINT_URL":      "http://localhost:4566",
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

// SetupAWSTests sets up the AWS test environment and returns the config
func SetupAWSTestsAndConfig() *config.Config {
	SetupAwsTestEnvironment()
	return GetTestConfig()
}
