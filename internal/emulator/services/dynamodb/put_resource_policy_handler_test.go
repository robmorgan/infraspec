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

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	state.Set("dynamodb:table:test-table", tableDesc)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
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
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:policy:arn:aws:dynamodb:us-east-1:000000000000:table/test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "ResourceArn is required")
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "Policy is required")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

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
	assert.Equal(t, "ResourceNotFoundException", errorData["__type"])
}

func TestPutResourcePolicy_InvalidResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "invalid-arn"
	policy := `{"Version":"2012-10-17","Statement":[]}`

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
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "Invalid ResourceArn format")
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	state.Set("dynamodb:table:test-table", tableDesc)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`

	// Create initial policy
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

	// Update the policy
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var responseData2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseData2)
	require.NoError(t, err)
	revisionId2 := responseData2["RevisionId"].(string)

	// Verify new revision ID is different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify updated policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:policy:arn:aws:dynamodb:us-east-1:000000000000:table/test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
	assert.Equal(t, revisionId2, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_ExpectedRevisionIdMismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	state.Set("dynamodb:table:test-table", tableDesc)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	wrongRevisionId := "wrong-revision-id"
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny"}]}`
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
	assert.Equal(t, "PolicyNotFoundException", errorData["__type"])
}
