package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/test/testhelpers"
)

func TestDynamoDBFeature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", string(os.PathSeparator), "features", "aws", "dynamodb", "dynamodb_table.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}

func TestS3Feature(t *testing.T) {
	cfg := testhelpers.SetupAWSTestsAndConfig()
	featurePath := filepath.Join("..", string(os.PathSeparator), "features", "aws", "s3", "s3_bucket.feature")

	err := runner.New(cfg).Run(featurePath)
	require.NoError(t, err)
}

func TestRdsFeature(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		featurePath string
	}{
		{filepath.Join("..", string(os.PathSeparator), "features", "aws", "rds", "rds_db_instance.feature")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.featurePath, func(t *testing.T) {
			t.Parallel()
			cfg := testhelpers.SetupAWSTestsAndConfig()

			err := runner.New(cfg).Run(testCase.featurePath)
			require.NoError(t, err)
		})
	}
}
