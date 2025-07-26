package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestHttpRequestsFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", string(os.PathSeparator), "features", "http", "http_requests.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}

func TestApiTestsFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", string(os.PathSeparator), "features", "http", "api_tests.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}
