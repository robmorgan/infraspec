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
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableData))

	// Put resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	revisionId, ok := output["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify policy was stored
	policyKey := "dynamodb:policy:" + resourceArn
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get(policyKey, &storedPolicy))
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

	var errorOutput map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorOutput))
	assert.Equal(t, "ValidationException", errorOutput["__type"])
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

	var errorOutput map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorOutput))
	assert.Equal(t, "ValidationException", errorOutput["__type"])
}

func TestPutResourcePolicy_ResourceNotFound(t *testing.T) {
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

	var errorOutput map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorOutput))
	assert.Equal(t, "ResourceNotFoundException", errorOutput["__type"])
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableData))

	// Put initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var output1 map[string]interface{}
	require.NoError(t, json.Unmarshal(resp1.Body, &output1))
	revisionId := output1["RevisionId"].(string)

	// Update policy with correct revision ID
	updatedPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"dynamodb:DeleteItem","Resource":"*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &updatedPolicy,
		ExpectedRevisionId: &revisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var output2 map[string]interface{}
	require.NoError(t, json.Unmarshal(resp2.Body, &output2))
	newRevisionId := output2["RevisionId"].(string)
	assert.NotEqual(t, revisionId, newRevisionId)
}

func TestPutResourcePolicy_InvalidRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableData))

	// Put initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	wrongRevisionId := "wrong-revision-id"
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var errorOutput map[string]interface{}
	require.NoError(t, json.Unmarshal(resp2.Body, &errorOutput))
	assert.Equal(t, "PolicyNotFoundException", errorOutput["__type"])
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableData))

	// Put resource policy on stream
	streamArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &streamArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	assert.NotEmpty(t, output["RevisionId"])
}

func TestPutResourcePolicy_WithConfirmRemoveSelfResourceAccess(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableData))

	// Put resource policy with confirmation
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	confirmRemove := true
	input := &PutResourcePolicyInput{
		ResourceArn:                     &resourceArn,
		Policy:                          &policy,
		ConfirmRemoveSelfResourceAccess: &confirmRemove,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify the confirmation flag was stored
	policyKey := "dynamodb:policy:" + resourceArn
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get(policyKey, &storedPolicy))
	assert.Equal(t, true, storedPolicy["ConfirmRemoveSelfResourceAccess"])
}
