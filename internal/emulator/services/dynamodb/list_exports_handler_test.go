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

	export1Key := fmt.Sprintf("dynamodb:export:%s:export1", tableName)
	export1Data := map[string]interface{}{
		"TableArn":     tableArn,
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export1", tableName),
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := fmt.Sprintf("dynamodb:export:%s:export2", tableName)
	export2Data := map[string]interface{}{
		"TableArn":     tableArn,
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", tableName),
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// Test without filter
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 exports")

	// Verify export summaries contain expected fields
	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})
		assert.Contains(t, summaryMap, "ExportArn")
		assert.Contains(t, summaryMap, "ExportStatus")
		assert.Contains(t, summaryMap, "ExportType")
	}
}

func TestListExports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create exports for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"
	tableArn1 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1)
	tableArn2 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2)

	export1Key := fmt.Sprintf("dynamodb:export:%s:export1", table1)
	export1Data := map[string]interface{}{
		"TableArn":     tableArn1,
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export1", table1),
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := fmt.Sprintf("dynamodb:export:%s:export2", table2)
	export2Data := map[string]interface{}{
		"TableArn":     tableArn2,
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export2", table2),
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err = state.Set(export2Key, export2Data)
	require.NoError(t, err)

	// Test with TableArn filter
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1, "Should have 1 export for table-1")
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	for i := 1; i <= 5; i++ {
		exportKey := fmt.Sprintf("dynamodb:export:%s:export%d", tableName, i)
		exportData := map[string]interface{}{
			"TableArn":     tableArn,
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export%d", tableName, i),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	// First page with MaxResults=2
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListExports",
		},
		Body:   []byte(`{"MaxResults": 2}`),
		Action: "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in first page")

	// Verify NextToken is present
	nextToken, hasNext := responseBody["NextToken"].(string)
	assert.True(t, hasNext, "Should have NextToken for more results")

	// Second page using NextToken
	reqBody := fmt.Sprintf(`{"MaxResults": 2, "NextToken": "%s"}`, nextToken)
	req.Body = []byte(reqBody)

	resp, err = service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in second page")
}
