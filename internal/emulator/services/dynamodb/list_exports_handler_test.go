package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some exports to state
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
	}
	err := state.Set("dynamodb:export:01234567890123-abcdef12", export1)
	require.NoError(t, err)

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2/export/01234567890124-abcdef13",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2",
	}
	err = state.Set("dynamodb:export:01234567890124-abcdef13", export2)
	require.NoError(t, err)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
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

	// Add some exports to state
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
	}
	err := state.Set("dynamodb:export:01234567890123-abcdef12", export1)
	require.NoError(t, err)

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2/export/01234567890124-abcdef13",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2",
	}
	err = state.Set("dynamodb:export:01234567890124-abcdef13", export2)
	require.NoError(t, err)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable", summary["TableArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add multiple exports to state
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/export/0123456789012" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		}
		err := state.Set("dynamodb:export:0123456789012"+string(rune('0'+i)), export)
		require.NoError(t, err)
	}

	maxResults := int32(3)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 3)

	// Should have NextToken since we have more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}
