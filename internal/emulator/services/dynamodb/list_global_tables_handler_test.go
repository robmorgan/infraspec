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
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Empty(t, output.GlobalTables)
	require.Nil(t, output.LastEvaluatedGlobalTableName)
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
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
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", globalTable2))

	// List all global tables
	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.GlobalTables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
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
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", globalTable2))

	// Filter by region
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.GlobalTables, 1)
	require.Equal(t, "global-table-1", *output.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		require.NoError(t, state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), globalTable))
	}

	// Request with limit
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.GlobalTables, 2)
	require.NotNil(t, output.LastEvaluatedGlobalTableName)
}

func TestListGlobalTables_PaginationWithExclusiveStart(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		require.NoError(t, state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), globalTable))
	}

	// Request first page
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.GlobalTables, 2)
	require.NotNil(t, output.LastEvaluatedGlobalTableName)

	// Request second page
	input2 := &ListGlobalTablesInput{
		Limit:                         &limit,
		ExclusiveStartGlobalTableName: output.LastEvaluatedGlobalTableName,
	}
	resp2, err := service.listGlobalTables(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var output2 ListGlobalTablesOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)
	require.Len(t, output2.GlobalTables, 2)
}
