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
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:test-table", tableDesc))

	// Test putting a resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	// Should return a revision ID
	revisionId, ok := result["RevisionId"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify the policy was stored
	policyKey := "dynamodb:resource_policy:" + resourceArn
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get(policyKey, &storedPolicy))
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/non-existent-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", result["__type"])
}

func TestPutResourcePolicy_UpdateWithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:test-table", tableDesc))

	// Put initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	firstRevisionId := result["RevisionId"].(string)

	// Update with correct revision ID
	updatedPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"dynamodb:DeleteTable"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &updatedPolicy,
		ExpectedRevisionId: &firstRevisionId,
	}
	inputJSON2, _ := json.Marshal(input2)

	req2 := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON2,
		Action:  "PutResourcePolicy",
	}

	resp2, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)
	secondRevisionId := result2["RevisionId"].(string)
	assert.NotEqual(t, firstRevisionId, secondRevisionId)
}

func TestPutResourcePolicy_UpdateWithWrongRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:test-table", tableDesc))

	// Put initial policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Try to update with wrong revision ID
	wrongRevisionId := "wrong-revision-id"
	updatedPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"dynamodb:DeleteTable"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &updatedPolicy,
		ExpectedRevisionId: &wrongRevisionId,
	}
	inputJSON2, _ := json.Marshal(input2)

	req2 := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON2,
		Action:  "PutResourcePolicy",
	}

	resp2, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)
	assert.Equal(t, "PolicyRevisionMismatchException", result2["__type"])
}
