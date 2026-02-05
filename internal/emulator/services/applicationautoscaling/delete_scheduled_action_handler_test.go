package applicationautoscaling

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func strPtr(s string) *string {
	return &s
}

func TestDeleteScheduledAction_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create a scheduled action first
	scheduledAction := map[string]interface{}{
		"ScheduledActionName": "test-action",
		"ServiceNamespace":    "dynamodb",
		"ResourceId":          "table/test-table",
		"ScalableDimension":   "dynamodb:table:ReadCapacityUnits",
	}
	key := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:test-action"
	err := state.Set(key, scheduledAction)
	require.NoError(t, err)

	// Delete the scheduled action
	input := &DeleteScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.1", resp.Headers["Content-Type"])

	// Verify response is empty JSON object
	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)
	assert.Empty(t, responseData)

	// Verify the scheduled action was deleted
	assert.False(t, state.Exists(key))
}

func TestDeleteScheduledAction_MissingScheduledActionName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DeleteScheduledActionRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ScheduledActionName")
}

func TestDeleteScheduledAction_MissingServiceNamespace(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DeleteScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ServiceNamespace")
}

func TestDeleteScheduledAction_MissingResourceId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DeleteScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ResourceId")
}

func TestDeleteScheduledAction_MissingScalableDimension(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DeleteScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ScalableDimension")
}

func TestDeleteScheduledAction_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &DeleteScheduledActionRequest{
		ScheduledActionName: strPtr("nonexistent-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.deleteScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ObjectNotFoundException", errorData["__type"])
	assert.Contains(t, errorData["message"], "not found")
}
