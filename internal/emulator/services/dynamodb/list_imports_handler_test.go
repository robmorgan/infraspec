package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListImports_Success(t *testing.T) {
	service, state := setupTestService(t)

	// Create test import data
	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/import/01234567890123-abcdef12",
		"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "DYNAMODB_JSON",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}
	import2 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-abcdef34",
		"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"ImportStatus":          "IN_PROGRESS",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef12", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef34", import2))

	// Test listing all imports
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	service, state := setupTestService(t)

	// Create test import data
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/import/01234567890123-abcdef12",
		"TableArn":              tableArn1,
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "DYNAMODB_JSON",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}
	import2 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-abcdef34",
		"TableArn":              tableArn2,
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef12", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef34", import2))

	// Test listing imports for specific table
	input := &ListImportsInput{
		TableArn: &tableArn1,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 1)
	assert.Equal(t, tableArn1, *output.ImportSummaryList[0].TableArn)
}

func TestListImports_EmptyList(t *testing.T) {
	service, _ := setupTestService(t)

	// Test listing when no imports exist
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 0)
}

func TestListImports_Pagination(t *testing.T) {
	service, state := setupTestService(t)

	// Create multiple imports
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import-" + string(rune('0'+i)),
			"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ImportStatus":          "COMPLETED",
			"InputFormat":           "DYNAMODB_JSON",
			"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
		}
		key := "dynamodb:import:import-" + string(rune('0'+i))
		require.NoError(t, state.Set(key, importData))
	}

	// Test pagination with PageSize
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 2)
	assert.NotNil(t, output.NextToken)
}
