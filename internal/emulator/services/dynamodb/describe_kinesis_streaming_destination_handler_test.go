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

func TestDescribeKinesisStreamingDestination_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Test with no Kinesis destinations
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.DescribeKinesisStreamingDestination",
		},
		Body:   []byte(fmt.Sprintf(`{"TableName":"%s"}`, tableName)),
		Action: "DescribeKinesisStreamingDestination",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, tableName, responseBody["TableName"])
	assert.Contains(t, responseBody, "KinesisDataStreamDestinations")
	destinations, ok := responseBody["KinesisDataStreamDestinations"].([]interface{})
	require.True(t, ok, "KinesisDataStreamDestinations should be an array")
	assert.Empty(t, destinations, "Should have no destinations initially")
}

func TestDescribeKinesisStreamingDestination_WithDestinations(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table-with-kinesis"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Add Kinesis destinations
	kinesisKey := fmt.Sprintf("dynamodb:kinesis-destinations:%s", tableName)
	destinations := []interface{}{
		map[string]interface{}{
			"StreamArn":         "arn:aws:kinesis:us-east-1:000000000000:stream/test-stream",
			"DestinationStatus": "ACTIVE",
		},
	}
	err = state.Set(kinesisKey, destinations)
	require.NoError(t, err)

	// Test with Kinesis destinations
	input := &DescribeKinesisStreamingDestinationInput{
		TableName: &tableName,
	}

	resp, err := service.describeKinesisStreamingDestination(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, tableName, responseBody["TableName"])
	returnedDestinations, ok := responseBody["KinesisDataStreamDestinations"].([]interface{})
	require.True(t, ok, "KinesisDataStreamDestinations should be an array")
	require.Len(t, returnedDestinations, 1, "Should have one destination")
}

func TestDescribeKinesisStreamingDestination_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "non-existent-table"
	input := &DescribeKinesisStreamingDestinationInput{
		TableName: &tableName,
	}

	resp, err := service.describeKinesisStreamingDestination(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ResourceNotFoundException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "not found")
}

func TestDescribeKinesisStreamingDestination_MissingTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	emptyTableName := ""
	input := &DescribeKinesisStreamingDestinationInput{
		TableName: &emptyTableName,
	}

	resp, err := service.describeKinesisStreamingDestination(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "TableName is required")
}
