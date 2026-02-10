package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  tableArn,
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Create some exports
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef01",
		"TableArn":     tableArn,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	require.NoError(t, state.Set("dynamodb:export:export-1", export1))

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef02",
		"TableArn":     tableArn,
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	require.NoError(t, state.Set("dynamodb:export:export-2", export2))

	// Test listing all exports
	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Verify first summary
	summary1 := summaries[0].(map[string]interface{})
	require.Contains(t, summary1["ExportArn"], "export-1")
	require.Equal(t, "COMPLETED", summary1["ExportStatus"])
	require.Equal(t, "FULL_EXPORT", summary1["ExportType"])

	// Verify second summary
	summary2 := summaries[1].(map[string]interface{})
	require.Contains(t, summary2["ExportArn"], "export-2")
	require.Equal(t, "IN_PROGRESS", summary2["ExportStatus"])
	require.Equal(t, "INCREMENTAL_EXPORT", summary2["ExportType"])
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test tables
	table1Name := "table-1"
	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + table1Name
	tableKey1 := "dynamodb:table:" + table1Name
	tableDesc1 := map[string]interface{}{
		"TableName": table1Name,
		"TableArn":  table1Arn,
	}
	require.NoError(t, state.Set(tableKey1, tableDesc1))

	table2Name := "table-2"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + table2Name
	tableKey2 := "dynamodb:table:" + table2Name
	tableDesc2 := map[string]interface{}{
		"TableName": table2Name,
		"TableArn":  table2Arn,
	}
	require.NoError(t, state.Set(tableKey2, tableDesc2))

	// Create exports for both tables
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/export/01234567890123-abcdef01",
		"TableArn":     table1Arn,
		"ExportStatus": "COMPLETED",
	}
	require.NoError(t, state.Set("dynamodb:export:export-1", export1))

	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/export/01234567890123-abcdef02",
		"TableArn":     table2Arn,
		"ExportStatus": "COMPLETED",
	}
	require.NoError(t, state.Set("dynamodb:export:export-2", export2))

	// Test listing exports filtered by table ARN
	input := &ListExportsInput{
		TableArn: &table1Arn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	// Verify it's the correct export
	summary := summaries[0].(map[string]interface{})
	require.Contains(t, summary["ExportArn"], "table-1")
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListExports_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "not found")
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/0123456789012" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
		}
		key := fmt.Sprintf("dynamodb:export:export-%d", i)
		require.NoError(t, state.Set(key, export))
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ExportSummaries"].([]interface{})
	require.True(t, ok)
	require.LessOrEqual(t, len(summaries), 2)
}
