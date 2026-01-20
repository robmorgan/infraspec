package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/testing/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - exports
	exportData1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set("dynamodb:export:export1", exportData1)
	require.NoError(t, err)

	exportData2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890124-abcdef13",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set("dynamodb:export:export2", exportData2)
	require.NoError(t, err)

	exportData3 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/export/01234567890125-abcdef14",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err = state.Set("dynamodb:export:export3", exportData3)
	require.NoError(t, err)

	// Test listing all exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 3)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	exportData1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890123-abcdef12",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err := state.Set("dynamodb:export:export1", exportData1)
	require.NoError(t, err)

	exportData2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/01234567890124-abcdef13",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}
	err = state.Set("dynamodb:export:export2", exportData2)
	require.NoError(t, err)

	exportData3 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/export/01234567890125-abcdef14",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	err = state.Set("dynamodb:export:export3", exportData3)
	require.NoError(t, err)

	// Test filtering by table ARN
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListExportsInput{
		TableArn: &tableArn,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 2)
}

func TestListExports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Empty(t, output.ExportSummaries)
}

func TestListExports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - 5 exports
	for i := 1; i <= 5; i++ {
		exportData := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/0123456789012" + string(rune('0'+i)),
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		err := state.Set("dynamodb:export:export"+string(rune('0'+i)), exportData)
		require.NoError(t, err)
	}

	// Request with max results of 2
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 2)
	assert.NotNil(t, output.NextToken)
}
