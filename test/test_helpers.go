package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/config"
)

// GetTestConfig returns a config suitable for testing
func GetTestConfig(t *testing.T) *config.Config {
	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	// override the config for testing
	config.Logging.SetDevelopmentLogger()

	return cfg
}

func configureEnvForTests() {
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	os.Setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
}

func writeLocalStackProviderConfig(terraformDir string) error {
	providerContent := `
provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true

  endpoints {
    dynamodb = "http://localhost:4566"
    s3       = "http://localhost:4566"
    # Add other services as needed
  }
}
`
	return os.WriteFile(filepath.Join(terraformDir, "provider_override.tf"), []byte(providerContent), 0644)
}
