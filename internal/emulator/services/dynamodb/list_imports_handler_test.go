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

	import1Key := fmt.Sprintf("dynamodb:import:%s:import1", tableName)
	import1Data := map[string]interface{}{
		"TableArn":     tableArn,
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", tableName),
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":    "my-bucket",
			"S3KeyPrefix": "data/",
		},
		"StartTime": float64(now - 3600),
		"EndTime":   float64(now),
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := fmt.Sprintf("dynamodb:import:%s:import2", tableName)
	import2Data := map[string]interface{}{
		"TableArn":     tableArn,
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", tableName),
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now),
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// Test without filter
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 imports")

	// Verify import summaries contain expected fields
	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})
		assert.Contains(t, summaryMap, "ImportArn")
		assert.Contains(t, summaryMap, "ImportStatus")
		assert.Contains(t, summaryMap, "TableArn")
		assert.Contains(t, summaryMap, "InputFormat")
		assert.Contains(t, summaryMap, "S3BucketSource")
	}
}

func TestListImports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"
	tableArn1 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1)
	tableArn2 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2)
	now := time.Now().Unix()

	import1Key := fmt.Sprintf("dynamodb:import:%s:import1", table1)
	import1Data := map[string]interface{}{
		"TableArn":     tableArn1,
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", table1),
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now),
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := fmt.Sprintf("dynamodb:import:%s:import2", table2)
	import2Data := map[string]interface{}{
		"TableArn":     tableArn2,
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", table2),
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(now),
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// Test with TableArn filter
	reqBody := fmt.Sprintf(`{"TableArn": "%s"}`, tableArn1)
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListImports",
		},
		Body:   []byte(reqBody),
		Action: "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1, "Should have 1 import for table-1")
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	now := time.Now().Unix()

	for i := 1; i <= 5; i++ {
		importKey := fmt.Sprintf("dynamodb:import:%s:import%d", tableName, i)
		importData := map[string]interface{}{
			"TableArn":     tableArn,
			"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import%d", tableName, i),
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
			"S3BucketSource": map[string]interface{}{
				"S3Bucket": "my-bucket",
			},
			"StartTime": float64(now),
		}
		err := state.Set(importKey, importData)
		require.NoError(t, err)
	}

	// First page with PageSize=2
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListImports",
		},
		Body:   []byte(`{"PageSize": 2}`),
		Action: "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in first page")

	// Verify NextToken is present
	nextToken, hasNext := responseBody["NextToken"].(string)
	assert.True(t, hasNext, "Should have NextToken for more results")

	// Second page using NextToken
	reqBody := fmt.Sprintf(`{"PageSize": 2, "NextToken": "%s"}`, nextToken)
	req.Body = []byte(reqBody)

	resp, err = service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in second page")
}
