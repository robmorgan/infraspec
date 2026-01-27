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

	// Create some import entries
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":    "my-bucket",
			"S3KeyPrefix": "imports/",
		},
		"InputFormat": "DYNAMODB_JSON",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2/import/01234567890123-87654321",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket-2",
		},
		"InputFormat": "ION",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-12345678", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-87654321", import2))

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
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

	// Create some import entries
	import1 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"InputFormat": "DYNAMODB_JSON",
	}
	import2 := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2/import/01234567890123-87654321",
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table-2",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket-2",
		},
		"InputFormat": "ION",
	}

	require.NoError(t, state.Set("dynamodb:import:01234567890123-12345678", import1))
	require.NoError(t, state.Set("dynamodb:import:01234567890123-87654321", import2))

	input := &ListImportsInput{
		TableArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/01234567890123-12345678", summary["ImportArn"])
}

func TestListImports_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create several import entries
	for i := 1; i <= 5; i++ {
		importData := map[string]interface{}{
			"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/0123456789012" + string(rune('0'+i)),
			"ImportStatus": "COMPLETED",
			"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
			"S3BucketSource": map[string]interface{}{
				"S3Bucket": "my-bucket",
			},
			"InputFormat": "DYNAMODB_JSON",
		}
		require.NoError(t, state.Set("dynamodb:import:0123456789012"+string(rune('0'+i)), importData))
	}

	// Request with PageSize = 2
	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}

func TestListImports_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 0)
}

func TestListImports_WithAllFields(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create import with all fields
	importData := map[string]interface{}{
		"ImportArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/01234567890123-12345678",
		"ImportStatus": "COMPLETED",
		"TableArn":     "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":    "my-bucket",
			"S3KeyPrefix": "imports/",
		},
		"CloudWatchLogGroupArn": "arn:aws:logs:us-east-1:123456789012:log-group:/aws/dynamodb/imports",
		"InputFormat":           "DYNAMODB_JSON",
		"StartTime":             1234567890.0,
		"EndTime":               1234567900.0,
	}
	require.NoError(t, state.Set("dynamodb:import:01234567890123-12345678", importData))

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/import/01234567890123-12345678", summary["ImportArn"])
	require.Equal(t, "COMPLETED", summary["ImportStatus"])
	require.NotNil(t, summary["S3BucketSource"])
	require.NotNil(t, summary["CloudWatchLogGroupArn"])
	require.Equal(t, "DYNAMODB_JSON", summary["InputFormat"])
	require.NotNil(t, summary["StartTime"])
	require.NotNil(t, summary["EndTime"])
}
