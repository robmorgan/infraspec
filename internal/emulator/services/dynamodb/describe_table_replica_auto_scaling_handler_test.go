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

func TestDescribeTableReplicaAutoScaling_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-global-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Test with no auto scaling configuration
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.DescribeTableReplicaAutoScaling",
		},
		Body:   []byte(fmt.Sprintf(`{"TableName":"%s"}`, tableName)),
		Action: "DescribeTableReplicaAutoScaling",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "TableAutoScalingDescription")
	autoScalingDesc, ok := responseBody["TableAutoScalingDescription"].(map[string]interface{})
	require.True(t, ok, "TableAutoScalingDescription should be an object")

	assert.Equal(t, tableName, autoScalingDesc["TableName"])
	assert.Contains(t, autoScalingDesc, "TableArn")
	assert.Contains(t, autoScalingDesc, "Replicas")
}

func TestDescribeTableReplicaAutoScaling_WithConfiguration(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table-with-autoscaling"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Add auto scaling configuration
	autoScalingKey := fmt.Sprintf("dynamodb:auto-scaling:%s", tableName)
	autoScalingConfig := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"Replicas": []interface{}{
			map[string]interface{}{
				"RegionName":    "us-east-1",
				"ReplicaStatus": "ACTIVE",
			},
		},
	}
	err = state.Set(autoScalingKey, autoScalingConfig)
	require.NoError(t, err)

	// Test with auto scaling configuration
	input := &DescribeTableReplicaAutoScalingInput{
		TableName: &tableName,
	}

	resp, err := service.describeTableReplicaAutoScaling(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	autoScalingDesc, ok := responseBody["TableAutoScalingDescription"].(map[string]interface{})
	require.True(t, ok, "TableAutoScalingDescription should be an object")

	assert.Equal(t, tableName, autoScalingDesc["TableName"])
	replicas, ok := autoScalingDesc["Replicas"].([]interface{})
	require.True(t, ok, "Replicas should be an array")
	require.Len(t, replicas, 1, "Should have one replica")
}

func TestDescribeTableReplicaAutoScaling_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "non-existent-table"
	input := &DescribeTableReplicaAutoScalingInput{
		TableName: &tableName,
	}

	resp, err := service.describeTableReplicaAutoScaling(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ResourceNotFoundException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "not found")
}

func TestDescribeTableReplicaAutoScaling_MissingTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	emptyTableName := ""
	input := &DescribeTableReplicaAutoScalingInput{
		TableName: &emptyTableName,
	}

	resp, err := service.describeTableReplicaAutoScaling(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "TableName is required")
}
