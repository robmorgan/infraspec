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

	// Create some test exports
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	state.Set("dynamodb:export:01234567890123-abcdef12", export1)
	state.Set("dynamodb:export:01234567890123-ghijkl34", export2)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test exports
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	state.Set("dynamodb:export:01234567890123-abcdef12", export1)
	state.Set("dynamodb:export:01234567890123-ghijkl34", export2)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ExportArn"], "test-table")
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListExports_WithMaxResults(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test exports
	for i := 0; i < 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/" + string(rune(i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		}
		state.Set("dynamodb:export:"+string(rune(i)), export)
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := responseData["NextToken"]
	assert.True(t, hasNextToken)
}
