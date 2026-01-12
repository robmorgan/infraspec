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
			map[string]interface{}{
				"RegionName": "ap-southeast-1",
			},
		},
	}

	err := state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:global-table:test-global-table-2", globalTable2)
	require.NoError(t, err)

	// Test list all global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables := responseData["GlobalTables"].([]interface{})
	assert.Equal(t, 2, len(globalTables))
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
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
			map[string]interface{}{
				"RegionName": "ap-southeast-1",
			},
		},
	}
	globalTable3 := map[string]interface{}{
		"GlobalTableName": "test-global-table-3",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}

	err := state.Set("dynamodb:global-table:test-global-table-1", globalTable1)
	require.NoError(t, err)
	err = state.Set("dynamodb:global-table:test-global-table-2", globalTable2)
	require.NoError(t, err)
	err = state.Set("dynamodb:global-table:test-global-table-3", globalTable3)
	require.NoError(t, err)

	// Test filter by region
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

	globalTables := responseData["GlobalTables"].([]interface{})
	assert.Equal(t, 2, len(globalTables))

	// Verify both global tables have us-east-1 replica
	for _, gt := range globalTables {
		gtMap := gt.(map[string]interface{})
		replicas := gtMap["ReplicationGroup"].([]interface{})
		hasRegion := false
		for _, replica := range replicas {
			replicaMap := replica.(map[string]interface{})
			if replicaMap["RegionName"] == "us-east-1" {
				hasRegion = true
				break
			}
		}
		assert.True(t, hasRegion)
	}
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables := responseData["GlobalTables"].([]interface{})
	assert.Equal(t, 0, len(globalTables))
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set("dynamodb:global-table:test-global-table-"+string(rune('0'+i)), globalTable)
		require.NoError(t, err)
	}

	// Test with limit
	limit := int32(3)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	globalTables := responseData["GlobalTables"].([]interface{})
	assert.Equal(t, 3, len(globalTables))

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := responseData["LastEvaluatedGlobalTableName"]
	assert.True(t, hasLastEvaluated)
}
