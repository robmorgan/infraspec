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

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListImports",
		},
		Body:   []byte("{}"),
		Action: "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ImportSummaryList")
	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Empty(t, summaries, "Should have no imports initially")
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	now := time.Now().Unix()

	import1Key := "dynamodb:import:import-id-1"
	import1Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import-id-1", tableName),
		"ImportStatus": "COMPLETED",
		"StartTime":    float64(now),
		"EndTime":      float64(now + 3600),
		"TableArn":     tableArn,
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import-id-2"
	import2Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import-id-2", tableName),
		"ImportStatus": "COMPLETED",
		"StartTime":    float64(now + 100),
		"EndTime":      float64(now + 3700),
		"TableArn":     tableArn,
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// List all imports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ImportSummaryList")
	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Len(t, summaries, 2, "Should have 2 imports")
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2"
	now := time.Now().Unix()

	import1Key := "dynamodb:import:import-1"
	import1Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/import/import-1",
		"ImportStatus": "COMPLETED",
		"StartTime":    float64(now),
		"EndTime":      float64(now + 3600),
		"TableArn":     table1Arn,
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import-2"
	import2Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/import/import-2",
		"ImportStatus": "COMPLETED",
		"StartTime":    float64(now),
		"EndTime":      float64(now + 3600),
		"TableArn":     table2Arn,
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// List imports for table1 only
	input := &ListImportsInput{
		TableArn: &table1Arn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1, "Should have 1 import for table-1")

	// Verify it's the correct table
	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, table1Arn, summary["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	now := time.Now().Unix()

	for i := 1; i <= 5; i++ {
		importKey := fmt.Sprintf("dynamodb:import:import-%d", i)
		importData := map[string]interface{}{
			"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import-%d", i),
			"ImportStatus": "COMPLETED",
			"StartTime":    float64(now + int64(i*100)),
			"EndTime":      float64(now + int64(i*100) + 3600),
			"TableArn":     tableArn,
		}
		err := state.Set(importKey, importData)
		require.NoError(t, err)
	}

	// List with pagination (2 results per page)
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 imports in first page")

	// Verify NextToken is present
	assert.Contains(t, responseBody, "NextToken")
	nextToken, ok := responseBody["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Fetch next page
	input.NextToken = &nextToken
	resp, err = service.listImports(context.Background(), input)
	require.NoError(t, err)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 imports in second page")
}
