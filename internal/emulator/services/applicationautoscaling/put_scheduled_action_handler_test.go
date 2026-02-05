package applicationautoscaling

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func TestPutScheduledAction_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		Schedule:            strPtr("cron(0 10 * * ? *)"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.1", resp.Headers["Content-Type"])

	// Verify response is empty JSON object
	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)
	assert.Empty(t, responseData)

	// Verify the scheduled action was created
	key := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:test-action"
	assert.True(t, state.Exists(key))

	// Verify stored data
	var storedAction ScheduledAction
	err = state.Get(key, &storedAction)
	require.NoError(t, err)
	assert.Equal(t, "test-action", *storedAction.ScheduledActionName)
	assert.Equal(t, ServiceNamespace("dynamodb"), storedAction.ServiceNamespace)
	assert.Equal(t, "table/test-table", *storedAction.ResourceId)
	assert.Equal(t, ScalableDimension("dynamodb:table:ReadCapacityUnits"), storedAction.ScalableDimension)
	assert.Equal(t, "cron(0 10 * * ? *)", *storedAction.Schedule)
	assert.NotNil(t, storedAction.CreationTime)
	assert.NotNil(t, storedAction.ScheduledActionARN)
}

func TestPutScheduledAction_WithAllOptionalFields(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	minCapacity := int32(5)
	maxCapacity := int32(10)

	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		Schedule:            strPtr("cron(0 10 * * ? *)"),
		Timezone:            strPtr("America/New_York"),
		StartTime:           &startTime,
		EndTime:             &endTime,
		ScalableTargetAction: &ScalableTargetAction{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		},
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify stored data includes all fields
	key := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:test-action"
	var storedAction ScheduledAction
	err = state.Get(key, &storedAction)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", *storedAction.Timezone)
	assert.NotNil(t, storedAction.StartTime)
	assert.NotNil(t, storedAction.EndTime)
	assert.NotNil(t, storedAction.ScalableTargetAction)
}

func TestPutScheduledAction_Update(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	// Create initial scheduled action
	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		Schedule:            strPtr("cron(0 10 * * ? *)"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Update with new schedule
	input.Schedule = strPtr("cron(0 12 * * ? *)")
	resp, err = service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify updated data
	key := "autoscaling:scheduled-action:dynamodb:table/test-table:dynamodb:table:ReadCapacityUnits:test-action"
	var storedAction ScheduledAction
	err = state.Get(key, &storedAction)
	require.NoError(t, err)
	assert.Equal(t, "cron(0 12 * * ? *)", *storedAction.Schedule)
}

func TestPutScheduledAction_MissingScheduledActionName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &PutScheduledActionRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ScheduledActionName")
}

func TestPutScheduledAction_MissingServiceNamespace(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ResourceId:          strPtr("table/test-table"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ServiceNamespace")
}

func TestPutScheduledAction_MissingResourceId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ScalableDimension:   ScalableDimension("dynamodb:table:ReadCapacityUnits"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ResourceId")
}

func TestPutScheduledAction_MissingScalableDimension(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	input := &PutScheduledActionRequest{
		ScheduledActionName: strPtr("test-action"),
		ServiceNamespace:    ServiceNamespace("dynamodb"),
		ResourceId:          strPtr("table/test-table"),
	}

	resp, err := service.putScheduledAction(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ScalableDimension")
}
