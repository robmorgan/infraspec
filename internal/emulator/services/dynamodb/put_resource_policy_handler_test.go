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

	// Policy must be JSON-encoded as a string value in the request body
	policyJSON, _ := json.Marshal(policy)
	body := fmt.Sprintf(`{"ResourceArn":"%s","Policy":%s}`, resourceArn, string(policyJSON))

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
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "RevisionId")
	revisionId, ok := responseBody["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)
}

func TestPutResourcePolicy_StoreAndRetrieve(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	putInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(policy),
	}

	putResp, err := service.putResourcePolicy(context.Background(), putInput)
	require.NoError(t, err)
	assert.Equal(t, 200, putResp.StatusCode)

	var putBody map[string]interface{}
	err = json.Unmarshal(putResp.Body, &putBody)
	require.NoError(t, err)
	revisionId := putBody["RevisionId"].(string)

	// Now retrieve the policy with GetResourcePolicy
	getInput := &GetResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
	}

	getResp, err := service.getResourcePolicy(context.Background(), getInput)
	require.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode)

	var getBody map[string]interface{}
	err = json.Unmarshal(getResp.Body, &getBody)
	require.NoError(t, err)

	assert.Equal(t, policy, getBody["Policy"])
	assert.Equal(t, resourceArn, getBody["ResourceArn"])
	assert.Equal(t, revisionId, getBody["RevisionId"])
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

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
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

func TestPutResourcePolicy_EmptyResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr(""),
		Policy:      strPtr(`{"Version":"2012-10-17","Statement":[]}`),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionId_NoExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[]}`),
		ExpectedRevisionId: strPtr("some-revision-id"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", responseBody["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionId_Mismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// First put a policy
	putInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(`{"Version":"2012-10-17","Statement":[]}`),
	}
	_, err := service.putResourcePolicy(context.Background(), putInput)
	require.NoError(t, err)

	// Try to update with wrong revision ID
	updateInput := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(`{"Version":"2012-10-17","Statement":[{"Effect":"Deny"}]}`),
		ExpectedRevisionId: strPtr("wrong-revision-id"),
	}

	resp, err := service.putResourcePolicy(context.Background(), updateInput)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "PolicyRevisionIdMismatchException", responseBody["__type"])
}

func TestPutResourcePolicy_UpdateWithCorrectRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	originalPolicy := `{"Version":"2012-10-17","Statement":[]}`

	// First put a policy
	putInput := &PutResourcePolicyInput{
		ResourceArn: strPtr(resourceArn),
		Policy:      strPtr(originalPolicy),
	}
	putResp, err := service.putResourcePolicy(context.Background(), putInput)
	require.NoError(t, err)

	var putBody map[string]interface{}
	err = json.Unmarshal(putResp.Body, &putBody)
	require.NoError(t, err)
	revisionId := putBody["RevisionId"].(string)

	// Update with correct revision ID
	newPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow"}]}`
	updateInput := &PutResourcePolicyInput{
		ResourceArn:        strPtr(resourceArn),
		Policy:             strPtr(newPolicy),
		ExpectedRevisionId: strPtr(revisionId),
	}

	updateResp, err := service.putResourcePolicy(context.Background(), updateInput)
	require.NoError(t, err)
	assert.Equal(t, 200, updateResp.StatusCode)

	var updateBody map[string]interface{}
	err = json.Unmarshal(updateResp.Body, &updateBody)
	require.NoError(t, err)
	newRevisionId := updateBody["RevisionId"].(string)
	assert.NotEmpty(t, newRevisionId)
	assert.NotEqual(t, revisionId, newRevisionId)
}
