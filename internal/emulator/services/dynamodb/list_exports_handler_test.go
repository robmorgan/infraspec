package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some export entries
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2/export/01234567890123-87654321",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-12345678", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-87654321", export2))

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some export entries
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2/export/01234567890123-87654321",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-12345678", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-87654321", export2))

	input := &ListExportsInput{
		TableArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/01234567890123-12345678", summary["ExportArn"])
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create several export entries
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/export/0123456789012" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
		}
		require.NoError(t, state.Set("dynamodb:export:0123456789012"+string(rune('0'+i)), export))
	}

	// Request with MaxResults = 2
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}

func TestListExports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 0)
}
