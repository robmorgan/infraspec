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

	// Create a test table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Action":"dynamodb:*","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.RevisionId)
	assert.NotEmpty(t, *output.RevisionId)

	// Verify the policy was stored
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, *output.RevisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17"}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["__type"], "ValidationException")
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

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["__type"], "ValidationException")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version":"2012-10-17"}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["__type"], "ResourceNotFoundException")
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put initial resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17","Statement":[]}`

	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var output1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &output1)
	require.NoError(t, err)
	firstRevisionId := *output1.RevisionId

	// Update the policy
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow"}]}`

	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var output2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)

	// Revision ID should be different
	assert.NotEqual(t, firstRevisionId, *output2.RevisionId)
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put initial resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17"}`

	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)

	var output1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &output1)
	require.NoError(t, err)
	currentRevisionId := *output1.RevisionId

	// Update with correct ExpectedRevisionId
	policy2 := `{"Version":"2012-10-17","Statement":[]}`

	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &currentRevisionId,
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
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put initial resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17"}`

	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong ExpectedRevisionId
	policy2 := `{"Version":"2012-10-17","Statement":[]}`
	wrongRevisionId := "wrong-revision-id"

	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["__type"], "ConditionalCheckFailedException")
}

func TestPutResourcePolicy_HandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17"}`

	requestBody := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(requestBody),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	assert.NotNil(t, output.RevisionId)
}
