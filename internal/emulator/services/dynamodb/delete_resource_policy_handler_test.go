package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a resource policy first
	policyDocument := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []interface{}{
			map[string]interface{}{
				"Effect":   "Allow",
				"Action":   "dynamodb:GetItem",
				"Resource": "*",
			},
		},
	}
	state.Set("dynamodb:resource-policy:test-table", policyDocument)

	input := &DeleteResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
	}

	resp, err := service.deleteResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// Response should contain RevisionId
	_, hasRevisionId := result["RevisionId"]
	assert.True(t, hasRevisionId)

	// Verify policy was deleted from state
	assert.False(t, state.Exists("dynamodb:resource-policy:test-table"))
}

func TestDeleteResourcePolicy_Idempotent(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// No policy exists, but delete should still succeed (idempotent)
	input := &DeleteResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
	}

	resp, err := service.deleteResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// Response should still contain RevisionId
	_, hasRevisionId := result["RevisionId"]
	assert.True(t, hasRevisionId)
}

func TestDeleteResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DeleteResourcePolicyInput{}

	resp, err := service.deleteResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "ResourceArn is required")
}

func TestDeleteResourcePolicy_InvalidResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DeleteResourcePolicyInput{
		ResourceArn: strPtr("invalid-arn"),
	}

	resp, err := service.deleteResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "Invalid ResourceArn format")
}
