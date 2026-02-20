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

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ImportSummaryList")
	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Empty(t, summaries)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Create two imports in state
	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-aaa"
	import1Data := map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"TableArn":     tableArn,
	}
	err := state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), import1Data)
	require.NoError(t, err)

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-bbb"
	import2Data := map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
		"TableArn":     tableArn,
	}
	err = state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), import2Data)
	require.NoError(t, err)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, summaryMap, "ImportArn")
		assert.Contains(t, summaryMap, "ImportStatus")
		assert.Equal(t, "COMPLETED", summaryMap["ImportStatus"])
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1/import/01234567890123-aaa"
	import1Data := map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"TableArn":     table1Arn,
	}
	err := state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), import1Data)
	require.NoError(t, err)

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2/import/01234567890123-bbb"
	import2Data := map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"TableArn":     table2Arn,
	}
	err = state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), import2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
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
	assert.Len(t, summaries, 1)

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, import1Arn, summaryMap["ImportArn"])
}

func TestListImports_WithPageSize(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Create 5 imports
	for i := 1; i <= 5; i++ {
		importArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/0123456789-import%d", i)
		importData := map[string]interface{}{
			"ImportArn":    importArn,
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
			"TableArn":     tableArn,
		}
		err := state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), importData)
		require.NoError(t, err)
	}

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
	assert.Len(t, summaries, 2)
	assert.Contains(t, responseBody, "NextToken")
}

func TestListImports_ViaHandleRequest(t *testing.T) {
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
	assert.Contains(t, responseBody, "ImportSummaryList")
}
