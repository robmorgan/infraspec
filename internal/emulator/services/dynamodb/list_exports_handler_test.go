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

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Empty(t, responseBody.ExportSummaries, "Should have no exports initially")
	assert.Nil(t, responseBody.NextToken, "Should have no next token")
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Export 1
	export1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/01234567890123-abcdef12", tableName)
	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    export1Arn,
		"TableArn":     tableArn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	// Export 2
	export2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/98765432109876-fedcba98", tableName)
	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    export2Arn,
		"TableArn":     tableArn,
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// List all exports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ExportSummaries, 2, "Should have two exports")

	// Verify summaries contain expected fields
	for _, summary := range responseBody.ExportSummaries {
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
	table1Name := "table1"
	table1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1Name)
	export1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/01234567890123-abcdef12", table1Name)
	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    export1Arn,
		"TableArn":     table1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	table2Name := "table2"
	table2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2Name)
	export2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/98765432109876-fedcba98", table2Name)
	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    export2Arn,
		"TableArn":     table2Arn,
		"ExportStatus": "COMPLETED",
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

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ExportSummaries, 1, "Should have only one export for table1")
	assert.Equal(t, export1Arn, *responseBody.ExportSummaries[0].ExportArn)
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	for i := 1; i <= 5; i++ {
		exportArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export%d", tableName, i)
		exportKey := fmt.Sprintf("dynamodb:export:export%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    exportArn,
			"TableArn":     tableArn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	// List exports with max results
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ExportSummaries, 2, "Should have only 2 exports due to max results")

	// Should have NextToken for pagination
	assert.NotNil(t, responseBody.NextToken)
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

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.NotNil(t, responseBody.ExportSummaries)
}
