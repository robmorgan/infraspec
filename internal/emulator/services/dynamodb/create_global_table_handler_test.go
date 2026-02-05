package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func TestCreateGlobalTable_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateGlobalTableInput{
		GlobalTableName: strPtr("test-global-table"),
		ReplicationGroup: []Replica{
			{RegionName: strPtr("us-east-1")},
			{RegionName: strPtr("us-west-2")},
		},
	}

	resp, err := service.createGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	globalTableDesc, ok := result["GlobalTableDescription"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-global-table", globalTableDesc["GlobalTableName"])
	assert.Equal(t, "ACTIVE", globalTableDesc["GlobalTableStatus"])

	replicas, ok := globalTableDesc["ReplicationGroup"].([]interface{})
	require.True(t, ok)
	assert.Len(t, replicas, 2)
}

func TestCreateGlobalTable_MissingGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateGlobalTableInput{
		ReplicationGroup: []Replica{
			{RegionName: strPtr("us-east-1")},
		},
	}

	resp, err := service.createGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "GlobalTableName is required")
}

func TestCreateGlobalTable_MissingReplicationGroup(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateGlobalTableInput{
		GlobalTableName: strPtr("test-global-table"),
	}

	resp, err := service.createGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "ReplicationGroup is required")
}

func TestCreateGlobalTable_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create the global table first
	existingGlobalTable := map[string]interface{}{
		"GlobalTableName":   "test-global-table",
		"GlobalTableStatus": "ACTIVE",
	}
	state.Set("dynamodb:globaltable:test-global-table", existingGlobalTable)

	input := &CreateGlobalTableInput{
		GlobalTableName: strPtr("test-global-table"),
		ReplicationGroup: []Replica{
			{RegionName: strPtr("us-east-1")},
		},
	}

	resp, err := service.createGlobalTable(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "GlobalTableAlreadyExistsException", result["__type"])
}
