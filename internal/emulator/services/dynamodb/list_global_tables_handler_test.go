package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test global tables
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

	state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	state.Set("dynamodb:global-table:test-global-table-2", globalTable2)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables, ok := responseData["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test global tables
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

	state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	state.Set("dynamodb:global-table:test-global-table-2", globalTable2)

	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables, ok := responseData["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 1)

	globalTable := globalTables[0].(map[string]interface{})
	assert.Equal(t, "test-global-table-1", globalTable["GlobalTableName"])
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables, ok := responseData["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 0)
}

func TestListGlobalTables_WithLimit(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test global tables
	for i := 0; i < 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		state.Set("dynamodb:global-table:test-global-table-"+string(rune('0'+i)), globalTable)
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables, ok := responseData["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, globalTables, 2)

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := responseData["LastEvaluatedGlobalTableName"]
	assert.True(t, hasLastEvaluated)
}
