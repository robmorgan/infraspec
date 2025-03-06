package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/robmorgan/infraspec/internal/config"
)

// GetDevelopmentLogger returns a logger suitable for development or testing environments
func GetDevelopmentLogger() *zap.SugaredLogger {
	logger := zap.Must(zap.NewDevelopment())
	defer logger.Sync() //nolint:errcheck

	return logger.Sugar()
}

// GetTestConfig returns a config suitable for testing
func GetTestConfig(t *testing.T) *config.Config {
	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	// override the config for testing
	cfg.Logger = GetDevelopmentLogger()

	return cfg
}

func writeLocalStackProviderConfig(terraformDir string) error {
	providerContent := `
provider "aws" {
  access_key                  = "test"
  secret_key                  = "test"
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
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
