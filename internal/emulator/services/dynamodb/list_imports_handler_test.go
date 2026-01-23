package dynamodb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListImports_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test imports
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "CSV",
		"StartTime":   time.Now().Unix(),
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-ghijkl34",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "another-bucket",
		},
		"InputFormat": "DYNAMODB_JSON",
		"StartTime":   time.Now().Unix(),
	}

	state.Set("dynamodb:import:01234567890123-abcdef12", import1)
	state.Set("dynamodb:import:01234567890123-ghijkl34", import2)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test imports
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

	state.Set("dynamodb:import:01234567890123-abcdef12", import1)
	state.Set("dynamodb:import:01234567890123-ghijkl34", import2)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Contains(t, summary["ImportArn"], "test-table")
	assert.Equal(t, "COMPLETED", summary["ImportStatus"])
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListImports_WithPageSize(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test imports
	for i := 0; i < 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/" + string(rune(i)),
			"ImportStatus": "COMPLETED",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
			"InputFormat":  "CSV",
		}
		state.Set("dynamodb:import:"+string(rune(i)), importData)
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := responseData["NextToken"]
	assert.True(t, hasNextToken)
}
