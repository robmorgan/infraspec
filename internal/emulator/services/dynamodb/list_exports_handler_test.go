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

	// Create test data - exports
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-a1b2c3d4",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/export/01234567890124-a1b2c3d5",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:export1", export1))
	require.NoError(t, state.Set("dynamodb:export:export2", export2))

	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-a1b2c3d4",
		"TableArn":     tableArn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/export/01234567890124-a1b2c3d5",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:export1", export1))
	require.NoError(t, state.Set("dynamodb:export:export2", export2))

	input := &ListExportsInput{
		TableArn: &tableArn,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	// Verify it's the correct export
	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ExportArn"], "test-table")
}

func TestListExports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	for i := 0; i < 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-a1b2c3d" + string(rune('0'+i)),
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		require.NoError(t, state.Set("dynamodb:export:export"+string(rune('0'+i)), export))
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := output["NextToken"]
	assert.True(t, hasNextToken)
}
