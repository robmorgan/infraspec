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

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global table entries in state
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	}

	err := state.Set("dynamodb:globaltable:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:globaltable:test-global-table-2", globalTable2)
	require.NoError(t, err)

	// Test listing all global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	globalTables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global table entries in state
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	}

	err := state.Set("dynamodb:globaltable:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:globaltable:test-global-table-2", globalTable2)
	require.NoError(t, err)

	// Test filtering by region
	input := &ListGlobalTablesInput{
		RegionName: strPtr("us-east-1"),
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	globalTables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 1)

	globalTable := globalTables[0].(map[string]interface{})
	assert.Equal(t, "test-global-table-1", globalTable["GlobalTableName"])
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global table entries
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": fmt.Sprintf("test-global-table-%d", i),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		err := state.Set(fmt.Sprintf("dynamodb:globaltable:test-global-table-%d", i), globalTable)
		require.NoError(t, err)
	}

	// Test first page
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	globalTables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 2)

	// Should have LastEvaluatedGlobalTableName
	lastEvaluatedName, ok := result["LastEvaluatedGlobalTableName"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, lastEvaluatedName)

	// Test second page
	input2 := &ListGlobalTablesInput{
		Limit:                         &limit,
		ExclusiveStartGlobalTableName: &lastEvaluatedName,
	}
	resp2, err := service.listGlobalTables(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)

	globalTables2, ok := result2["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables2, 2)
}

func TestListGlobalTables_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	globalTables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, globalTables)
}
