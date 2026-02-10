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

func TestListImports_Success(t *testing.T) {
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

	// Create some imports
	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef01",
		"TableArn":              tableArn,
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "DYNAMODB_JSON",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
		"EndTime":               float64(1234567890),
		"StartTime":             float64(1234567800),
	}
	require.NoError(t, state.Set("dynamodb:import:import-1", import1))

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef02",
		"TableArn":     tableArn,
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "CSV",
		"StartTime":    float64(1234567900),
	}
	require.NoError(t, state.Set("dynamodb:import:import-2", import2))

	// Test listing all imports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Verify first summary
	summary1 := summaries[0].(map[string]interface{})
	require.Contains(t, summary1["ImportArn"], "import-1")
	require.Equal(t, "COMPLETED", summary1["ImportStatus"])
	require.Equal(t, "DYNAMODB_JSON", summary1["InputFormat"])
	require.NotNil(t, summary1["CloudWatchLogGroupArn"])
	require.NotNil(t, summary1["EndTime"])

	// Verify second summary
	summary2 := summaries[1].(map[string]interface{})
	require.Contains(t, summary2["ImportArn"], "import-2")
	require.Equal(t, "IN_PROGRESS", summary2["ImportStatus"])
	require.Equal(t, "CSV", summary2["InputFormat"])
}

func TestListImports_FilterByTableArn(t *testing.T) {
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

	// Create imports for both tables
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/import/01234567890123-abcdef01",
		"TableArn":     table1Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
	}
	require.NoError(t, state.Set("dynamodb:import:import-1", import1))

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/import/01234567890123-abcdef02",
		"TableArn":     table2Arn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	require.NoError(t, state.Set("dynamodb:import:import-2", import2))

	// Test listing imports filtered by table ARN
	input := &ListImportsInput{
		TableArn: &table1Arn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	// Verify it's the correct import
	summary := summaries[0].(map[string]interface{})
	require.Contains(t, summary["ImportArn"], "table-1")
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListImports_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "not found")
}

func TestListImports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	for i := 1; i <= 5; i++ {
		importDesc := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/0123456789012" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"InputFormat":  "DYNAMODB_JSON",
		}
		key := fmt.Sprintf("dynamodb:import:import-%d", i)
		require.NoError(t, state.Set(key, importDesc))
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.LessOrEqual(t, len(summaries), 2)
}
