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

func TestListGlobalTables_WithGlobalTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	globalTable1 := "global-table-1"
	globalTable1Key := fmt.Sprintf("dynamodb:globaltable:%s", globalTable1)
	globalTable1Data := map[string]interface{}{
		"GlobalTableName": globalTable1,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(globalTable1Key, globalTable1Data)
	require.NoError(t, err)

	globalTable2 := "global-table-2"
	globalTable2Key := fmt.Sprintf("dynamodb:globaltable:%s", globalTable2)
	globalTable2Data := map[string]interface{}{
		"GlobalTableName": globalTable2,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(globalTable2Key, globalTable2Data)
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
	assert.Contains(t, responseBody, "GlobalTables")
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Len(t, tables, 2, "Should have 2 global tables")
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables in different regions
	globalTable1 := "global-table-1"
	globalTable1Key := fmt.Sprintf("dynamodb:globaltable:%s", globalTable1)
	globalTable1Data := map[string]interface{}{
		"GlobalTableName": globalTable1,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(globalTable1Key, globalTable1Data)
	require.NoError(t, err)

	globalTable2 := "global-table-2"
	globalTable2Key := fmt.Sprintf("dynamodb:globaltable:%s", globalTable2)
	globalTable2Data := map[string]interface{}{
		"GlobalTableName": globalTable2,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(globalTable2Key, globalTable2Data)
	require.NoError(t, err)

	// List global tables in us-east-1 only
	region := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &region,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 1, "Should have 1 global table in us-east-1")

	// Verify it's the correct table
	table := tables[0].(map[string]interface{})
	assert.Equal(t, globalTable1, table["GlobalTableName"])
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableKey := fmt.Sprintf("dynamodb:globaltable:%s", tableName)
		tableData := map[string]interface{}{
			"GlobalTableName": tableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set(tableKey, tableData)
		require.NoError(t, err)
	}

	// List with pagination (2 results per page)
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
	require.True(t, ok)
	assert.Len(t, tables, 2, "Should have 2 global tables in first page")

	// Verify LastEvaluatedGlobalTableName is present
	assert.Contains(t, responseBody, "LastEvaluatedGlobalTableName")
	lastTableName, ok := responseBody["LastEvaluatedGlobalTableName"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, lastTableName)

	// Fetch next page
	input.ExclusiveStartGlobalTableName = &lastTableName
	resp, err = service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok = responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2, "Should have 2 global tables in second page")
}
