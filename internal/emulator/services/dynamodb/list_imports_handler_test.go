package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListImports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create some import data
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"InputFormat":  "CSV",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-ghijkl34",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef12", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-ghijkl34", import2))

	input := &ListImportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ImportSummaryList, 2)
}

func TestListImports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Setup: Create some import data
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"InputFormat":  "CSV",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-ghijkl34",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-abcdef12", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-ghijkl34", import2))

	input := &ListImportsInput{
		TableArn: &tableArn,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ImportSummaryList, 1)
	require.Equal(t, "COMPLETED", string(output.ImportSummaryList[0].ImportStatus))
}

func TestListImports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create multiple imports
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/import-" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"InputFormat":  "CSV",
		}
		require.NoError(t, state.Set("dynamodb:import:import-"+string(rune('0'+i)), importData))
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ImportSummaryList, 2)
	require.NotNil(t, output.NextToken) // Should have more results
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListImportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ImportSummaryList, 0)
	require.Nil(t, output.NextToken)
}
