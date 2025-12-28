package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeGlobalTableSettings_Success(t *testing.T) {
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
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}

	stateKey := fmt.Sprintf("dynamodb:globaltable:%s", globalTableName)
	err := state.Set(stateKey, globalTableDesc)
	require.NoError(t, err)

	// Test DescribeGlobalTableSettings
	input := &DescribeGlobalTableSettingsInput{
		GlobalTableName: strPtr(globalTableName),
	}

	resp, err := service.describeGlobalTableSettings(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, globalTableName, responseBody["GlobalTableName"])
	assert.Contains(t, responseBody, "ReplicaSettings")

	replicaSettings, ok := responseBody["ReplicaSettings"].([]interface{})
	require.True(t, ok, "ReplicaSettings should be an array")
	assert.Len(t, replicaSettings, 2, "Should have 2 replica settings")

	// Check first replica
	replica1, ok := replicaSettings[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "us-east-1", replica1["RegionName"])
	assert.Equal(t, "ACTIVE", replica1["ReplicaStatus"])
}

func TestDescribeGlobalTableSettings_MissingGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with missing GlobalTableName
	input := &DescribeGlobalTableSettingsInput{}

	resp, err := service.describeGlobalTableSettings(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "GlobalTableName is required")
}

func TestDescribeGlobalTableSettings_GlobalTableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with non-existent global table
	input := &DescribeGlobalTableSettingsInput{
		GlobalTableName: strPtr("NonExistentGlobalTable"),
	}

	resp, err := service.describeGlobalTableSettings(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "GlobalTableNotFoundException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "Global table not found")
}

func TestDescribeGlobalTableSettings_EmptyGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with empty GlobalTableName
	input := &DescribeGlobalTableSettingsInput{
		GlobalTableName: strPtr(""),
	}

	resp, err := service.describeGlobalTableSettings(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}

func TestDescribeGlobalTableSettings_NoReplicas(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a global table without replicas
	globalTableName := "TestGlobalTableNoReplicas"
	globalTableDesc := map[string]interface{}{
		"GlobalTableName":   globalTableName,
		"GlobalTableStatus": "ACTIVE",
		"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", globalTableName),
	}

	stateKey := fmt.Sprintf("dynamodb:globaltable:%s", globalTableName)
	err := state.Set(stateKey, globalTableDesc)
	require.NoError(t, err)

	input := &DescribeGlobalTableSettingsInput{
		GlobalTableName: strPtr(globalTableName),
	}

	resp, err := service.describeGlobalTableSettings(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, globalTableName, responseBody["GlobalTableName"])
}
