package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeGlobalTable_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a mock global table in state
	globalTableName := "TestGlobalTable"
	globalTableDesc := map[string]interface{}{
		"GlobalTableName":   globalTableName,
		"GlobalTableStatus": "ACTIVE",
		"CreationDateTime":  float64(time.Now().Unix()),
		"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", globalTableName),
		"ReplicationGroup": []map[string]interface{}{
			{
				"RegionName": "us-east-1",
			},
			{
				"RegionName": "us-west-2",
			},
		},
	}

	stateKey := fmt.Sprintf("dynamodb:globaltable:%s", globalTableName)
	err := state.Set(stateKey, globalTableDesc)
	require.NoError(t, err)

	// Test DescribeGlobalTable
	input := &DescribeGlobalTableInput{
		GlobalTableName: strPtr(globalTableName),
	}

	resp, err := service.describeGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "GlobalTableDescription")
	globalTableDescResult, ok := responseBody["GlobalTableDescription"].(map[string]interface{})
	require.True(t, ok, "GlobalTableDescription should be an object")
	assert.Equal(t, globalTableName, globalTableDescResult["GlobalTableName"])
	assert.Equal(t, "ACTIVE", globalTableDescResult["GlobalTableStatus"])
}

func TestDescribeGlobalTable_MissingGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with missing GlobalTableName
	input := &DescribeGlobalTableInput{}

	resp, err := service.describeGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "GlobalTableName is required")
}

func TestDescribeGlobalTable_GlobalTableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with non-existent global table
	input := &DescribeGlobalTableInput{
		GlobalTableName: strPtr("NonExistentGlobalTable"),
	}

	resp, err := service.describeGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "GlobalTableNotFoundException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "Global table not found")
}

func TestDescribeGlobalTable_EmptyGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with empty GlobalTableName
	input := &DescribeGlobalTableInput{
		GlobalTableName: strPtr(""),
	}

	resp, err := service.describeGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}
