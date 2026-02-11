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
	table2Name := "global-table-2"

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
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

	// Test without filter
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

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2, "Should have 2 global tables")

	// Verify table structure
	for _, table := range tables {
		tableMap := table.(map[string]interface{})
		assert.Contains(t, tableMap, "GlobalTableName")
		assert.Contains(t, tableMap, "ReplicationGroup")
	}
}

func TestListGlobalTables_WithRegionFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	table1Name := "global-table-1"
	table2Name := "global-table-2"

	// Table 1 has replicas in us-east-1 and us-west-2
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
	}
	err := state.Set(table1Key, table1Data)
	require.NoError(t, err)

	// Table 2 only has replica in us-east-1
	table2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	table2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
		},
	}
	err = state.Set(table2Key, table2Data)
	require.NoError(t, err)

	// Test with RegionName filter for us-west-2
	reqBody := `{"RegionName": "us-west-2"}`
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

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 1, "Should have 1 global table with replica in us-west-2")
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
		}
		err := state.Set(tableKey, tableData)
		require.NoError(t, err)
	}

	// First page with Limit=2
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListGlobalTables",
		},
		Body:   []byte(`{"Limit": 2}`),
		Action: "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2, "Should have 2 tables in first page")

	// Verify LastEvaluatedGlobalTableName is present
	lastEvaluated, hasNext := responseBody["LastEvaluatedGlobalTableName"].(string)
	assert.True(t, hasNext, "Should have LastEvaluatedGlobalTableName for more results")

	// Second page using ExclusiveStartGlobalTableName
	reqBody := fmt.Sprintf(`{"Limit": 2, "ExclusiveStartGlobalTableName": "%s"}`, lastEvaluated)
	req.Body = []byte(reqBody)

	resp, err = service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok = responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2, "Should have 2 tables in second page")
}
