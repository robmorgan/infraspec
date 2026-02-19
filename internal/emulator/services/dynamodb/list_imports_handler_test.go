package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListImports_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "ImportSummaryList")
	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok, "ImportSummaryList should be an array")
	assert.Empty(t, summaries)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/aaa111"
	state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"InputFormat":  "DYNAMODB_JSON",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
	})

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/bbb222"
	state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "FAILED",
		"TableArn":     tableArn,
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	for _, s := range summaries {
		m := s.(map[string]interface{})
		assert.Contains(t, m, "ImportArn")
		assert.Contains(t, m, "ImportStatus")
		assert.Contains(t, m, "TableArn")
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	table2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/import/aaa"
	state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     table1Arn,
	})

	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/import/bbb"
	state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     table2Arn,
	})

	// Filter by table1 ARN
	input := &ListImportsInput{
		TableArn: strPtr(table1Arn),
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	m := summaries[0].(map[string]interface{})
	assert.Equal(t, import1Arn, m["ImportArn"])
	assert.Equal(t, table1Arn, m["TableArn"])
}

func TestListImports_ViaHandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListImports",
		},
		Body:   []byte("{}"),
		Action: "ListImports",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "ImportSummaryList")
}
