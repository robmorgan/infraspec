package applicationautoscaling

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeScheduledActions_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create a couple of scheduled actions
	action1 := map[string]interface{}{
		"ScheduledActionName": "action-1",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
		"Schedule":            "cron(0 10 * * ? *)",
	}
	action2 := map[string]interface{}{
		"ScheduledActionName": "action-2",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:WriteCapacityUnits",
		"Schedule":            "cron(0 12 * * ? *)",
	}

	key1 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:action-1"
	key2 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:WriteCapacityUnits:action-2"

	err := state.Set(key1, action1)
	require.NoError(t, err)
	err = state.Set(key2, action2)
	require.NoError(t, err)

	// Describe all scheduled actions
	input := &DescribeScheduledActionsRequest{
		ServiceNamespace: ServiceNamespace("dynamodb"),
	}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.1", resp.Headers["Content-Type"])

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	actions, ok := responseData["ScheduledActions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, actions, 2)
}

func TestDescribeScheduledActions_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DescribeScheduledActionsRequest{
		ServiceNamespace: ServiceNamespace("dynamodb"),
	}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	// ScheduledActions should be nil/null when empty (not an empty array)
	_, ok := responseData["ScheduledActions"]
	// Empty slice marshals to null, not []
	assert.True(t, !ok || responseData["ScheduledActions"] == nil)
}

func TestDescribeScheduledActions_FilterByScheduledActionNames(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create scheduled actions
	action1 := map[string]interface{}{
		"ScheduledActionName": "action-1",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}
	action2 := map[string]interface{}{
		"ScheduledActionName": "action-2",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}

	key1 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:action-1"
	key2 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:action-2"

	err := state.Set(key1, action1)
	require.NoError(t, err)
	err = state.Set(key2, action2)
	require.NoError(t, err)

	// Filter by specific action name
	input := &DescribeScheduledActionsRequest{
		ServiceNamespace:     ServiceNamespace("dynamodb"),
		ScheduledActionNames: []string{"action-1"},
	}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	actions, ok := responseData["ScheduledActions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, actions, 1)

	action := actions[0].(map[string]interface{})
	assert.Equal(t, "action-1", action["ScheduledActionName"])
}

func TestDescribeScheduledActions_FilterByResourceId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create scheduled actions for different resources
	action1 := map[string]interface{}{
		"ScheduledActionName": "action-1",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/table-1",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}
	action2 := map[string]interface{}{
		"ScheduledActionName": "action-2",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/table-2",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}

	key1 := "autoscaling:scheduled-action:dynamodb:table/table-1:dynamodb:table:ReadCapacityUnits:action-1"
	key2 := "autoscaling:scheduled-action:dynamodb:table/table-2:dynamodb:table:ReadCapacityUnits:action-2"

	err := state.Set(key1, action1)
	require.NoError(t, err)
	err = state.Set(key2, action2)
	require.NoError(t, err)

	// Filter by resource ID
	input := &DescribeScheduledActionsRequest{
		ServiceNamespace: ServiceNamespace("dynamodb"),
		ResourceId:       strPtr("table/table-1"),
	}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	actions, ok := responseData["ScheduledActions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, actions, 1)

	action := actions[0].(map[string]interface{})
	assert.Equal(t, "table/table-1", action["ResourceId"])
}

func TestDescribeScheduledActions_FilterByScalableDimension(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create scheduled actions for different dimensions
	action1 := map[string]interface{}{
		"ScheduledActionName": "action-1",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}
	action2 := map[string]interface{}{
		"ScheduledActionName": "action-2",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:WriteCapacityUnits",
	}

	key1 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:action-1"
	key2 := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:WriteCapacityUnits:action-2"

	err := state.Set(key1, action1)
	require.NoError(t, err)
	err = state.Set(key2, action2)
	require.NoError(t, err)

	// Filter by scalable dimension
	input := &DescribeScheduledActionsRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	actions, ok := responseData["ScheduledActions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, actions, 1)

	action := actions[0].(map[string]interface{})
	assert.Equal(t, "dynamodb:table:ReadCapacityUnits", action["ScalableDimension"])
}

func TestDescribeScheduledActions_MissingServiceNamespace(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DescribeScheduledActionsRequest{}

	resp, err := service.describeScheduledActions(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ServiceNamespace")
}
