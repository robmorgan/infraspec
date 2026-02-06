package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	export1Key := "dynamodb:export:export-1"
	export1Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export-2"
	export2Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890124-12345679",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// Test ListExports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(summaries))
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create exports for different tables
	export1Key := "dynamodb:export:export-1"
	export1Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table1/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table1",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export-2"
	export2Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table2/export/01234567890124-12345679",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table2",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, len(summaries))

	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ExportArn"], "table1")
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test without any exports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	for i := 1; i <= 5; i++ {
		exportKey := "dynamodb:export:export-" + string(rune('0'+i))
		exportData := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test/export/" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	maxResults := int32(3)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 3)

	// Should have NextToken since we have more results
	if len(summaries) == 3 {
		_, hasNextToken := response["NextToken"]
		assert.True(t, hasNextToken)
	}
}
