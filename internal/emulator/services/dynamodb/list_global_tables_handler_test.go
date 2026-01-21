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

	// Create test data - global tables
	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", gt1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", gt2))

	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	tables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", gt1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", gt2))

	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	tables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 1)

	// Verify it's the correct table
	table := tables[0].(map[string]interface{})
	assert.Equal(t, "global-table-1", table["GlobalTableName"])
}

func TestListGlobalTables_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	tables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 0)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	for i := 0; i < 5; i++ {
		gt := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		require.NoError(t, state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), gt))
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	tables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2)

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := output["LastEvaluatedGlobalTableName"]
	assert.True(t, hasLastEvaluated)
}

func TestListGlobalTables_ExclusiveStart(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	for i := 0; i < 3; i++ {
		gt := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		require.NoError(t, state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), gt))
	}

	exclusiveStart := "global-table-0"
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: &exclusiveStart,
	}
	resp, err := service.listGlobalTables(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	tables, ok := output["GlobalTables"].([]interface{})
	require.True(t, ok)
	// Should get tables after global-table-0
	assert.GreaterOrEqual(t, len(tables), 1)
}
