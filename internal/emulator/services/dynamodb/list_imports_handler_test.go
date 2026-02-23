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

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import-001"
	err := state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"InputFormat":  "CSV",
	})
	require.NoError(t, err)

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import-002"
	err = state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "FAILED",
		"TableArn":     tableArn,
		"InputFormat":  "DYNAMODB_JSON",
	})
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

	for _, s := range summaries {
		sm, ok := s.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, sm, "ImportArn")
		assert.Contains(t, sm, "ImportStatus")
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/import/001"
	err := state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     table1Arn,
		"InputFormat":  "CSV",
	})
	require.NoError(t, err)

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/import/002"
	err = state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     table2Arn,
		"InputFormat":  "CSV",
	})
	require.NoError(t, err)

	// Filter by table1 only
	input := &ListImportsInput{
		TableArn: strPtr(table1Arn),
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	sm, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, import1Arn, sm["ImportArn"])
	assert.Equal(t, table1Arn, sm["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table"

	for i := 1; i <= 5; i++ {
		importArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/paginated-table/import/import-%03d", i)
		err := state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), map[string]interface{}{
			"ImportArn":    importArn,
			"ImportStatus": "COMPLETED",
			"TableArn":     tableArn,
			"InputFormat":  "CSV",
		})
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
