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

	// Create a table first
	tableName := "test-table"
	tableKey := "dynamodb:table:test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	revisionId, ok := response["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
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
	require.NotNil(t, resp)
	assert.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", response["__type"])
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
	require.NotNil(t, resp)
	assert.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", response["__type"])
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
	require.NotNil(t, resp)
	assert.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", response["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table
	tableName := "test-table"
	tableKey := "dynamodb:table:test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`

	// Put initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.NotNil(t, resp1)
	assert.Equal(t, 200, resp1.StatusCode)

	var response1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &response1)
	require.NoError(t, err)
	revisionId1 := response1["RevisionId"].(string)

	// Update policy with new version
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.NotNil(t, resp2)
	assert.Equal(t, 200, resp2.StatusCode)

	var response2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &response2)
	require.NoError(t, err)
	revisionId2 := response2["RevisionId"].(string)

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify updated policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table
	tableName := "test-table"
	tableKey := "dynamodb:table:test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[]}`

	// Put initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	var response1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &response1)
	require.NoError(t, err)
	revisionId1 := response1["RevisionId"].(string)

	// Try to update with correct ExpectedRevisionId
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.NotNil(t, resp2)
	assert.Equal(t, 200, resp2.StatusCode)

	// Try to update with incorrect ExpectedRevisionId
	wrongRevisionId := "wrong-revision-id"
	input3 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp3, err := service.putResourcePolicy(context.Background(), input3)
	require.NoError(t, err)
	require.NotNil(t, resp3)
	assert.Equal(t, 400, resp3.StatusCode)

	var response3 map[string]interface{}
	err = json.Unmarshal(resp3.Body, &response3)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", response3["__type"])
}
