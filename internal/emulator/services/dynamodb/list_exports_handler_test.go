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
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "COMPLETED",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	require.NoError(t, state.Set("dynamodb:export:test-export-1", export1))
	require.NoError(t, state.Set("dynamodb:export:test-export-2", export2))

	// Test without filter
	input := &ListExportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(summaries))
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn1,
	}

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "COMPLETED",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     tableArn2,
	}

	require.NoError(t, state.Set("dynamodb:export:test-export-1", export1))
	require.NoError(t, state.Set("dynamodb:export:test-export-2", export2))

	// Test with table ARN filter
	input := &ListExportsInput{
		TableArn: &tableArn1,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(summaries))

	// Verify it's the correct export
	firstSummary := summaries[0].(map[string]interface{})
	assert.Contains(t, firstSummary["ExportArn"], "test-table")
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	for i := 0; i < 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/" + string(rune('a'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		}
		require.NoError(t, state.Set("dynamodb:export:test-export-"+string(rune('a'+i)), export))
	}

	// Test with max results
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(summaries))

	// Should have NextToken since there are more results
	_, hasNextToken := result["NextToken"]
	assert.True(t, hasNextToken)
}
