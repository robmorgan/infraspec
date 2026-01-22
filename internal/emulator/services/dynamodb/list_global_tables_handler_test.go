package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListGlobalTables",
		},
		Body:   []byte("{}"),
		Action: "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "GlobalTables")
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Empty(t, tables, "Should have no global tables initially")
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	table1Name := "global-table-1"
	table1Key := fmt.Sprintf("dynamodb:global-table:%s", table1Name)
	table1Data := map[string]interface{}{
		"GlobalTableName": table1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
		"Replicas": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	table2Name := "global-table-2"
	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
		"Replicas": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

	// List all global tables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Len(t, tables, 2, "Should have two global tables")

	// Verify tables contain expected fields
	for _, table := range tables {
		tableMap, ok := table.(map[string]interface{})
		require.True(t, ok, "Each table should be an object")
		assert.Contains(t, tableMap, "GlobalTableName")
		assert.Contains(t, tableMap, "ReplicationGroup")
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables with different regions
	table1Name := "global-table-1"
	table1Key := fmt.Sprintf("dynamodb:global-table:%s", table1Name)
	table1Data := map[string]interface{}{
		"GlobalTableName": table1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
		},
		"Replicas": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
		},
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	table2Name := "global-table-2"
	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
		"Replicas": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

	// List global tables for us-east-1 only
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Len(t, tables, 1, "Should have only one global table for us-east-1")

	tableMap, ok := tables[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Name, tableMap["GlobalTableName"])
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableKey := fmt.Sprintf("dynamodb:global-table:%s", tableName)
		tableData := map[string]interface{}{
			"GlobalTableName": tableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
			"Replicas": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set(tableKey, tableData)
		require.NoError(t, err)
	}

	// List global tables with limit
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Len(t, tables, 2, "Should have only 2 global tables due to limit")

	// Should have LastEvaluatedGlobalTableName for pagination
	assert.Contains(t, responseBody, "LastEvaluatedGlobalTableName")
}

func TestListGlobalTables_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListGlobalTables should work with no parameters
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "GlobalTables")
}
