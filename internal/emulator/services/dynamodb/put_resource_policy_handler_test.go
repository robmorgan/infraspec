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
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Test PutResourcePolicy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	revisionId, ok := response["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify policy was stored
	policyKey := "dynamodb:resource-policy:" + resourceArn
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_UpdateWithCorrectRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Create initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var response1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &response1)
	require.NoError(t, err)
	revisionId1 := response1["RevisionId"].(string)

	// Update policy with correct revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:PutItem","Resource":"*"}]}`

	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var response2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &response2)
	require.NoError(t, err)
	revisionId2 := response2["RevisionId"].(string)
	assert.NotEqual(t, revisionId1, revisionId2)
}

func TestPutResourcePolicy_UpdateWithIncorrectRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Create initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with incorrect revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:PutItem","Resource":"*"}]}`
	wrongRevisionId := "wrong-revision-id"

	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var response2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &response2)
	require.NoError(t, err)
	assert.Contains(t, response2["message"], "not found")
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

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "ResourceArn is required")
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

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "Policy is required")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/non-existent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "not found")
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Test PutResourcePolicy with stream ARN
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName + "/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetRecords","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	revisionId, ok := response["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)
}
