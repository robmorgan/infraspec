package applicationautoscaling

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPredictiveScalingForecast_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		PolicyName:        strPtr("test-policy"),
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.1", resp.Headers["Content-Type"])

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	// Verify response structure
	// Note: LoadForecast may be omitted when empty due to json omitempty
	// CapacityForecast and UpdateTime should always be present
	_, hasCapacityForecast := responseData["CapacityForecast"]
	assert.True(t, hasCapacityForecast, "CapacityForecast should be present")

	_, hasUpdateTime := responseData["UpdateTime"]
	assert.True(t, hasUpdateTime, "UpdateTime should be present")
}

func TestGetPredictiveScalingForecast_MissingServiceNamespace(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		PolicyName:        strPtr("test-policy"),
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ServiceNamespace")
}

func TestGetPredictiveScalingForecast_MissingResourceId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		PolicyName:        strPtr("test-policy"),
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ResourceId")
}

func TestGetPredictiveScalingForecast_MissingScalableDimension(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace: ServiceNamespace("dynamodb"),
		ResourceId:       strPtr("table/test-table"),
		PolicyName:       strPtr("test-policy"),
		StartTime:        &startTime,
		EndTime:          &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ScalableDimension")
}

func TestGetPredictiveScalingForecast_MissingPolicyName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "PolicyName")
}

func TestGetPredictiveScalingForecast_MissingStartTime(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		PolicyName:        strPtr("test-policy"),
		EndTime:           &endTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "StartTime")
}

func TestGetPredictiveScalingForecast_MissingEndTime(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewApplicationAutoScalingService(state, validator)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	input := &GetPredictiveScalingForecastRequest{
		ServiceNamespace:  ServiceNamespace("dynamodb"),
		ResourceId:        strPtr("table/test-table"),
		ScalableDimension: ScalableDimension("dynamodb:table:ReadCapacityUnits"),
		PolicyName:        strPtr("test-policy"),
		StartTime:         &startTime,
	}

	resp, err := service.getPredictiveScalingForecast(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "EndTime")
}
