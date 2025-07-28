package test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestHttpRequestsFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", "features", "http", "http_requests.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}

func TestApiTestsFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", "features", "http", "api_tests.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}

func TestRetryApiTestsFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", "features", "http", "retry_api_tests.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}
