package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListGlobalTables_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create some global table data
	usEast1 := "us-east-1"
	usWest2 := "us-west-2"

	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": usEast1},
			map[string]interface{}{"RegionName": usWest2},
		},
	}
	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": usEast1},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", gt1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", gt2))

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
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.GlobalTables, 2)
}

func TestListGlobalTables_WithRegionFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	usEast1 := "us-east-1"
	usWest2 := "us-west-2"
	euWest1 := "eu-west-1"

	// Setup: Create some global table data
	gt1 := map[string]interface{}{
		"GlobalTableName": "global-table-1",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": usEast1},
			map[string]interface{}{"RegionName": usWest2},
		},
	}
	gt2 := map[string]interface{}{
		"GlobalTableName": "global-table-2",
		"ReplicationGroup": []interface{}{
			map[string]interface{}{"RegionName": euWest1},
		},
	}

	require.NoError(t, state.Set("dynamodb:global-table:global-table-1", gt1))
	require.NoError(t, state.Set("dynamodb:global-table:global-table-2", gt2))

	input := &ListGlobalTablesInput{
		RegionName: &usEast1,
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
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.GlobalTables, 1)
	require.Equal(t, "global-table-1", *output.GlobalTables[0].GlobalTableName)
}

func TestListGlobalTables_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	usEast1 := "us-east-1"

	// Setup: Create multiple global tables
	for i := 1; i <= 5; i++ {
		gt := map[string]interface{}{
			"GlobalTableName": "global-table-" + string(rune('0'+i)),
			"ReplicationGroup": []interface{}{
				map[string]interface{}{"RegionName": usEast1},
			},
		}
		require.NoError(t, state.Set("dynamodb:global-table:global-table-"+string(rune('0'+i)), gt))
	}

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
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.GlobalTables, 2)
	require.NotNil(t, output.LastEvaluatedGlobalTableName) // Should have more results
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
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListGlobalTablesOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.GlobalTables, 0)
	require.Nil(t, output.LastEvaluatedGlobalTableName)
}
