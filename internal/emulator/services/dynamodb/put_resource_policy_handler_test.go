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

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	// Put resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "RevisionId")
	revisionId, ok := responseBody["RevisionId"].(string)
	require.True(t, ok, "RevisionId should be a string")
	assert.NotEmpty(t, revisionId, "RevisionId should not be empty")

	// Verify policy was stored
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
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
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "ResourceArn is required")
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
	assert.Contains(t, responseBody["message"], "Policy is required")
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ResourceNotFoundException", responseBody["__type"])
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Put initial policy
	policy1 := `{"Version":"2012-10-17","Statement":[]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var responseBody1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &responseBody1)
	require.NoError(t, err)
	revisionId1 := responseBody1["RevisionId"].(string)

	// Update policy with matching revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"*","Resource":"*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var responseBody2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)
	revisionId2 := responseBody2["RevisionId"].(string)

	// Verify new revision ID is different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify updated policy was stored
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
	assert.Equal(t, revisionId2, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_RevisionMismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Put initial policy
	policy1 := `{"Version":"2012-10-17","Statement":[]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"*","Resource":"*"}]}`
	wrongRevisionId := "wrong-revision-id"
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var responseBody2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)

	assert.Equal(t, "PolicyRevisionMismatch", responseBody2["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionIdWithoutExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[]}`
	revisionId := "some-revision-id"

	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy,
		ExpectedRevisionId: &revisionId,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "PolicyNotFoundException", responseBody["__type"])
}
