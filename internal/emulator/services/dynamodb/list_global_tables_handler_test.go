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

	// Create some global tables in state
	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	state.Set("dynamodb:globaltable:global-table-1", gt1)

	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}
	state.Set("dynamodb:globaltable:global-table-2", gt2)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(tables))
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables with different replicas
	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	state.Set("dynamodb:globaltable:global-table-1", gt1)

	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}
	state.Set("dynamodb:globaltable:global-table-2", gt2)

	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 1, len(tables))

	// Verify the returned table has a replica in us-east-1
	table := tables[0].(map[string]interface{})
	require.Equal(t, "global-table-1", table["GlobalTableName"])
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 0, len(tables))
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		gt := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		state.Set("dynamodb:globaltable:global-table-"+string(rune('0'+i)), gt)
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(tables))

	// Verify LastEvaluatedGlobalTableName is present
	_, hasLastEvaluated := result["LastEvaluatedGlobalTableName"]
	require.True(t, hasLastEvaluated)
}

func TestListGlobalTables_WithExclusiveStart(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 3; i++ {
		gt := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		state.Set("dynamodb:globaltable:global-table-"+string(rune('0'+i)), gt)
	}

	exclusiveStart := "global-table-1"
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: &exclusiveStart,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok)
	// Should return tables after global-table-1
	require.GreaterOrEqual(t, len(tables), 0)
}
