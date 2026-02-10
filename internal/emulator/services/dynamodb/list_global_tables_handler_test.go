package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some global tables
	globalTable1 := map[string]interface{}{
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
	require.NoError(t, state.Set("dynamodb:globaltable:global-table-1", globalTable1))

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	require.NoError(t, state.Set("dynamodb:globaltable:global-table-2", globalTable2))

	// Test listing all global tables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	globalTables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, globalTables, 2)

	// Verify first global table
	gt1 := globalTables[0].(map[string]interface{})
	require.Equal(t, "global-table-1", gt1["GlobalTableName"])

	// Verify second global table
	gt2 := globalTables[1].(map[string]interface{})
	require.Equal(t, "global-table-2", gt2["GlobalTableName"])
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables with different regions
	globalTable1 := map[string]interface{}{
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
	require.NoError(t, state.Set("dynamodb:globaltable:global-table-1", globalTable1))

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	require.NoError(t, state.Set("dynamodb:globaltable:global-table-2", globalTable2))

	// Test filtering by region
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	globalTables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, globalTables, 1)

	// Verify it's the correct global table
	gt := globalTables[0].(map[string]interface{})
	require.Equal(t, "global-table-1", gt["GlobalTableName"])
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	globalTables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Empty(t, globalTables)
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": fmt.Sprintf("global-table-%d", i),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		key := fmt.Sprintf("dynamodb:globaltable:global-table-%d", i)
		require.NoError(t, state.Set(key, globalTable))
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	globalTables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(globalTables))

	// Verify LastEvaluatedGlobalTableName is present
	_, hasLastEvaluated := output["LastEvaluatedGlobalTableName"]
	require.True(t, hasLastEvaluated)
}

func TestListGlobalTables_WithExclusiveStartGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": fmt.Sprintf("global-table-%d", i),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		key := fmt.Sprintf("dynamodb:globaltable:global-table-%d", i)
		require.NoError(t, state.Set(key, globalTable))
	}

	// Use exclusive start to skip first 2 tables
	exclusiveStart := "global-table-2"
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: &exclusiveStart,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	globalTables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 3, len(globalTables)) // Should get tables 3, 4, 5

	// Verify first table is global-table-3
	gt := globalTables[0].(map[string]interface{})
	require.Equal(t, "global-table-3", gt["GlobalTableName"])
}
