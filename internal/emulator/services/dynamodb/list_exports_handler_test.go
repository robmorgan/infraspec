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
	require.True(t, ok)
	assert.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	// Seed two exports
	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/export-1"
	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/export-2"

	state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
	})
	state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     tableArn,
	})

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
		sMap, ok := s.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, sMap, "ExportArn")
		assert.Contains(t, sMap, "ExportStatus")
		assert.Contains(t, sMap, "ExportType")
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2"

	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/export/exp-a"
	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/export/exp-b"

	state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn1,
	})
	state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn2,
	})

	input := &ListExportsInput{
		TableArn: &tableArn1,
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

	sMap := summaries[0].(map[string]interface{})
	assert.Equal(t, export1Arn, sMap["ExportArn"])
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table"

	// Seed 4 exports
	for i := 1; i <= 4; i++ {
		arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table/export/export-%d", i)
		state.Set(fmt.Sprintf("dynamodb:export:%s", arn), map[string]interface{}{
			"ExportArn":    arn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     tableArn,
		})
	}

	limit := int32(2)
	input := &ListExportsInput{
		MaxResults: &limit,
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

func TestListExports_NoExportsForTable(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Seed an export for a different table
	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/export/exp-1"
	state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
	})

	targetArn := "arn:aws:dynamodb:us-east-1:000000000000:table/empty-table"
	input := &ListExportsInput{
		TableArn: &targetArn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, summaries)
}
