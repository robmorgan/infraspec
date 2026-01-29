package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	service, state := setupTestService(t)

	// Create test global table data
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", globalTable2))

	// Test listing all global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	service, state := setupTestService(t)

	// Create test global table data
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	globalTable2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", globalTable2))

	// Test listing global tables with replica in specific region
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 1)
	assert.Equal(t, "global-table-1", *output.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	service, _ := setupTestService(t)

	// Test listing when no global tables exist
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 0)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	service, state := setupTestService(t)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		key := "dynamodb:global-table:global-table-" + string(rune('0'+i))
		require.NoError(t, state.Set(key, globalTable))
	}

	// Test pagination with Limit
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 2)
	assert.NotNil(t, output.LastEvaluatedGlobalTableName)
}

func TestListGlobalTables_PaginationWithExclusiveStart(t *testing.T) {
	service, state := setupTestService(t)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		key := "dynamodb:global-table:global-table-" + string(rune('0'+i))
		require.NoError(t, state.Set(key, globalTable))
	}

	// Test pagination with ExclusiveStartGlobalTableName
	exclusiveStart := "global-table-2"
	limit := int32(10)
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: &exclusiveStart,
		Limit:                         &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Should return tables after global-table-2
	assert.True(t, len(output.GlobalTables) >= 1)
	if len(output.GlobalTables) > 0 {
		assert.NotEqual(t, "global-table-2", *output.GlobalTables[0].GlobalTableName)
	}
}
