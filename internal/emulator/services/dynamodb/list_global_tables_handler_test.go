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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Empty(t, responseBody.GlobalTables, "Should have no global tables initially")
	assert.Nil(t, responseBody.LastEvaluatedGlobalTableName, "Should have no last evaluated table name")
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	globalTable1Name := "global-table-1"
	globalTable1Key := fmt.Sprintf("dynamodb:global-table:%s", globalTable1Name)
	globalTable1Data := map[string]interface{}{
		"GlobalTableName": globalTable1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	err := state.Set(globalTable1Key, globalTable1Data)
	require.NoError(t, err)

	globalTable2Name := "global-table-2"
	globalTable2Key := fmt.Sprintf("dynamodb:global-table:%s", globalTable2Name)
	globalTable2Data := map[string]interface{}{
		"GlobalTableName": globalTable2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "eu-west-1"},
		},
	}
	err = state.Set(globalTable2Key, globalTable2Data)
	require.NoError(t, err)

	// List all global tables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.GlobalTables, 2, "Should have two global tables")

	// Verify tables contain expected fields
	for _, table := range responseBody.GlobalTables {
		assert.NotNil(t, table.GlobalTableName)
		assert.NotEmpty(t, table.ReplicationGroup)
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables in different regions
	globalTable1Name := "global-table-1"
	globalTable1Key := fmt.Sprintf("dynamodb:global-table:%s", globalTable1Name)
	globalTable1Data := map[string]interface{}{
		"GlobalTableName": globalTable1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1"},
			map[string]interface{}{"RegionName": "us-west-2"},
		},
	}
	err := state.Set(globalTable1Key, globalTable1Data)
	require.NoError(t, err)

	globalTable2Name := "global-table-2"
	globalTable2Key := fmt.Sprintf("dynamodb:global-table:%s", globalTable2Name)
	globalTable2Data := map[string]interface{}{
		"GlobalTableName": globalTable2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1"},
			map[string]interface{}{"RegionName": "ap-southeast-1"},
		},
	}
	err = state.Set(globalTable2Key, globalTable2Data)
	require.NoError(t, err)

	// List global tables in us-east-1
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.GlobalTables, 1, "Should have only one global table in us-east-1")
	assert.Equal(t, globalTable1Name, *responseBody.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 1; i <= 5; i++ {
		globalTableName := fmt.Sprintf("global-table-%d", i)
		globalTableKey := fmt.Sprintf("dynamodb:global-table:%s", globalTableName)
		globalTableData := map[string]interface{}{
			"GlobalTableName": globalTableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1"},
			},
		}
		err := state.Set(globalTableKey, globalTableData)
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.GlobalTables, 2, "Should have only 2 global tables due to limit")

	// Should have LastEvaluatedGlobalTableName for pagination
	assert.NotNil(t, responseBody.LastEvaluatedGlobalTableName)
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.NotNil(t, responseBody.GlobalTables)
}
