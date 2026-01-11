package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some export entries in state
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-87654321",
		"ExportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"ExportType":   "INCREMENTAL_EXPORT",
	}

	err := state.Set("dynamodb:export:01234567890123-12345678", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:01234567890123-87654321", export2)
	require.NoError(t, err)

	// Test listing all exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some export entries in state
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-12345678",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-87654321",
		"ExportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"ExportType":   "FULL_EXPORT",
	}

	err := state.Set("dynamodb:export:01234567890123-12345678", export1)
	require.NoError(t, err)
	err = state.Set("dynamodb:export:01234567890123-87654321", export2)
	require.NoError(t, err)

	// Test filtering by table ARN
	input := &ListExportsInput{
		TableArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"),
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1", summary["TableArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple export entries
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/0123456789012-%d", i),
			"ExportStatus": "COMPLETED",
			"TableArn":     fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table-%d", i),
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(fmt.Sprintf("dynamodb:export:0123456789012-%d", i), export)
		require.NoError(t, err)
	}

	// Test first page
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken
	nextToken, ok := result["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Test second page
	input2 := &ListExportsInput{
		MaxResults: &maxResults,
		NextToken:  &nextToken,
	}
	resp2, err := service.listExports(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)

	summaries2, ok := result2["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries2, 2)
}

func TestListExports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, summaries)
}
