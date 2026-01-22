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

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListExports",
		},
		Body:   []byte("{}"),
		Action: "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ExportSummaries")
	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Empty(t, summaries, "Should have no exports initially")
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export1", tableName),
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", tableName),
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// List all exports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have two exports")

	// Verify export summaries contain expected fields
	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok, "Each summary should be an object")
		assert.Contains(t, summaryMap, "ExportArn")
		assert.Contains(t, summaryMap, "ExportStatus")
		assert.Contains(t, summaryMap, "TableArn")
		assert.Equal(t, tableArn, summaryMap["TableArn"])
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create exports for different tables
	table1Name := "table1"
	table1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1Name)

	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export1", table1Name),
		"ExportStatus": "COMPLETED",
		"TableArn":     table1Arn,
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	table2Name := "table2"
	table2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2Name)

	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", table2Name),
		"ExportStatus": "COMPLETED",
		"TableArn":     table2Arn,
		"ExportType":   "FULL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// List exports for table1 only
	input := &ListExportsInput{
		TableArn: &table1Arn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Len(t, summaries, 1, "Should have only one export for table1")

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Arn, summaryMap["TableArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableName := "test-table-paginated"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	for i := 1; i <= 5; i++ {
		exportKey := fmt.Sprintf("dynamodb:export:export%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export%d", tableName, i),
			"ExportStatus": "COMPLETED",
			"TableArn":     tableArn,
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	// List exports with limit
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have only 2 exports due to limit")

	// Should have NextToken for pagination
	assert.Contains(t, responseBody, "NextToken")
}

func TestListExports_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListExports should work with no parameters
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ExportSummaries")
}
