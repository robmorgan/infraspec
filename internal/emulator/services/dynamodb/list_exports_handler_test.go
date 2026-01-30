package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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
	now := time.Now().Unix()

	export1Key := "dynamodb:export:export-id-1"
	export1Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export-id-1", tableName),
		"ExportStatus": "COMPLETED",
		"ExportTime":   float64(now),
		"TableArn":     tableArn,
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export-id-2"
	export2Data := map[string]interface{}{
		"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/export/export-id-2", tableName),
		"ExportStatus": "COMPLETED",
		"ExportTime":   float64(now + 100),
		"TableArn":     tableArn,
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
	assert.Contains(t, responseBody, "ExportSummaries")
	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have 2 exports")
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create exports for different tables
	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2"
	now := time.Now().Unix()

	export1Key := "dynamodb:export:export-1"
	export1Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/export/export-1",
		"ExportStatus": "COMPLETED",
		"ExportTime":   float64(now),
		"TableArn":     table1Arn,
	}
	err := state.Set(export1Key, export1Data)
	require.NoError(t, err)

	export2Key := "dynamodb:export:export-2"
	export2Data := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/export/export-2",
		"ExportStatus": "COMPLETED",
		"ExportTime":   float64(now),
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1, "Should have 1 export for table-1")

	// Verify it's the correct table
	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, table1Arn, summary["TableArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	now := time.Now().Unix()

	for i := 1; i <= 5; i++ {
		exportKey := fmt.Sprintf("dynamodb:export:export-%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export-%d", i),
			"ExportStatus": "COMPLETED",
			"ExportTime":   float64(now + int64(i*100)),
			"TableArn":     tableArn,
		}
		err := state.Set(exportKey, exportData)
		require.NoError(t, err)
	}

	// List with pagination (2 results per page)
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
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 exports in first page")

	// Verify NextToken is present
	assert.Contains(t, responseBody, "NextToken")
	nextToken, ok := responseBody["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Fetch next page
	input.NextToken = &nextToken
	resp, err = service.listExports(context.Background(), input)
	require.NoError(t, err)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 exports in second page")
}
