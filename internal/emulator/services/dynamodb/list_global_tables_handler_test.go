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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "GlobalTables")
	tables, ok := result["GlobalTables"].([]interface{})
	require.True(t, ok, "GlobalTables should be an array")
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create two global tables
	table1Name := "global-table-1"
	table2Name := "global-table-2"

	state.Set(fmt.Sprintf("dynamodb:globaltable:%s", table1Name), map[string]interface{}{
		"GlobalTableName":   table1Name,
		"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", table1Name),
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
			map[string]interface{}{"RegionName": "eu-west-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	state.Set(fmt.Sprintf("dynamodb:globaltable:%s", table2Name), map[string]interface{}{
		"GlobalTableName":   table2Name,
		"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", table2Name),
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "ap-southeast-1", "ReplicaStatus": "ACTIVE"},
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
	require.True(t, ok)
	assert.Len(t, tables, 2)

	for _, tbl := range tables {
		m := tbl.(map[string]interface{})
		assert.Contains(t, m, "GlobalTableName")
		assert.Contains(t, m, "ReplicationGroup")
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1Name := "global-table-us"
	table2Name := "global-table-eu"

	state.Set(fmt.Sprintf("dynamodb:globaltable:%s", table1Name), map[string]interface{}{
		"GlobalTableName":   table1Name,
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	state.Set(fmt.Sprintf("dynamodb:globaltable:%s", table2Name), map[string]interface{}{
		"GlobalTableName":   table2Name,
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	// Filter to only us-east-1
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
	require.True(t, ok)
	assert.Len(t, tables, 1)

	m := tables[0].(map[string]interface{})
	assert.Equal(t, table1Name, m["GlobalTableName"])
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create 3 global tables
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("global-table-%d", i)
		state.Set(fmt.Sprintf("dynamodb:globaltable:%s", name), map[string]interface{}{
			"GlobalTableName":   name,
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
			},
		})
	}

	// Request with limit of 2
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
	require.True(t, ok)
	assert.Len(t, tables, 2)

	// Should have pagination token
	assert.Contains(t, result, "LastEvaluatedGlobalTableName")
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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "GlobalTables")
}
