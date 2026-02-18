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

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, summaries)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	importArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/01ABC"
	importArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/02DEF"

	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn1), map[string]interface{}{
		"ImportArn":    importArn1,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
	})
	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn2), map[string]interface{}{
		"ImportArn":    importArn2,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two"
	importArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-one/import/01ABC"
	importArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-two/import/02DEF"

	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn1), map[string]interface{}{
		"ImportArn":    importArn1,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn1,
	})
	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn2), map[string]interface{}{
		"ImportArn":    importArn2,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn2,
	})

	input := &ListImportsInput{
		TableArn: strPtr(tableArn1),
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, importArn1, summary["ImportArn"])
	assert.Equal(t, tableArn1, summary["TableArn"])
}

func TestListImports_ImportSummaryFields(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	importArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/01ABC"

	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), map[string]interface{}{
		"ImportArn":    importArn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"InputFormat":  "DYNAMODB_JSON",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket": "my-bucket",
		},
		"StartTime": float64(1700000000),
		"EndTime":   float64(1700001000),
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, importArn, summary["ImportArn"])
	assert.Equal(t, "COMPLETED", summary["ImportStatus"])
	assert.Equal(t, tableArn, summary["TableArn"])
	assert.Equal(t, "DYNAMODB_JSON", summary["InputFormat"])
	assert.Contains(t, summary, "S3BucketSource")
	assert.Equal(t, float64(1700000000), summary["StartTime"])
	assert.Equal(t, float64(1700001000), summary["EndTime"])
}

func TestListImports_PageSize(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	// Create 5 imports
	for i := 1; i <= 5; i++ {
		importArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/0%d", i)
		state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), map[string]interface{}{
			"ImportArn":    importArn,
			"ImportStatus": "COMPLETED",
			"TableArn":     tableArn,
		})
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ImportSummaryList"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 2)

	// NextToken should be present since there are more results
	assert.Contains(t, result, "NextToken")
}

func TestListImports_NoNextTokenWhenAllReturned(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	importArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/import/01ABC"

	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), map[string]interface{}{
		"ImportArn":    importArn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// NextToken should NOT be present when all results are returned
	assert.NotContains(t, result, "NextToken")
}
