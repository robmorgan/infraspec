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
	revisionId, ok := responseBody["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)
}

func TestPutResourcePolicy_StoresPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/policy-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify the policy was stored by retrieving it
	getInput := &GetResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	getResp, err := service.getResourcePolicy(context.Background(), getInput)
	require.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode)

	var getBody map[string]interface{}
	err = json.Unmarshal(getResp.Body, &getBody)
	require.NoError(t, err)

	assert.Equal(t, policy, getBody["Policy"])
	assert.Equal(t, resourceArn, getBody["ResourceArn"])
	assert.Contains(t, getBody, "RevisionId")
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17"}`
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

func TestPutResourcePolicy_NilResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version":"2012-10-17"}`
	input := &PutResourcePolicyInput{
		ResourceArn: nil,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", responseBody["__type"])
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

func TestPutResourcePolicy_ExpectedRevisionId_Match(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/revision-table"

	// Seed an existing policy with a known revision ID
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	state.Set(policyKey, map[string]interface{}{
		"Policy":     `{"Version":"2012-10-17","Statement":[]}`,
		"RevisionId": "rev-123",
	})

	newPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"*","Resource":"*"}]}`
	expectedRevision := "rev-123"
	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &expectedRevision,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// New revision should differ from old
	assert.Contains(t, responseBody, "RevisionId")
	assert.NotEqual(t, "rev-123", responseBody["RevisionId"])
}

func TestPutResourcePolicy_ExpectedRevisionId_Mismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/mismatch-table"

	// Seed an existing policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	state.Set(policyKey, map[string]interface{}{
		"Policy":     `{"Version":"2012-10-17"}`,
		"RevisionId": "rev-abc",
	})

	policy := `{"Version":"2012-10-17","Statement":[]}`
	wrongRevision := "rev-wrong"
	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy,
		ExpectedRevisionId: &wrongRevision,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ConditionalCheckFailedException", responseBody["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionId_NoPriorPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/no-prior-table"
	policy := `{"Version":"2012-10-17"}`
	expectedRevision := "rev-nonexistent"

	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy,
		ExpectedRevisionId: &expectedRevision,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ConditionalCheckFailedException", responseBody["__type"])
}

func TestPutResourcePolicy_StreamArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	streamArn := "arn:aws:dynamodb:us-east-1:000000000000:table/my-table/stream/2024-01-01T00:00:00.000"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetRecords","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &streamArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify it can be retrieved
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
	assert.Equal(t, streamArn, getBody["ResourceArn"])
}
