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

	// Create some global table data
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}

	err := state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:global-table:test-global-table-2", globalTable2)
	require.NoError(t, err)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2)
}

func TestListGlobalTables_WithRegionFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global table data
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}

	err := state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:global-table:test-global-table-2", globalTable2)
	require.NoError(t, err)

	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 1)

	table := tables[0].(map[string]interface{})
	assert.Equal(t, "test-global-table-1", table["GlobalTableName"])
}

func TestListGlobalTables_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 0)
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set("dynamodb:global-table:test-global-table-"+string(rune(i)), globalTable)
		require.NoError(t, err)
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(tables), 2)

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := response["LastEvaluatedGlobalTableName"]
	assert.True(t, hasLastEvaluated)
}
