package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/testing/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - global tables
	globalTableData1 := map[string]interface{}{
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
	err := state.Set("dynamodb:global-table:global-table-1", globalTableData1)
	require.NoError(t, err)

	globalTableData2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set("dynamodb:global-table:global-table-2", globalTableData2)
	require.NoError(t, err)

	// Test listing all global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	globalTableData1 := map[string]interface{}{
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
	err := state.Set("dynamodb:global-table:global-table-1", globalTableData1)
	require.NoError(t, err)

	globalTableData2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set("dynamodb:global-table:global-table-2", globalTableData2)
	require.NoError(t, err)

	globalTableData3 := map[string]interface{}{
		"GlobalTableName": "global-table-3",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
			map[string]interface{}{
				"RegionName": "eu-central-1",
			},
		},
	}
	err = state.Set("dynamodb:global-table:global-table-3", globalTableData3)
	require.NoError(t, err)

	// Test filtering by region
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Should return tables that have a replica in us-east-1
	assert.Len(t, output.GlobalTables, 2)
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Empty(t, output.GlobalTables)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - 5 global tables
	for i := 1; i <= 5; i++ {
		globalTableData := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), globalTableData)
		require.NoError(t, err)
	}

	// Request with limit of 2
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 2)
	assert.NotNil(t, output.LastEvaluatedGlobalTableName)
}
