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
	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports in state
	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	export1Data := map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
	}
	err := state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), export1Data)
	require.NoError(t, err)

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/05678"
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

	for _, s := range summaries {
		sm, ok := s.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, sm, "ExportArn")
		assert.Contains(t, sm, "ExportStatus")
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"

	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/export/001"
	err := state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table1Arn,
	})
	require.NoError(t, err)

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/export/002"
	err = state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     table2Arn,
	})
	require.NoError(t, err)

	// Filter by table1 only
	input := &ListExportsInput{
		TableArn: strPtr(table1Arn),
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	sm, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, export1Arn, sm["ExportArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table"

	for i := 1; i <= 5; i++ {
		exportArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table/export/%03d", i)
		err := state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn), map[string]interface{}{
			"ExportArn":    exportArn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     tableArn,
		})
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
