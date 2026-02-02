package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Empty(t, output.ImportSummaryList)
	require.Nil(t, output.NextToken)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/import/12345",
		"TableArn":              tableArn1,
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/67890",
		"TableArn":     tableArn2,
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:12345", import1))
	require.NoError(t, state.Set("dynamodb:import:67890", import2))

	// List all imports
	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ImportSummaryList, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/import/12345",
		"TableArn":     tableArn1,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/67890",
		"TableArn":     tableArn2,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:12345", import1))
	require.NoError(t, state.Set("dynamodb:import:67890", import2))

	// Filter by table ARN
	input := &ListImportsInput{
		TableArn: &tableArn1,
	}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ImportSummaryList, 1)
	require.Equal(t, tableArn1, *output.ImportSummaryList[0].TableArn)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    tableArn + "/import/" + string(rune('0'+i)),
			"TableArn":     tableArn,
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
		}
		require.NoError(t, state.Set("dynamodb:import:"+string(rune('0'+i)), importData))
	}

	// Request with limit
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ImportSummaryList, 2)
	require.NotNil(t, output.NextToken)
}

func TestListImports_PaginationWithNextToken(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    tableArn + "/import/" + string(rune('0'+i)),
			"TableArn":     tableArn,
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
		}
		require.NoError(t, state.Set("dynamodb:import:"+string(rune('0'+i)), importData))
	}

	// Request first page
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ImportSummaryList, 2)
	require.NotNil(t, output.NextToken)

	// Request second page
	input2 := &ListImportsInput{
		PageSize:  &pageSize,
		NextToken: output.NextToken,
	}
	resp2, err := service.listImports(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var output2 ListImportsOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)
	require.LessOrEqual(t, len(output2.ImportSummaryList), 2)
}
