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
	assert.NotNil(t, responseBody.ImportSummaryList)
	assert.Empty(t, responseBody.ImportSummaryList, "Should have no imports initially")
	assert.Nil(t, responseBody.NextToken)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":             fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", tableName),
		"TableArn":              tableArn,
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "DYNAMODB_JSON",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":             fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", tableName),
		"TableArn":              tableArn,
		"ImportStatus":          "IN_PROGRESS",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ImportSummaryList, 2)

	// Check first import
	assert.NotNil(t, responseBody.ImportSummaryList[0].ImportArn)
	assert.Equal(t, ImportStatus("COMPLETED"), responseBody.ImportSummaryList[0].ImportStatus)
	assert.Equal(t, InputFormat("DYNAMODB_JSON"), responseBody.ImportSummaryList[0].InputFormat)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"
	tableArn1 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1)
	tableArn2 := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2)

	import1Key := "dynamodb:import:import1"
	import1Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import1", table1),
		"TableArn":     tableArn1,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import2"
	import2Data := map[string]interface{}{
		"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import2", table2),
		"TableArn":     tableArn2,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should only return imports for table1
	assert.Len(t, responseBody.ImportSummaryList, 1)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Create multiple test imports
	for i := 1; i <= 5; i++ {
		importKey := fmt.Sprintf("dynamodb:import:import%d", i)
		importData := map[string]interface{}{
			"ImportArn":    fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/import/import%d", tableName, i),
			"TableArn":     tableArn,
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
		}
		err := state.Set(importKey, importData)
		require.NoError(t, err)
	}

	// Request with PageSize
	reqBody := `{"PageSize": 2}`
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

	var responseBody ListImportsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should return only 2 results
	assert.Len(t, responseBody.ImportSummaryList, 2)
	// Should have NextToken for more results
	assert.NotNil(t, responseBody.NextToken)
}
