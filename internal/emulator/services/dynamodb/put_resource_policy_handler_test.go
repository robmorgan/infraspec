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

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify RevisionId is returned
	assert.Contains(t, responseBody, "RevisionId")
	revisionId, ok := responseBody["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)
}

func TestPutResourcePolicy_StoresPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	_, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)

	// Verify the policy was stored in state
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)

	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.NotEmpty(t, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[]}`
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	// Put first policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}
	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var body1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &body1)
	require.NoError(t, err)
	revisionId1 := body1["RevisionId"].(string)

	// Put second policy (update)
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}
	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var body2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &body2)
	require.NoError(t, err)
	revisionId2 := body2["RevisionId"].(string)

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify the stored policy is the updated one
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
}

func TestPutResourcePolicy_WithExpectedRevisionId_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	existingRevisionId := "existing-revision-123"

	// Pre-populate policy with a known revision ID
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	existingPolicyData := map[string]interface{}{
		"Policy":      `{"Version":"2012-10-17","Statement":[]}`,
		"ResourceArn": resourceArn,
		"RevisionId":  existingRevisionId,
	}
	err := state.Set(policyKey, existingPolicyData)
	require.NoError(t, err)

	// Update with correct expected revision ID
	newPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &existingRevisionId,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Contains(t, responseBody, "RevisionId")
}

func TestPutResourcePolicy_WithExpectedRevisionId_Mismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"

	// Pre-populate policy with a known revision ID
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	existingPolicyData := map[string]interface{}{
		"Policy":      `{"Version":"2012-10-17","Statement":[]}`,
		"ResourceArn": resourceArn,
		"RevisionId":  "actual-revision-id",
	}
	err := state.Set(policyKey, existingPolicyData)
	require.NoError(t, err)

	// Try to update with wrong revision ID
	wrongRevisionId := "wrong-revision-id"
	newPolicy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "TransactionCanceledException", responseBody["__type"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17","Statement":[]}`
	emptyArn := ""
	input := &PutResourcePolicyInput{
		ResourceArn: &emptyArn,
		Policy:      &policy,
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
	emptyPolicy := ""
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &emptyPolicy,
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

func TestPutResourcePolicy_NilPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      nil,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}

func TestPutResourcePolicy_ViaHandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(fmt.Sprintf(`{"ResourceArn":"%s","Policy":%q}`, resourceArn, policy)),
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
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	streamArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &streamArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Contains(t, responseBody, "RevisionId")

	// Verify the policy can be retrieved
	getInput := &GetResourcePolicyInput{
		ResourceArn: &streamArn,
	}
	getResp, err := service.getResourcePolicy(context.Background(), getInput)
	require.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode)

	var getBody map[string]interface{}
	err = json.Unmarshal(getResp.Body, &getBody)
	require.NoError(t, err)
	assert.Equal(t, policy, getBody["Policy"])
}
