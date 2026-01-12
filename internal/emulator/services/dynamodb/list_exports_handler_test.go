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
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890123-abcdefgh",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890124-ijklmnop",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export3 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/other-table/export/01234567890125-qrstuvwx",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/other-table",
	}

	err := state.Set("dynamodb:export:export-1", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:export-2", export2)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:export-3", export3)
	require.NoError(t, err)

	// Test list all exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ExportSummaries"].([]interface{})
	assert.Equal(t, 3, len(summaries))
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890123-abcdefgh",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890124-ijklmnop",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export3 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/other-table/export/01234567890125-qrstuvwx",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/other-table",
	}

	err := state.Set("dynamodb:export:export-1", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:export-2", export2)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:export-3", export3)
	require.NoError(t, err)

	// Test filter by table ARN
	tableArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ExportSummaries"].([]interface{})
	assert.Equal(t, 2, len(summaries))
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ExportSummaries"].([]interface{})
	assert.Equal(t, 0, len(summaries))
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test exports
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/0123456789012" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
		}
		err := state.Set("dynamodb:export:export-"+string(rune('0'+i)), export)
		require.NoError(t, err)
	}

	// Test with MaxResults limit
	maxResults := int32(3)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ExportSummaries"].([]interface{})
	assert.Equal(t, 3, len(summaries))

	// Should have NextToken since there are more results
	_, hasNextToken := responseData["NextToken"]
	assert.True(t, hasNextToken)
}
