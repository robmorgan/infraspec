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

	// Create some import data
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abc",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "CSV",
		"StartTime":   1234567890.0,
		"EndTime":     1234567900.0,
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-def",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket-2",
		},
		"InputFormat": "DYNAMODB_JSON",
		"StartTime":   1234567890.0,
	}

	err := state.Set("dynamodb:import:01234567890123-abc", import1)
	require.NoError(t, err)
	err = state.Set("dynamodb:import:01234567890123-def", import2)
	require.NoError(t, err)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListImports_WithTableArnFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some import data
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abc",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-def",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
	}

	err := state.Set("dynamodb:import:01234567890123-abc", import1)
	require.NoError(t, err)
	err = state.Set("dynamodb:import:01234567890123-def", import2)
	require.NoError(t, err)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abc", summary["ImportArn"])
}

func TestListImports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListImports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	for i := 1; i <= 30; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/" + string(rune(i)),
			"ImportStatus": "COMPLETED",
		}
		err := state.Set("dynamodb:import:"+string(rune(i)), importData)
		require.NoError(t, err)
	}

	pageSize := int32(10)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 10)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	assert.True(t, hasNextToken)
}
