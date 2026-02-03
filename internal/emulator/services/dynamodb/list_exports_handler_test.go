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
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export1",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export2",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     tableArn,
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// List all exports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, output.ExportSummaries, 2, "Should have two exports")

	// Verify summaries contain expected fields
	for _, summary := range output.ExportSummaries {
		assert.NotNil(t, summary.ExportArn)
		assert.NotEmpty(t, summary.ExportStatus)
		assert.NotEmpty(t, summary.ExportType)
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create exports for different tables
	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table1/export/export1",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table1Arn,
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2"
	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table2/export/export2",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table2Arn,
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

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 1, "Should have only one export for table1")
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	for i := 1; i <= 5; i++ {
		exportKey := fmt.Sprintf("dynamodb:export:export%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export%d", i),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     tableArn,
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

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 2, "Should have only 2 exports due to limit")

	// Should have NextToken for pagination
	assert.NotNil(t, output.NextToken, "Should have NextToken when there are more results")
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

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.ExportSummaries)
}
