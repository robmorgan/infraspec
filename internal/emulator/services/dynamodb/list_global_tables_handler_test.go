package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Empty(t, tables)
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some global tables to state
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "TestGlobalTable1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set("dynamodb:global-table:TestGlobalTable1", globalTable1)
	require.NoError(t, err)

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "TestGlobalTable2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set("dynamodb:global-table:TestGlobalTable2", globalTable2)
	require.NoError(t, err)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
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

	// Add some global tables to state
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "TestGlobalTable1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set("dynamodb:global-table:TestGlobalTable1", globalTable1)
	require.NoError(t, err)

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "TestGlobalTable2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set("dynamodb:global-table:TestGlobalTable2", globalTable2)
	require.NoError(t, err)

	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 1)

	table := tables[0].(map[string]interface{})
	require.Equal(t, "TestGlobalTable1", table["GlobalTableName"])
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add multiple global tables to state
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "TestGlobalTable" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set("dynamodb:global-table:TestGlobalTable"+string(rune('0'+i)), globalTable)
		require.NoError(t, err)
	}

	limit := int32(3)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	tables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 3)

	// Should have LastEvaluatedGlobalTableName since we have more results
	_, hasLastEvaluated := response["LastEvaluatedGlobalTableName"]
	require.True(t, hasLastEvaluated)
}
