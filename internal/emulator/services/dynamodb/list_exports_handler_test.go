package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports_Success(t *testing.T) {
	service, state := setupTestService(t)

	// Create test export data
	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/export/01234567890123-abcdef12",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1",
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-abcdef34",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"ExportStatus": "IN_PROGRESS",
		"ExportType":   "INCREMENTAL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef12", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef34", export2))

	// Test listing all exports
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 2)
}

func TestListExports_FilterByTableArn(t *testing.T) {
	service, state := setupTestService(t)

	// Create test export data
	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	export1 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-1/export/01234567890123-abcdef12",
		"TableArn":     tableArn1,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}
	export2 := map[string]interface{}{
		"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/export/01234567890123-abcdef34",
		"TableArn":     tableArn2,
		"ExportStatus": "COMPLETED",
		"ExportType":   "FULL_EXPORT",
	}

	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef12", export1))
	require.NoError(t, state.Set("dynamodb:export:01234567890123-abcdef34", export2))

	// Test listing exports for specific table
	input := &ListExportsInput{
		TableArn: &tableArn1,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 1)
	assert.Equal(t, tableArn1, *output.ExportSummaries[0].ExportArn)
}

func TestListExports_EmptyList(t *testing.T) {
	service, _ := setupTestService(t)

	// Test listing when no exports exist
	input := &ListExportsInput{}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 0)
}

func TestListExports_Pagination(t *testing.T) {
	service, state := setupTestService(t)

	// Create multiple exports
	for i := 1; i <= 5; i++ {
		export := map[string]interface{}{
			"ExportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/export/export-" + string(rune('0'+i)),
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ExportStatus": "COMPLETED",
			"ExportType":   "FULL_EXPORT",
		}
		key := "dynamodb:export:export-" + string(rune('0'+i))
		require.NoError(t, state.Set(key, export))
	}

	// Test pagination with MaxResults
	maxResults := int32(2)
	input := &ListExportsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listExports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListExportsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ExportSummaries, 2)
	assert.NotNil(t, output.NextToken)
}
