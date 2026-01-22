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

	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", tableName),
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now),
		"EndTime":   float64(now + 100),
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", tableName),
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now + 200),
		"EndTime":   float64(now + 300),
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
	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Len(t, summaries, 2, "Should have two imports")

	// Verify import summaries contain expected fields
	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok, "Each summary should be an object")
		assert.Contains(t, summaryMap, "ImportArn")
		assert.Contains(t, summaryMap, "ImportStatus")
		assert.Contains(t, summaryMap, "TableArn")
		assert.Equal(t, tableArn, summaryMap["TableArn"])
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	table1Name := "table1"
	table1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1Name)
	now := time.Now().Unix()

	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", table1Name),
		"ImportStatus": "COMPLETED",
		"TableArn":     table1Arn,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now),
		"EndTime":   float64(now + 100),
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	table2Name := "table2"
	table2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2Name)

	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", table2Name),
		"ImportStatus": "COMPLETED",
		"TableArn":     table2Arn,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now + 200),
		"EndTime":   float64(now + 300),
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
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Len(t, summaries, 1, "Should have only one import for table1")

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Arn, summaryMap["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableName := "test-table-paginated"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	now := time.Now().Unix()

	for i := 1; i <= 5; i++ {
		importKey := fmt.Sprintf("dynamodb:import:import%d", i)
		importData := map[string]interface{}{
			"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import%d", tableName, i),
			"ImportStatus": "COMPLETED",
			"TableArn":     tableArn,
			"S3BucketSource": map[string]interface{}{
				"S3Bucket": "my-bucket",
			},
			"StartTime": float64(now + int64(i*100)),
			"EndTime":   float64(now + int64(i*100+50)),
		}
		err := state.Set(importKey, importData)
		require.NoError(t, err)
	}

	// List imports with page size
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
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Len(t, summaries, 2, "Should have only 2 imports due to page size")

	// Should have NextToken for pagination
	assert.Contains(t, responseBody, "NextToken")
}

func TestListImports_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListImports should work with no parameters
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ImportSummaryList")
}
