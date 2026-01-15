package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create some export data
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef12", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-ghijkl34", export2))

	input := &ListExportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ExportSummaries, 2)
}

func TestListExports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create some export data
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     tableArn,
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-ghijkl34",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef12", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-ghijkl34", export2))

	input := &ListExportsInput{
		TableArn: &tableArn,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ExportSummaries, 1)
	require.Equal(t, "COMPLETED", string(output.ExportSummaries[0].ExportStatus))
}

func TestListExports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create multiple exports
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export-" + string(rune('0'+i)),
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		}
		require.NoError(t, state.Set("dynamodb:export:export-"+string(rune('0'+i)), export))
	}

	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ExportSummaries, 2)
	require.NotNil(t, output.NextToken) // Should have more results
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListExports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ExportSummaries, 0)
	require.Nil(t, output.NextToken)
}
