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

func TestListImports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some import entries in state
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": time.Now().Unix(),
		"EndTime":   time.Now().Unix(),
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-87654321",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"InputFormat":  "DYNAMODB_JSON",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "another-bucket",
		},
		"StartTime": time.Now().Unix(),
	}

	err := state.Set("dynamodb:import:01234567890123-12345678", import1)
	require.NoError(t, err)
	err = state.Set("dynamodb:import:01234567890123-87654321", import2)
	require.NoError(t, err)

	// Test listing all imports
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some import entries in state
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"InputFormat":  "CSV",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-87654321",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"InputFormat":  "CSV",
	}

	err := state.Set("dynamodb:import:01234567890123-12345678", import1)
	require.NoError(t, err)
	err = state.Set("dynamodb:import:01234567890123-87654321", import2)
	require.NoError(t, err)

	// Test filtering by table ARN
	input := &ListImportsInput{
		TableArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"),
	}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1", summary["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple import entries
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/0123456789012-%d", i),
			"ImportStatus": "COMPLETED",
			"TableArn":     fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table-%d", i),
			"InputFormat":  "CSV",
		}
		err := state.Set(fmt.Sprintf("dynamodb:import:0123456789012-%d", i), importData)
		require.NoError(t, err)
	}

	// Test first page
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken
	nextToken, ok := result["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Test second page
	input2 := &ListImportsInput{
		PageSize:  &pageSize,
		NextToken: &nextToken,
	}
	resp2, err := service.listImports(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)

	summaries2, ok := result2["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries2, 2)
}

func TestListImports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no imports
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, summaries)
}
