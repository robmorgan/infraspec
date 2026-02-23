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

	assert.Contains(t, responseBody, "GlobalTables")
	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables in state
	for i := 1; i <= 3; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableData := map[string]interface{}{
			"GlobalTableName":   tableName,
			"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", tableName),
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName":    "us-east-1",
					"ReplicaStatus": "ACTIVE",
				},
			},
		}
		err := state.Set(fmt.Sprintf("dynamodb:globaltable:%s", tableName), tableData)
		require.NoError(t, err)
	}

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 3)

	for _, tbl := range tables {
		tm, ok := tbl.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, tm, "GlobalTableName")
		assert.Contains(t, tm, "ReplicationGroup")
	}
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create 5 global tables
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table-%d", i)
		tableData := map[string]interface{}{
			"GlobalTableName":   tableName,
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup":  []interface{}{},
		}
		err := state.Set(fmt.Sprintf("dynamodb:globaltable:%s", tableName), tableData)
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

func TestListGlobalTables_ExclusiveStartGlobalTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableNames := []string{"alpha", "beta", "gamma"}
	for _, name := range tableNames {
		tableData := map[string]interface{}{
			"GlobalTableName":   name,
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup":  []interface{}{},
		}
		err := state.Set(fmt.Sprintf("dynamodb:globaltable:%s", name), tableData)
		require.NoError(t, err)
	}

	// Start after "alpha"
	input := &ListGlobalTablesInput{
		ExclusiveStartGlobalTableName: strPtr("alpha"),
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	// Should get tables after "alpha"
	assert.LessOrEqual(t, len(tables), 2)
}

func TestListGlobalTables_NoLastEvaluatedWhenNoMore(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "only-table"
	tableData := map[string]interface{}{
		"GlobalTableName":   tableName,
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup":  []interface{}{},
	}
	err := state.Set(fmt.Sprintf("dynamodb:globaltable:%s", tableName), tableData)
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
	assert.Len(t, tables, 1)
	assert.NotContains(t, responseBody, "LastEvaluatedGlobalTableName")
}
