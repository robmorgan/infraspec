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
	assert.NotNil(t, responseBody.GlobalTables)
	assert.Empty(t, responseBody.GlobalTables, "Should have no global tables initially")
	assert.Nil(t, responseBody.LastEvaluatedGlobalTableName)
}

func TestListGlobalTables_WithGlobalTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	gt1Name := "global-table-1"
	gt1Key := fmt.Sprintf("dynamodb:global-table:%s", gt1Name)
	gt1Data := map[string]interface{}{
		"GlobalTableName": gt1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	gt2Name := "global-table-2"
	gt2Key := fmt.Sprintf("dynamodb:global-table:%s", gt2Name)
	gt2Data := map[string]interface{}{
		"GlobalTableName": gt2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.GlobalTables, 2)

	// Check first global table
	assert.NotNil(t, responseBody.GlobalTables[0].GlobalTableName)
	assert.NotNil(t, responseBody.GlobalTables[0].ReplicationGroup)
}

func TestListGlobalTables_FilterByRegion(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables with different replicas
	gt1Name := "global-table-1"
	gt1Key := fmt.Sprintf("dynamodb:global-table:%s", gt1Name)
	gt1Data := map[string]interface{}{
		"GlobalTableName": gt1Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "us-east-1",
			},
			map[string]interface{}{
				"RegionName": "us-west-2",
			},
		},
	}
	err := state.Set(gt1Key, gt1Data)
	require.NoError(t, err)

	gt2Name := "global-table-2"
	gt2Key := fmt.Sprintf("dynamodb:global-table:%s", gt2Name)
	gt2Data := map[string]interface{}{
		"GlobalTableName": gt2Name,
		"ReplicationGroup": []interface{}{
			map[string]interface{}{
				"RegionName": "eu-west-1",
			},
		},
	}
	err = state.Set(gt2Key, gt2Data)
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should only return global tables with replicas in us-east-1
	assert.Len(t, responseBody.GlobalTables, 1)
	assert.Equal(t, gt1Name, *responseBody.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test global tables
	for i := 1; i <= 5; i++ {
		gtName := fmt.Sprintf("global-table-%d", i)
		gtKey := fmt.Sprintf("dynamodb:global-table:%s", gtName)
		gtData := map[string]interface{}{
			"GlobalTableName": gtName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set(gtKey, gtData)
		require.NoError(t, err)
	}

	// Request with Limit
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should return only 2 results
	assert.Len(t, responseBody.GlobalTables, 2)
	// Should have LastEvaluatedGlobalTableName for more results
	assert.NotNil(t, responseBody.LastEvaluatedGlobalTableName)
}

func TestListGlobalTables_PaginationWithExclusiveStart(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test global tables
	gtNames := []string{"global-table-1", "global-table-2", "global-table-3"}
	for _, gtName := range gtNames {
		gtKey := fmt.Sprintf("dynamodb:global-table:%s", gtName)
		gtData := map[string]interface{}{
			"GlobalTableName": gtName,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		}
		err := state.Set(gtKey, gtData)
		require.NoError(t, err)
	}

	// Request with ExclusiveStartGlobalTableName
	reqBody := fmt.Sprintf(`{"ExclusiveStartGlobalTableName": "%s", "Limit": 2}`, gtNames[0])
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

	var responseBody ListGlobalTablesOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should start from the next table after the exclusive start
	assert.LessOrEqual(t, len(responseBody.GlobalTables), 2)
}
