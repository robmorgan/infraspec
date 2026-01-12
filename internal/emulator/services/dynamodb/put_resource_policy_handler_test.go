package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	err := state.Set("dynamodb:table:test-table", tableDesc)
	require.NoError(t, err)

	// Test putting a resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	// Should return a RevisionId
	revisionId, ok := responseData["RevisionId"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify the policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	err := state.Set("dynamodb:table:test-table", tableDesc)
	require.NoError(t, err)

	// Put an initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var responseData1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &responseData1)
	require.NoError(t, err)
	revisionId1 := responseData1["RevisionId"].(string)

	// Update the policy with a new one
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var responseData2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseData2)
	require.NoError(t, err)
	revisionId2 := responseData2["RevisionId"].(string)

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify the new policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
	assert.Equal(t, revisionId2, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	err := state.Set("dynamodb:table:test-table", tableDesc)
	require.NoError(t, err)

	// Put an initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	var responseData1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &responseData1)
	require.NoError(t, err)
	revisionId1 := responseData1["RevisionId"].(string)

	// Update with correct revision ID - should succeed
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestPutResourcePolicy_WithWrongExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:123456789012:table/test-table",
	}
	err := state.Set("dynamodb:table:test-table", tableDesc)
	require.NoError(t, err)

	// Put an initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID - should fail
	wrongRevisionId := "wrong-revision-id"
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorData)
	require.NoError(t, err)
	assert.Contains(t, errorData["__type"], "PolicyNotFoundException")
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Contains(t, errorData["__type"], "ValidationException")
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Contains(t, errorData["__type"], "ValidationException")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/non-existent-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Contains(t, errorData["__type"], "ResourceNotFoundException")
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with stream ARN (doesn't require table to exist in this simple implementation)
	resourceArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	revisionId, ok := responseData["RevisionId"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify the policy was stored with stream key
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resource-policy:stream:test-table/stream/2024-01-01T00:00:00.000", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy, storedPolicy["Policy"])
}
