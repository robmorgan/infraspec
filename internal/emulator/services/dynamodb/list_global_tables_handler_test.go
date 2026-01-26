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
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	table2Name := "global-table-2"
	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "GlobalTables")
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Equal(t, 2, len(tables), "Should have two global tables")
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables with different regions
	table1Name := "global-table-1"
	table1Key := fmt.Sprintf("dynamodb:global-table:%s", table1Name)
	table1Data := map[string]interface{}{
		"GlobalTableName": table1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	table2Name := "global-table-2"
	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

	// Filter by us-east-1 region
	reqBody := `{"RegionName": "us-east-1"}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListGlobalTables",
		},
		Body:   []byte(reqBody),
		Action: "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Equal(t, 1, len(tables), "Should have one global table in us-east-1")

	// Verify it's the correct table
	if len(tables) > 0 {
		table := tables[0].(map[string]interface{})
		assert.Equal(t, table1Name, table["GlobalTableName"])
	}
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test global tables
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableKey := fmt.Sprintf("dynamodb:global-table:%s", tableName)
		tableData := map[string]interface{}{
			"GlobalTableName": tableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		err := state.Set(tableKey, tableData)
		require.NoError(t, err)
	}

	// Request with Limit = 2
	reqBody := `{"Limit": 2}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListGlobalTables",
		},
		Body:   []byte(reqBody),
		Action: "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify pagination
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Equal(t, 2, len(tables), "Should return only 2 results due to Limit")

	// Should have a LastEvaluatedGlobalTableName since there are more results
	assert.Contains(t, responseBody, "LastEvaluatedGlobalTableName", "Should have LastEvaluatedGlobalTableName for pagination")
}

func TestListGlobalTables_WithExclusiveStart(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	for i := 1; i <= 3; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableKey := fmt.Sprintf("dynamodb:global-table:%s", tableName)
		tableData := map[string]interface{}{
			"GlobalTableName": tableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		err := state.Set(tableKey, tableData)
		require.NoError(t, err)
	}

	// Request starting from global-table-1
	reqBody := `{"ExclusiveStartGlobalTableName": "global-table-1"}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListGlobalTables",
		},
		Body:   []byte(reqBody),
		Action: "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify results don't include global-table-1
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")

	// Should have tables after global-table-1
	for _, table := range tables {
		tableMap := table.(map[string]interface{})
		assert.NotEqual(t, "global-table-1", tableMap["GlobalTableName"])
	}
}
