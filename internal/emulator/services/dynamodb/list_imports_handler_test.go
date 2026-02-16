package dynamodb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some imports to state
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "DYNAMODB_JSON",
		"StartTime":   time.Now().Unix(),
		"EndTime":     time.Now().Unix(),
	}
	err := state.Set("dynamodb:import:01234567890123-abcdef12", import1)
	require.NoError(t, err)

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2/import/01234567890124-abcdef13",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket2",
		},
		"InputFormat": "CSV",
		"StartTime":   time.Now().Unix(),
	}
	err = state.Set("dynamodb:import:01234567890124-abcdef13", import2)
	require.NoError(t, err)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some imports to state
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/import/01234567890123-abcdef12",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
		"InputFormat":  "DYNAMODB_JSON",
	}
	err := state.Set("dynamodb:import:01234567890123-abcdef12", import1)
	require.NoError(t, err)

	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2/import/01234567890124-abcdef13",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable2",
		"InputFormat":  "CSV",
	}
	err = state.Set("dynamodb:import:01234567890124-abcdef13", import2)
	require.NoError(t, err)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	input := &ListImportsInput{
		TableArn: &tableArn,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable", summary["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add multiple imports to state
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable/import/0123456789012" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"TableArn":     "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
			"InputFormat":  "DYNAMODB_JSON",
		}
		err := state.Set("dynamodb:import:0123456789012"+string(rune('0'+i)), importData)
		require.NoError(t, err)
	}

	pageSize := int32(3)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 3)

	// Should have NextToken since we have more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}
