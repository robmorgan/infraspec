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

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ExportSummaries")
	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Create two exports in state
	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef01"
	export1Data := map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
	}
	err := state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), export1Data)
	require.NoError(t, err)

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef02"
	export2Data := map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     tableArn,
	}
	err = state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), export2Data)
	require.NoError(t, err)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, summaryMap, "ExportArn")
		assert.Contains(t, summaryMap, "ExportStatus")
		assert.Equal(t, "COMPLETED", summaryMap["ExportStatus"])
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2"

	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1/export/01234567890123-aaa"
	export1Data := map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table1Arn,
	}
	err := state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), export1Data)
	require.NoError(t, err)

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table2/export/01234567890123-bbb"
	export2Data := map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table2Arn,
	}
	err = state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), export2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
	input := &ListExportsInput{
		TableArn: &table1Arn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, export1Arn, summaryMap["ExportArn"])
}

func TestListExports_WithMaxResults(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Create 5 exports
	for i := 1; i <= 5; i++ {
		exportArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/0123456789-export%d", i)
		exportData := map[string]interface{}{
			"ExportArn":    exportArn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     tableArn,
		}
		err := state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn), exportData)
		require.NoError(t, err)
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
	assert.Contains(t, responseBody, "NextToken")
}

func TestListExports_ViaHandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListExports",
		},
		Body:   []byte("{}"),
		Action: "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Contains(t, responseBody, "ExportSummaries")
}
