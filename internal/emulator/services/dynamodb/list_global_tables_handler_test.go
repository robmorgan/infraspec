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
	require.True(t, ok)
	assert.Empty(t, tables)
}

func TestListGlobalTables_WithGlobalTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Seed global tables
	state.Set("dynamodb:globaltable:global-table-1", map[string]interface{}{
		"GlobalTableName":   "global-table-1",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/global-table-1",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
			map[string]interface{}{"RegionName": "eu-west-1", "ReplicaStatus": "ACTIVE"},
		},
	})
	state.Set("dynamodb:globaltable:global-table-2", map[string]interface{}{
		"GlobalTableName":   "global-table-2",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/global-table-2",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-west-2", "ReplicaStatus": "ACTIVE"},
		},
	})

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

	for _, t_ := range tables {
		tMap, ok := t_.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, tMap, "GlobalTableName")
		assert.Contains(t, tMap, "ReplicationGroup")
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Table with us-east-1 replica
	state.Set("dynamodb:globaltable:table-a", map[string]interface{}{
		"GlobalTableName":   "table-a",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/table-a",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	// Table with only eu-west-1 replica
	state.Set("dynamodb:globaltable:table-b", map[string]interface{}{
		"GlobalTableName":   "table-b",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/table-b",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "eu-west-1", "ReplicaStatus": "ACTIVE"},
		},
	})

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
	assert.Len(t, tables, 1)

	tMap := tables[0].(map[string]interface{})
	assert.Equal(t, "table-a", tMap["GlobalTableName"])
}

func TestListGlobalTables_FilterByRegion_NoMatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	state.Set("dynamodb:globaltable:table-x", map[string]interface{}{
		"GlobalTableName":   "table-x",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/table-x",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	region := "ap-southeast-1"
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
	assert.Empty(t, tables)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Seed 4 global tables
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("gt-%d", i)
		state.Set(fmt.Sprintf("dynamodb:globaltable:%s", name), map[string]interface{}{
			"GlobalTableName":   name,
			"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", name),
			"GlobalTableStatus": "ACTIVE",
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tables, 2)

	// Should have LastEvaluatedGlobalTableName for pagination
	assert.Contains(t, responseBody, "LastEvaluatedGlobalTableName")
}

func TestListGlobalTables_ReplicationGroupInResponse(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	state.Set("dynamodb:globaltable:multi-region", map[string]interface{}{
		"GlobalTableName":   "multi-region",
		"GlobalTableArn":    "arn:aws:dynamodb::000000000000:global-table/multi-region",
		"GlobalTableStatus": "ACTIVE",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": "us-east-1", "ReplicaStatus": "ACTIVE"},
			map[string]interface{}{"RegionName": "eu-west-1", "ReplicaStatus": "ACTIVE"},
			map[string]interface{}{"RegionName": "ap-south-1", "ReplicaStatus": "ACTIVE"},
		},
	})

	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	tables, ok := responseBody["GlobalTables"].([]interface{})
	require.True(t, ok)
	require.Len(t, tables, 1)

	tMap := tables[0].(map[string]interface{})
	replicas, ok := tMap["ReplicationGroup"].([]interface{})
	require.True(t, ok)
	assert.Len(t, replicas, 3)

	// Each replica entry should have RegionName
	for _, r := range replicas {
		rMap, ok := r.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, rMap, "RegionName")
	}
}
