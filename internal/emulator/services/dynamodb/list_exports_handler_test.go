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
	assert.NotNil(t, responseBody.ExportSummaries)
	assert.Empty(t, responseBody.ExportSummaries, "Should have no exports initially")
	assert.Nil(t, responseBody.NextToken)
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
		"TableArn":     tableArn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", tableName),
		"TableArn":     tableArn,
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

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

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ExportSummaries, 2)

	// Check first export
	assert.NotNil(t, responseBody.ExportSummaries[0].ExportArn)
	assert.Equal(t, ExportStatus("COMPLETED"), responseBody.ExportSummaries[0].ExportStatus)
	assert.Equal(t, ExportType("FULL_EXPORT"), responseBody.ExportSummaries[0].ExportType)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"
	tableArn1 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1)
	tableArn2 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2)

	export1Key := "dynamodb:export:export1"
	export1Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export1", table1),
		"TableArn":     tableArn1,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export2"
	export2Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", table2),
		"TableArn":     tableArn2,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
	reqBody := fmt.Sprintf(`{"TableArn": "%s"}`, tableArn1)
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListExports",
		},
		Body:   []byte(reqBody),
		Action: "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should only return exports for table1
	assert.Len(t, responseBody.ExportSummaries, 1)
	assert.Equal(t, tableArn1, *responseBody.ExportSummaries[0].ExportArn)
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Create multiple test exports
	for i := 1; i <= 5; i++ {
		exportKey := fmt.Sprintf("dynamodb:export:export%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export%d", tableName, i),
			"TableArn":     tableArn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	// Request with MaxResults
	reqBody := `{"MaxResults": 2}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListExports",
		},
		Body:   []byte(reqBody),
		Action: "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListExportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should return only 2 results
	assert.Len(t, responseBody.ExportSummaries, 2)
	// Should have NextToken for more results
	assert.NotNil(t, responseBody.NextToken)
}
