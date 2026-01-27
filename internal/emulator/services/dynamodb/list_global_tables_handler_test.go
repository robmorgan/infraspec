package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global table entries
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

	require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-2", globalTable2))

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global table entries
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

	require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-2", globalTable2))

	input := &ListGlobalTablesInput{
		RegionName: strPtr("us-east-1"),
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 1)

	table := tables[0].(map[string]interface{})
	require.Equal(t, "test-global-table-1", table["GlobalTableName"])
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create several global table entries
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-"+string(rune('0'+i)), globalTable))
	}

	// Request with Limit = 2
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 2)

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := response["LastEvaluatedGlobalTableName"]
	require.True(t, hasLastEvaluated)
}

func TestListGlobalTables_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 0)
}

func TestListGlobalTables_WithExclusiveStartGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create several global table entries
	for i := 1; i <= 3; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		require.NoError(t, state.Set("dynamodb:globaltable:test-global-table-"+string(rune('0'+i)), globalTable))
	}

	// Request starting after the first table
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: strPtr("test-global-table-1"),
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	// Should return remaining tables (at least 1)
	require.Greater(t, len(tables), 0)
}
