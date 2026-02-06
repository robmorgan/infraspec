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

	// Create test global tables
	gt1Key := "dynamodb:global-table:global-table-1"
	gt1Data := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	gt2Key := "dynamodb:global-table:global-table-2"
	gt2Data := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
	require.NoError(t, err)

	// Test ListGlobalTables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	globalTables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(globalTables))
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables with different regions
	gt1Key := "dynamodb:global-table:global-table-1"
	gt1Data := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	gt2Key := "dynamodb:global-table:global-table-2"
	gt2Data := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
	require.NoError(t, err)

	// Filter by us-east-1 region
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	globalTables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, len(globalTables))

	table := globalTables[0].(map[string]interface{})
	assert.Equal(t, "global-table-1", table["GlobalTableName"])
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test without any global tables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	globalTables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(globalTables))
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		gtKey := "dynamodb:global-table:global-table-" + string(rune('0'+i))
		gtData := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set(gtKey, gtData)
		require.NoError(t, err)
	}

	limit := int32(3)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	globalTables, ok := response["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(globalTables), 3)

	// Should have LastEvaluatedGlobalTableName since we have more results
	if len(globalTables) == 3 {
		_, hasLastEvaluated := response["LastEvaluatedGlobalTableName"]
		assert.True(t, hasLastEvaluated)
	}
}
