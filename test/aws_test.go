package test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
)

func TestDynamoDBFeature(t *testing.T) {
	cfg := GetTestConfig(t)
	featurePath := filepath.Join("features", "aws", "dynamodb.feature")
	fixtureDir := filepath.Join(filepath.Dir(featurePath), "fixtures", "dynamodb-with-autoscaling")

	err := writeLocalStackProviderConfig(fixtureDir)
	require.NoError(t, err)

	err = runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}
