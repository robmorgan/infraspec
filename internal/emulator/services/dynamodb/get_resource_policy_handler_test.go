package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func TestGetResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no policy attached
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.GetResourcePolicy",
		},
		Body:   []byte(fmt.Sprintf(`{"ResourceArn":"%s"}`, resourceArn)),
		Action: "GetResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, resourceArn, responseBody["ResourceArn"])
	// Policy should not be present if no policy is attached
	assert.NotContains(t, responseBody, "Policy")
}

func TestGetResourcePolicy_WithPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a resource with a policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-with-policy"
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	policyDocument := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	policyData := map[string]interface{}{
		"Policy":     policyDocument,
		"RevisionId": "12345",
	}
	err := state.Set(policyKey, policyData)
	require.NoError(t, err)

	// Test retrieving the policy
	input := &GetResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	resp, err := service.getResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, resourceArn, responseBody["ResourceArn"])
	assert.Equal(t, policyDocument, responseBody["Policy"])
	assert.Equal(t, "12345", responseBody["RevisionId"])
}

func TestGetResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with a stream ARN
	streamArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/stream/2024-01-01T00:00:00.000"
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", streamArn)
	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	policyData := map[string]interface{}{
		"Policy": policyDocument,
	}
	err := state.Set(policyKey, policyData)
	require.NoError(t, err)

	// Test retrieving the stream policy
	input := &GetResourcePolicyInput{
		ResourceArn: &streamArn,
	}

	resp, err := service.getResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, streamArn, responseBody["ResourceArn"])
	assert.Equal(t, policyDocument, responseBody["Policy"])
}

func TestGetResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	emptyArn := ""
	input := &GetResourcePolicyInput{
		ResourceArn: &emptyArn,
	}

	resp, err := service.getResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "ResourceArn is required")
}

func TestGetResourcePolicy_NilResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &GetResourcePolicyInput{
		ResourceArn: nil,
	}

	resp, err := service.getResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
}
