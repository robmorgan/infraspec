package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListImports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test imports
	import1Key := "dynamodb:import:import-1"
	import1Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "CSV",
		"StartTime":   float64(1234567890),
		"EndTime":     float64(1234567900),
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import-2"
	import2Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890124-12345679",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "DYNAMODB_JSON",
		"StartTime":   float64(1234567890),
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// Test ListImports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(summaries))
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	import1Key := "dynamodb:import:import-1"
	import1Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table1/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table1",
	}
	err := state.Set(import1Key, import1Data)
	require.NoError(t, err)

	import2Key := "dynamodb:import:import-2"
	import2Data := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table2/import/01234567890124-12345679",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table2",
	}
	err = state.Set(import2Key, import2Data)
	require.NoError(t, err)

	// Filter by table1 ARN
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/table1"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, len(summaries))

	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ImportArn"], "table1")
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test without any imports
	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListImports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	for i := 1; i <= 5; i++ {
		importKey := "dynamodb:import:import-" + string(rune('0'+i))
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test/import/" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test",
		}
		err := state.Set(importKey, importData)
		require.NoError(t, err)
	}

	pageSize := int32(3)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 3)

	// Should have NextToken since we have more results
	if len(summaries) == 3 {
		_, hasNextToken := response["NextToken"]
		assert.True(t, hasNextToken)
	}
}
