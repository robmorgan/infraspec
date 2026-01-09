package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}

	require.NoError(t, state.Set("dynamodb:global_table:test-global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global_table:test-global-table-2", globalTable2))

	// Test without filter
	input := &ListGlobalTablesInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(tables))
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	globalTable1 := map[string]interface{}{
		"GlobalTableName": "test-global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}

	globalTable2 := map[string]interface{}{
		"GlobalTableName": "test-global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}

	require.NoError(t, state.Set("dynamodb:global_table:test-global-table-1", globalTable1))
	require.NoError(t, state.Set("dynamodb:global_table:test-global-table-2", globalTable2))

	// Test with region filter
	regionName := "us-east-1"
	input := &ListGlobalTablesInput{
		RegionName: &regionName,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(tables))

	// Verify it's the correct table
	firstTable := tables[0].(map[string]interface{})
	assert.Equal(t, "test-global-table-1", firstTable["GlobalTableName"])
}

func TestListGlobalTables_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListGlobalTablesInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(tables))
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	for i := 0; i < 5; i++ {
		globalTable := map[string]interface{}{
			"GlobalTableName": "test-global-table-" + string(rune('a'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		require.NoError(t, state.Set("dynamodb:global_table:test-global-table-"+string(rune('a'+i)), globalTable))
	}

	// Test with limit
	limit := int32(2)
	input := &ListGlobalTablesInput{
		Limit: &limit,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListGlobalTables",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	tables, ok := result["GlobalTables"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(tables))

	// Should have LastEvaluatedGlobalTableName since there are more results
	_, hasLastEvaluated := result["LastEvaluatedGlobalTableName"]
	assert.True(t, hasLastEvaluated)
}
