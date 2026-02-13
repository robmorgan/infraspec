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

	// Create some export data
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abc",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-def",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	err := state.Set("dynamodb:export:01234567890123-abc", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:01234567890123-def", export2)
	require.NoError(t, err)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListExports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some export data
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abc",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-def",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	err := state.Set("dynamodb:export:01234567890123-abc", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:01234567890123-def", export2)
	require.NoError(t, err)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abc", summary["ExportArn"])
}

func TestListExports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	for i := 1; i <= 30; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/" + string(rune(i)),
			"ExportStatus": "COMPLETED",
		}
		err := state.Set("dynamodb:export:"+string(rune(i)), export)
		require.NoError(t, err)
	}

	maxResults := int32(10)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 10)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	assert.True(t, hasNextToken)
}
