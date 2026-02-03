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
	table1Name := "global-table1"
	region1 := "us-east-1"
	region2 := "us-west-2"

	gt1Key := fmt.Sprintf("dynamodb:global-table:%s", table1Name)
	gt1Data := map[string]interface{}{
		"GlobalTableName": table1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": region1,
			},
			map[string]interface{}{
				"RegionName": region2,
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	table2Name := "global-table2"
	gt2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	gt2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": region1,
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
	require.NoError(t, err)

	// List all global tables
	input := &ListGlobalTablesInput{}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, output.GlobalTables, 2, "Should have two global tables")

	// Verify tables contain expected fields
	for _, table := range output.GlobalTables {
		assert.NotNil(t, table.GlobalTableName)
		assert.NotNil(t, table.ReplicationGroup)
		assert.NotEmpty(t, table.ReplicationGroup)
	}
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create global tables with different regions
	region1 := "us-east-1"
	region2 := "us-west-2"
	region3 := "eu-west-1"

	table1Name := "global-table1"
	gt1Key := fmt.Sprintf("dynamodb:global-table:%s", table1Name)
	gt1Data := map[string]interface{}{
		"GlobalTableName": table1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": region1,
			},
			map[string]interface{}{
				"RegionName": region2,
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	table2Name := "global-table2"
	gt2Key := fmt.Sprintf("dynamodb:global-table:%s", table2Name)
	gt2Data := map[string]interface{}{
		"GlobalTableName": table2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": region3,
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
	require.NoError(t, err)

	// List global tables in us-east-1 region only
	input := &ListGlobalTablesInput{
		RegionName: &region1,
	}

	resp, err := service.listGlobalTables(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 1, "Should have only one global table in us-east-1")
	assert.Equal(t, table1Name, *output.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple global tables
	region := "us-east-1"
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("global-table%d", i)
		gtKey := fmt.Sprintf("dynamodb:global-table:%s", tableName)
		gtData := map[string]interface{}{
			"GlobalTableName": tableName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": region,
				},
			},
		}
		err := state.Set(gtKey, gtData)
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

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.GlobalTables, 2, "Should have only 2 global tables due to limit")

	// Should have LastEvaluatedGlobalTableName for pagination
	assert.NotNil(t, output.LastEvaluatedGlobalTableName, "Should have LastEvaluatedGlobalTableName when there are more results")
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

	var output ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.GlobalTables)
}
