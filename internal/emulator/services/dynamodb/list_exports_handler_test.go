package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Empty(t, output.ExportSummaries)
	require.Nil(t, output.NextToken)
}

func TestListExports_WithExports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/export/12345",
		"TableArn":     tableArn1,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/67890",
		"TableArn":     tableArn2,
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:12345", export1))
	require.NoError(t, state.Set("dynamodb:export:67890", export2))

	// List all exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ExportSummaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test exports
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/export/12345",
		"TableArn":     tableArn1,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/67890",
		"TableArn":     tableArn2,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:12345", export1))
	require.NoError(t, state.Set("dynamodb:export:67890", export2))

	// Filter by table ARN
	input := &ListExportsInput{
		TableArn: &tableArn1,
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ExportSummaries, 1)
	require.Equal(t, tableArn1, *output.ExportSummaries[0].ExportArn)
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple exports
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    tableArn + "/export/" + string(rune('0'+i)),
			"TableArn":     tableArn,
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		require.NoError(t, state.Set("dynamodb:export:"+string(rune('0'+i)), export))
	}

	// Request with limit
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ExportSummaries, 2)
	require.NotNil(t, output.NextToken)
}
