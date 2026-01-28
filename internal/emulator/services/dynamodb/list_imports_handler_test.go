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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Empty(t, responseBody.ImportSummaryList, "Should have no imports initially")
	assert.Nil(t, responseBody.NextToken, "Should have no next token")
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Import 1
	import1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/01234567890123-abcdef12", tableName)
	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":    import1Arn,
		"TableArn":     tableArn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":      "my-bucket",
			"S3KeyPrefix":   "imports/",
			"S3BucketOwner": "123456789012",
		},
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	// Import 2
	import2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/98765432109876-fedcba98", tableName)
	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":    import2Arn,
		"TableArn":     tableArn,
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket-2",
		},
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// List all imports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ImportSummaryList, 2, "Should have two imports")

	// Verify summaries contain expected fields
	for _, summary := range responseBody.ImportSummaryList {
		assert.NotNil(t, summary.ImportArn)
		assert.NotEmpty(t, summary.ImportStatus)
		assert.NotEmpty(t, summary.InputFormat)
		assert.NotNil(t, summary.TableArn)
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	table1Name := "table1"
	table1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1Name)
	import1Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/01234567890123-abcdef12", table1Name)
	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":    import1Arn,
		"TableArn":     table1Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	table2Name := "table2"
	table2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2Name)
	import2Arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/98765432109876-fedcba98", table2Name)
	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":    import2Arn,
		"TableArn":     table2Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ImportSummaryList, 1, "Should have only one import for table1")
	assert.Equal(t, import1Arn, *responseBody.ImportSummaryList[0].ImportArn)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	for i := 1; i <= 5; i++ {
		importArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import%d", tableName, i)
		importKey := fmt.Sprintf("dynamodb:import:import%d", i)
		importData := map[string]interface{}{
			"ImportArn":    importArn,
			"TableArn":     tableArn,
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ImportSummaryList, 2, "Should have only 2 imports due to page size")

	// Should have NextToken for pagination
	assert.NotNil(t, responseBody.NextToken)
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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.NotNil(t, responseBody.ImportSummaryList)
}
