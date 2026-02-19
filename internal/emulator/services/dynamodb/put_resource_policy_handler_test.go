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

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

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

	assert.Contains(t, result, "RevisionId")
	revisionId, ok := result["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)
}

func TestPutResourcePolicy_PolicyStoredAndRetrievable(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	putInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}

	putResp, err := service.putResourcePolicy(context.Background(), putInput)
	require.NoError(t, err)
	assert.Equal(t, 200, putResp.StatusCode)

	// Retrieve the policy using GetResourcePolicy
	getInput := &GetResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
	}

	getResp, err := service.getResourcePolicy(context.Background(), getInput)
	require.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode)

	var getResult map[string]interface{}
	err = json.Unmarshal(getResp.Body, &getResult)
	require.NoError(t, err)

	assert.Equal(t, policy, getResult["Policy"])
	assert.Equal(t, resourceArn, getResult["ResourceArn"])
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
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/my-table"),
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

func TestPutResourcePolicy_WithExpectedRevisionId_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	// First, create initial policy
	firstInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}
	firstResp, err := service.putResourcePolicy(context.Background(), firstInput)
	require.NoError(t, err)
	assert.Equal(t, 200, firstResp.StatusCode)

	var firstResult map[string]interface{}
	err = json.Unmarshal(firstResp.Body, &firstResult)
	require.NoError(t, err)
	firstRevisionId := firstResult["RevisionId"].(string)

	// Update with correct expected revision ID
	updatedPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny"}]}`
	secondInput := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(updatedPolicy),
		ExpectedRevisionId: strPtr(firstRevisionId),
	}
	secondResp, err := service.putResourcePolicy(context.Background(), secondInput)
	require.NoError(t, err)
	assert.Equal(t, 200, secondResp.StatusCode)

	var secondResult map[string]interface{}
	err = json.Unmarshal(secondResp.Body, &secondResult)
	require.NoError(t, err)

	// New revision ID should differ from the first
	secondRevisionId := secondResult["RevisionId"].(string)
	assert.NotEqual(t, firstRevisionId, secondRevisionId)
}

func TestPutResourcePolicy_WithExpectedRevisionId_Conflict(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	// Create initial policy
	firstInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}
	_, err := service.putResourcePolicy(context.Background(), firstInput)
	require.NoError(t, err)

	// Attempt update with wrong revision ID
	secondInput := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[{"Effect":"Deny"}]}`),
		ExpectedRevisionId: strPtr("wrong-revision-id"),
	}
	secondResp, err := service.putResourcePolicy(context.Background(), secondInput)
	require.NoError(t, err)
	assert.Equal(t, 400, secondResp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(secondResp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "TransactionConflictException", result["__type"])
}

func TestPutResourcePolicy_WithExpectedRevisionId_NoPriorPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"

	// Attempt to put policy with expected revision ID but no existing policy
	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[]}`),
		ExpectedRevisionId: strPtr("some-revision-id"),
	}
	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", result["__type"])
}

func TestPutResourcePolicy_ViaHandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table"
	body := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"{\"Version\":\"2012-10-17\",\"Statement\":[]}"}`, resourceArn)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(body),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "RevisionId")
}
