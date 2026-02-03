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
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	importArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import1"
	importArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import2"
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)

	import1Key := "dynamodb:import:import1"
	import1Data := ImportSummary{
		ImportArn:    &importArn1,
		ImportStatus: "COMPLETED",
		InputFormat:  "DYNAMODB_JSON",
		TableArn:     &tableArn,
		StartTime:    &startTime,
		EndTime:      &endTime,
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import2"
	import2Data := ImportSummary{
		ImportArn:    &importArn2,
		ImportStatus: "IN_PROGRESS",
		InputFormat:  "CSV",
		TableArn:     &tableArn,
		StartTime:    &startTime,
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// List all imports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, output.ImportSummaryList, 2, "Should have two imports")

	// Verify summaries contain expected fields
	for _, summary := range output.ImportSummaryList {
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
	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	importArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table1/import/import1"
	startTime := time.Now()

	import1Key := "dynamodb:import:import1"
	import1Data := ImportSummary{
		ImportArn:    &importArn1,
		ImportStatus: "COMPLETED",
		InputFormat:  "DYNAMODB_JSON",
		TableArn:     &table1Arn,
		StartTime:    &startTime,
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2"
	importArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table2/import/import2"
	import2Key := "dynamodb:import:import2"
	import2Data := ImportSummary{
		ImportArn:    &importArn2,
		ImportStatus: "COMPLETED",
		InputFormat:  "CSV",
		TableArn:     &table2Arn,
		StartTime:    &startTime,
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

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 1, "Should have only one import for table1")
	assert.Equal(t, table1Arn, *output.ImportSummaryList[0].TableArn)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	startTime := time.Now()

	for i := 1; i <= 5; i++ {
		importArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import%d", i)
		importKey := fmt.Sprintf("dynamodb:import:import%d", i)
		importData := ImportSummary{
			ImportArn:    &importArn,
			ImportStatus: "COMPLETED",
			InputFormat:  "DYNAMODB_JSON",
			TableArn:     &tableArn,
			StartTime:    &startTime,
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

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ImportSummaryList, 2, "Should have only 2 imports due to page size")

	// Should have NextToken for pagination
	assert.NotNil(t, output.NextToken, "Should have NextToken when there are more results")
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

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.ImportSummaryList)
}
