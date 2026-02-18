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

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// RevisionId should be returned
	assert.Contains(t, result, "RevisionId")
	assert.NotEmpty(t, result["RevisionId"])
}

func TestPutResourcePolicy_PolicyStoredCorrectly(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify policy is stored in state
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)

	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.NotEmpty(t, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[]}`
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"*","Resource":"*"}]}`

	// Put initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var result1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &result1)
	require.NoError(t, err)
	revisionId1, _ := result1["RevisionId"].(string)

	// Update policy without expected revision ID (should succeed)
	input2 := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy2),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)
	revisionId2, _ := result2["RevisionId"].(string)

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)
}

func TestPutResourcePolicy_WithExpectedRevisionId_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Set up an existing policy with a known revision ID
	existingRevisionId := "existing-revision-123"
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	err := state.Set(policyKey, map[string]interface{}{
		"Policy":      `{"Version":"2012-10-17","Statement":[]}`,
		"ResourceArn": resourceArn,
		"RevisionId":  existingRevisionId,
	})
	require.NoError(t, err)

	// Update with correct expected revision ID
	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"*","Resource":"*"}]}`),
		ExpectedRevisionId: strPtr(existingRevisionId),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// Should have a new revision ID
	newRevisionId, _ := result["RevisionId"].(string)
	assert.NotEmpty(t, newRevisionId)
	assert.NotEqual(t, existingRevisionId, newRevisionId)
}

func TestPutResourcePolicy_WithExpectedRevisionId_Mismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Set up an existing policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	err := state.Set(policyKey, map[string]interface{}{
		"Policy":      `{"Version":"2012-10-17","Statement":[]}`,
		"ResourceArn": resourceArn,
		"RevisionId":  "current-revision",
	})
	require.NoError(t, err)

	// Try to update with wrong expected revision ID
	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[]}`),
		ExpectedRevisionId: strPtr("wrong-revision"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, "TransactionConflictException", result["__type"])
}

func TestPutResourcePolicy_WithExpectedRevisionId_NoPolicyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Try to update with expected revision ID but no existing policy
	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[]}`),
		ExpectedRevisionId: strPtr("some-revision"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, "PolicyNotFoundException", result["__type"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		Policy: strPtr(`{"Version":"2012-10-17","Statement":[]}`),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "ResourceArn is required")
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "Policy is required")
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Policies can also be attached to streams
	streamArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr(streamArn),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "RevisionId")
	assert.NotEmpty(t, result["RevisionId"])
}
