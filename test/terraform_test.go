package test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestHelloWorldFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("../", "features", "terraform", "hello_world.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}
