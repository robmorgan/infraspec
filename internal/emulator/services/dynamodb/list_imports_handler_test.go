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

	// Create some import data in state
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
	}
	state.Set("dynamodb:import:01234567890123-abcdef12", import1)

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/01234567890124-abcdef13",
		"ImportStatus": "IN_PROGRESS",
		"InputFormat":  "CSV",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
	}
	state.Set("dynamodb:import:01234567890124-abcdef13", import2)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(summaries))
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create imports for different tables
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "DYNAMODB_JSON",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table-1",
	}
	state.Set("dynamodb:import:01234567890123-abcdef12", import1)

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/import/01234567890124-abcdef13",
		"ImportStatus": "COMPLETED",
		"InputFormat":  "CSV",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/table-2",
	}
	state.Set("dynamodb:import:01234567890124-abcdef13", import2)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 1, len(summaries))

	// Verify the returned summary is for table-1
	summary := summaries[0].(map[string]interface{})
	require.Contains(t, summary["ImportArn"], "table-1")
}

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 0, len(summaries))
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple imports
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/import/" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"InputFormat":  "DYNAMODB_JSON",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
		}
		state.Set("dynamodb:import:"+string(rune('0'+i)), importData)
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(summaries))

	// Verify NextToken is present
	_, hasNextToken := result["NextToken"]
	require.True(t, hasNextToken)
}
