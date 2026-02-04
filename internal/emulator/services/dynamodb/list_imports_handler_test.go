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
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ImportSummaryList")
	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, summaries)
}

func TestListImports_WithImports(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/imported-table"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/imported-table/import/import-1"
	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/imported-table/import/import-2"

	state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn,
		"InputFormat":  "CSV",
		"S3BucketSource": map[string]interface{}{
			"S3Bucket":    "my-bucket",
			"S3KeyPrefix": "imports/",
		},
	})
	state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "IN_PROGRESS",
		"TableArn":     tableArn,
		"InputFormat":  "DYNAMODB_JSON",
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	for _, s := range summaries {
		sMap, ok := s.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, sMap, "ImportArn")
		assert.Contains(t, sMap, "ImportStatus")
		assert.Contains(t, sMap, "TableArn")
	}
}

func TestListImports_FilterByTableArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn1 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1"
	tableArn2 := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2"

	import1Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-1/import/imp-a"
	import2Arn := "arn:aws:dynamodb:us-east-1:000000000000:table/table-2/import/imp-b"

	state.Set(fmt.Sprintf("dynamodb:import:%s", import1Arn), map[string]interface{}{
		"ImportArn":    import1Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn1,
		"InputFormat":  "CSV",
	})
	state.Set(fmt.Sprintf("dynamodb:import:%s", import2Arn), map[string]interface{}{
		"ImportArn":    import2Arn,
		"ImportStatus": "COMPLETED",
		"TableArn":     tableArn2,
		"InputFormat":  "ION",
	})

	input := &ListImportsInput{
		TableArn: &tableArn1,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	sMap := summaries[0].(map[string]interface{})
	assert.Equal(t, import1Arn, sMap["ImportArn"])
	assert.Equal(t, tableArn1, sMap["TableArn"])
}

func TestListImports_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/paginated"

	// Seed 5 imports
	for i := 1; i <= 5; i++ {
		arn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/paginated/import/imp-%d", i)
		state.Set(fmt.Sprintf("dynamodb:import:%s", arn), map[string]interface{}{
			"ImportArn":    arn,
			"ImportStatus": "COMPLETED",
			"TableArn":     tableArn,
			"InputFormat":  "CSV",
		})
	}

	pageSize := int32(2)
	input := &ListImportsInput{
		PageSize: &pageSize,
	}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	assert.Contains(t, responseBody, "NextToken")
}

func TestListImports_WithS3BucketSource(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	importArn := "arn:aws:dynamodb:us-east-1:000000000000:table/s3-table/import/imp-s3"
	s3Source := map[string]interface{}{
		"S3Bucket":    "data-bucket",
		"S3KeyPrefix": "exports/2024/",
	}

	state.Set(fmt.Sprintf("dynamodb:import:%s", importArn), map[string]interface{}{
		"ImportArn":      importArn,
		"ImportStatus":   "COMPLETED",
		"TableArn":       "arn:aws:dynamodb:us-east-1:000000000000:table/s3-table",
		"InputFormat":    "CSV",
		"S3BucketSource": s3Source,
	})

	input := &ListImportsInput{}

	resp, err := service.listImports(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ImportSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	sMap := summaries[0].(map[string]interface{})
	assert.Contains(t, sMap, "S3BucketSource")
	s3SourceResp, ok := sMap["S3BucketSource"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "data-bucket", s3SourceResp["S3Bucket"])
}
