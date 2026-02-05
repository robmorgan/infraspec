package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func TestDescribeExport_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a mock export in state
	exportArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/export/01234567890123-abcdefgh"
	exportDesc := map[string]interface{}{
		"ExportArn":       exportArn,
		"ExportStatus":    "COMPLETED",
		"StartTime":       float64(time.Now().Unix()),
		"EndTime":         float64(time.Now().Unix()),
		"ExportManifest":  "s3://my-bucket/export-manifest.json",
		"TableArn":        "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		"TableId":         "test-table-id",
		"ExportTime":      float64(time.Now().Unix()),
		"S3Bucket":        "my-bucket",
		"S3Prefix":        "exports/",
		"ExportFormat":    "DYNAMODB_JSON",
		"BilledSizeBytes": 1024,
		"ItemCount":       10,
	}

	stateKey := fmt.Sprintf("dynamodb:export:%s", exportArn)
	err := state.Set(stateKey, exportDesc)
	require.NoError(t, err)

	// Test DescribeExport
	input := &DescribeExportInput{
		ExportArn: strPtr(exportArn),
	}

	resp, err := service.describeExport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ExportDescription")
	exportDescResult, ok := responseBody["ExportDescription"].(map[string]interface{})
	require.True(t, ok, "ExportDescription should be an object")
	assert.Equal(t, exportArn, exportDescResult["ExportArn"])
	assert.Equal(t, "COMPLETED", exportDescResult["ExportStatus"])
}

func TestDescribeExport_MissingExportArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with missing ExportArn
	input := &DescribeExportInput{}

	resp, err := service.describeExport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "ExportArn is required")
}

func TestDescribeExport_ExportNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with non-existent export
	input := &DescribeExportInput{
		ExportArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/NonExistent/export/12345"),
	}

	resp, err := service.describeExport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ExportNotFoundException", responseBody["__type"])
}

func TestDescribeExport_EmptyExportArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with empty ExportArn
	input := &DescribeExportInput{
		ExportArn: strPtr(""),
	}

	resp, err := service.describeExport(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}
