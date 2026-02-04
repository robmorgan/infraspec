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

func TestDescribeImport_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a mock import in state
	importArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/import/01234567890123-abcdefgh"
	importDesc := map[string]interface{}{
		"ImportArn":          importArn,
		"ImportStatus":       "COMPLETED",
		"TableArn":           "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		"TableId":            "test-table-id",
		"StartTime":          float64(time.Now().Unix()),
		"EndTime":            float64(time.Now().Unix()),
		"ProcessedSizeBytes": 1024,
		"ProcessedItemCount": 10,
		"ImportedItemCount":  10,
		"ErrorCount":         0,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":    "my-bucket",
			"S3KeyPrefix": "imports/",
		},
		"InputFormat": "DYNAMODB_JSON",
	}

	stateKey := fmt.Sprintf("dynamodb:import:%s", importArn)
	err := state.Set(stateKey, importDesc)
	require.NoError(t, err)

	// Test DescribeImport
	input := &DescribeImportInput{
		ImportArn: strPtr(importArn),
	}

	resp, err := service.describeImport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ImportTableDescription")
	importDescResult, ok := responseBody["ImportTableDescription"].(map[string]interface{})
	require.True(t, ok, "ImportTableDescription should be an object")
	assert.Equal(t, importArn, importDescResult["ImportArn"])
	assert.Equal(t, "COMPLETED", importDescResult["ImportStatus"])
}

func TestDescribeImport_MissingImportArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with missing ImportArn
	input := &DescribeImportInput{}

	resp, err := service.describeImport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "ImportArn is required")
}

func TestDescribeImport_ImportNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with non-existent import
	input := &DescribeImportInput{
		ImportArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/NonExistent/import/12345"),
	}

	resp, err := service.describeImport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ImportNotFoundException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "Import not found")
}

func TestDescribeImport_EmptyImportArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with empty ImportArn
	input := &DescribeImportInput{
		ImportArn: strPtr(""),
	}

	resp, err := service.describeImport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}

func TestDescribeImport_InProgressStatus(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create an import in progress
	importArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/import/01234567890123-inprogress"
	importDesc := map[string]interface{}{
		"ImportArn":          importArn,
		"ImportStatus":       "IN_PROGRESS",
		"TableArn":           "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		"StartTime":          float64(time.Now().Unix()),
		"ProcessedSizeBytes": 512,
		"ProcessedItemCount": 5,
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "CSV",
	}

	stateKey := fmt.Sprintf("dynamodb:import:%s", importArn)
	err := state.Set(stateKey, importDesc)
	require.NoError(t, err)

	input := &DescribeImportInput{
		ImportArn: strPtr(importArn),
	}

	resp, err := service.describeImport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	importDescResult := responseBody["ImportTableDescription"].(map[string]interface{})
	assert.Equal(t, "IN_PROGRESS", importDescResult["ImportStatus"])
	assert.NotContains(t, importDescResult, "EndTime", "EndTime should not be present for in-progress imports")
}
