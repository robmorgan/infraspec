package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/testing/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListImports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - imports
	importData1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":      "my-bucket",
			"S3KeyPrefix":   "data/",
			"S3BucketOwner": "000000000000",
		},
	}
	err := state.Set("dynamodb:import:import1", importData1)
	require.NoError(t, err)

	importData2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890124-abcdef13",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
	}
	err = state.Set("dynamodb:import:import2", importData2)
	require.NoError(t, err)

	importData3 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/import/01234567890125-abcdef14",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "ION",
	}
	err = state.Set("dynamodb:import:import3", importData3)
	require.NoError(t, err)

	// Test listing all imports
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 3)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	importData1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	err := state.Set("dynamodb:import:import1", importData1)
	require.NoError(t, err)

	importData2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890124-abcdef13",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
	}
	err = state.Set("dynamodb:import:import2", importData2)
	require.NoError(t, err)

	importData3 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/import/01234567890125-abcdef14",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "ION",
	}
	err = state.Set("dynamodb:import:import3", importData3)
	require.NoError(t, err)

	// Test filtering by table ARN
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 2)
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Empty(t, output.ImportSummaryList)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - 5 imports
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/0123456789012" + string(rune('0'+i)),
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
		}
		err := state.Set("dynamodb:import:import"+string(rune('0'+i)), importData)
		require.NoError(t, err)
	}

	// Request with page size of 2
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 2)
	assert.NotNil(t, output.NextToken)
}
