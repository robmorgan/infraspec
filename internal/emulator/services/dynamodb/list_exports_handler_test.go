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

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, summaries)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	exportArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/01ABC"
	exportArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/02DEF"

	// Store exports
	state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn1), map[string]interface{}{
		"ExportArn":    exportArn1,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
	})
	state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn2), map[string]interface{}{
		"ExportArn":    exportArn2,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn,
	})

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"
	exportArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/export/01ABC"
	exportArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/export/02DEF"

	state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn1), map[string]interface{}{
		"ExportArn":    exportArn1,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn1,
	})
	state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn2), map[string]interface{}{
		"ExportArn":    exportArn2,
		"ExportStatus": "COMPLETED",
		"TableArn":     tableArn2,
	})

	// Filter by table ARN 1
	input := &ListExportsInput{
		TableArn: strPtr(tableArn1),
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, exportArn1, summary["ExportArn"])
	assert.Equal(t, tableArn1, summary["TableArn"])
}

func TestListExports_ExportSummaryFields(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	exportArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/01ABC"

	state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn), map[string]interface{}{
		"ExportArn":    exportArn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
		"ExportTime":   float64(1700000000),
	})

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, exportArn, summary["ExportArn"])
	assert.Equal(t, "COMPLETED", summary["ExportStatus"])
	assert.Equal(t, "FULL_EXPORT", summary["ExportType"])
	assert.Equal(t, tableArn, summary["TableArn"])
	assert.Equal(t, float64(1700000000), summary["ExportTime"])
}

func TestListExports_MaxResults(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	// Create 5 exports
	for i := 1; i <= 5; i++ {
		exportArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/my-table/export/0%d", i)
		state.Set(fmt.Sprintf("dynamodb:export:%s", exportArn), map[string]interface{}{
			"ExportArn":    exportArn,
			"ExportStatus": "COMPLETED",
			"TableArn":     tableArn,
		})
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ExportSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 2)

	// NextToken should be present since there are more results
	assert.Contains(t, result, "NextToken")
}
