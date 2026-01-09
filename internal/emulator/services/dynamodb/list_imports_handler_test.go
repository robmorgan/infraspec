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
	import1 := map[string]interface{}{
		"ImportArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus":          "COMPLETED",
		"InputFormat":           "DYNAMODB_JSON",
		"TableArn":              "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:000000000000:log-group:/aws/dynamodb/imports",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":      "my-bucket",
			"S3KeyPrefix":   "imports/",
			"S3BucketOwner": "000000000000",
		},
	}

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-ghijkl34",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "CSV",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket-2",
		},
	}

	require.NoError(t, state.Set("dynamodb:import:test-import-1", import1))
	require.NoError(t, state.Set("dynamodb:import:test-import-2", import2))

	// Test without filter
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(summaries))
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2"

	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
		"TableArn":     tableArn1,
	}

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-2/import/01234567890123-ghijkl34",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"TableArn":     tableArn2,
	}

	require.NoError(t, state.Set("dynamodb:import:test-import-1", import1))
	require.NoError(t, state.Set("dynamodb:import:test-import-2", import2))

	// Test with table ARN filter
	input := &ListImportsInput{
		TableArn: &tableArn1,
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(summaries))

	// Verify it's the correct import
	firstSummary := summaries[0].(map[string]interface{})
	assert.Contains(t, firstSummary["ImportArn"], "test-table")
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	for i := 0; i < 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/" + string(rune('a'+i)),
			"ImportStatus": "COMPLETED",
			"InputFormat":  "DYNAMODB_JSON",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		}
		require.NoError(t, state.Set("dynamodb:import:test-import-"+string(rune('a'+i)), importData))
	}

	// Test with page size
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(summaries))

	// Should have NextToken since there are more results
	_, hasNextToken := result["NextToken"]
	assert.True(t, hasNextToken)
}
