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

	// Create test data - imports
	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-a1b2c3d4",
		"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "CSV",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-import-bucket",
		},
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/import/01234567890124-a1b2c3d5",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:import1", import1))
	require.NoError(t, state.Set("dynamodb:import:import2", import2))

	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-a1b2c3d4",
		"TableArn":     tableArn,
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/other-table/import/01234567890124-a1b2c3d5",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/other-table",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "DYNAMODB_JSON",
	}

	require.NoError(t, state.Set("dynamodb:import:import1", import1))
	require.NoError(t, state.Set("dynamodb:import:import2", import2))

	input := &ListImportsInput{
		TableArn: &tableArn,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	// Verify it's the correct import
	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ImportArn"], "test-table")
}

func TestListImports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	for i := 0; i < 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-a1b2c3d" + string(rune('0'+i)),
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"ImportStatus": "COMPLETED",
			"InputFormat":  "CSV",
		}
		require.NoError(t, state.Set("dynamodb:import:import"+string(rune('0'+i)), importData))
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}
	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := output["NextToken"]
	assert.True(t, hasNextToken)
}
