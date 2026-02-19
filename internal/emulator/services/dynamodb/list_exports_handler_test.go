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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "ExportSummaries")
	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok, "ExportSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	// Create two export entries
	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/01234567890123456"
	export1 := map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"ExportTime":   float64(1000000),
	}
	state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), export1)

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/98765432109876543"
	export2 := map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"ExportTime":   float64(2000000),
	}
	state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), export2)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	for _, s := range summaries {
		m := s.(map[string]interface{})
		assert.Contains(t, m, "ExportArn")
		assert.Contains(t, m, "ExportStatus")
		assert.Contains(t, m, "TableArn")
	}
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"

	export1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/export/aaa"
	state.Set(fmt.Sprintf("dynamodb:export:%s", export1Arn), map[string]interface{}{
		"ExportArn":    export1Arn,
		"ExportStatus": "COMPLETED",
		"TableArn":     table1Arn,
	})

	export2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/export/bbb"
	state.Set(fmt.Sprintf("dynamodb:export:%s", export2Arn), map[string]interface{}{
		"ExportArn":    export2Arn,
		"ExportStatus": "COMPLETED",
		"TableArn":     table2Arn,
	})

	// Filter by table1 ARN
	input := &ListExportsInput{
		TableArn: strPtr(table1Arn),
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	m := summaries[0].(map[string]interface{})
	assert.Equal(t, export1Arn, m["ExportArn"])
	assert.Equal(t, table1Arn, m["TableArn"])
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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "ExportSummaries")
}
