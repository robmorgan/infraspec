package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithGlobalTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables
	state.Set("dynamodb:globaltable:table-a", map[string]interface{}{
		"GlobalTableName":   "table-a",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/table-a",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	})
	state.Set("dynamodb:globaltable:table-b", map[string]interface{}{
		"GlobalTableName":   "table-b",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/table-b",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	})

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, tables, 2)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	state.Set("dynamodb:globaltable:table-a", map[string]interface{}{
		"GlobalTableName":   "table-a",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	})
	state.Set("dynamodb:globaltable:table-b", map[string]interface{}{
		"GlobalTableName":   "table-b",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	})

	// Filter by region us-east-1
	input := &ListGlobalTablesInput{
		RegionName: strPtr("us-east-1"),
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, tables, 1)

	table := tables[0].(map[string]interface{})
	assert.Equal(t, "table-a", table["GlobalTableName"])
}

func TestListGlobalTables_FilterByRegionNoMatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	state.Set("dynamodb:globaltable:table-a", map[string]interface{}{
		"GlobalTableName":   "table-a",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
		},
	})

	input := &ListGlobalTablesInput{
		RegionName: strPtr("eu-central-1"),
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithLimit(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create 3 global tables
	for _, name := range []string{"table-a", "table-b", "table-c"} {
		state.Set("dynamodb:globaltable:"+name, map[string]interface{}{
			"GlobalTableName":   name,
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		})
	}

	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, tables, 2)

	// LastEvaluatedGlobalTableName should be present
	assert.Contains(t, result, "LastEvaluatedGlobalTableName")
}

func TestListGlobalTables_ReplicationGroupIncluded(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	state.Set("dynamodb:globaltable:my-table", map[string]interface{}{
		"GlobalTableName":   "my-table",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	require.Len(t, tables, 1)

	table := tables[0].(map[string]interface{})
	assert.Equal(t, "my-table", table["GlobalTableName"])
	assert.Contains(t, table, "ReplicationGroup")
}
