package test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
)

func TestHelloWorldFeature(t *testing.T) {
	cfg := GetTestConfig(t)
	featurePath := filepath.Join("features", "terraform", "helloworld.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}
