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

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "GlobalTables")
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create two global tables
	globalTable1 := map[string]interface{}{
		"GlobalTableName":   "global-table-1",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/global-table-1",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	err := state.Set("dynamodb:globaltable:global-table-1", globalTable1)
	require.NoError(t, err)

	globalTable2 := map[string]interface{}{
		"GlobalTableName":   "global-table-2",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/global-table-2",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	}
	err = state.Set("dynamodb:globaltable:global-table-2", globalTable2)
	require.NoError(t, err)

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2)

	for _, table := range tables {
		tableMap, ok := table.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, tableMap, "GlobalTableName")
		assert.Contains(t, tableMap, "ReplicationGroup")
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables in different regions
	globalTable1 := map[string]interface{}{
		"GlobalTableName":   "global-table-us",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
		},
	}
	err := state.Set("dynamodb:globaltable:global-table-us", globalTable1)
	require.NoError(t, err)

	globalTable2 := map[string]interface{}{
		"GlobalTableName":   "global-table-eu",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	}
	err = state.Set("dynamodb:globaltable:global-table-eu", globalTable2)
	require.NoError(t, err)

	// Filter to only us-east-1 tables
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
	require.True(t, ok)
	assert.Len(t, tables, 1)

	tableMap, ok := tables[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "global-table-us", tableMap["GlobalTableName"])
}

func TestListGlobalTables_WithLimit(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create 5 global tables
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		globalTable := map[string]interface{}{
			"GlobalTableName":   tableName,
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		err := state.Set(fmt.Sprintf("dynamodb:globaltable:%s", tableName), globalTable)
		require.NoError(t, err)
	}

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
	assert.Len(t, tables, 2)
	assert.Contains(t, responseBody, "LastEvaluatedGlobalTableName")
}

func TestListGlobalTables_ViaHandleRequest(t *testing.T) {
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
	assert.Contains(t, responseBody, "GlobalTables")
}
