package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/testing/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.RevisionId)
	assert.NotEmpty(t, *output.RevisionId)

	// Verify policy was stored
	policyKey := "dynamodb:resource-policy:" + resourceArn
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
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
	testhelpers.AssertResponseStatus(t, resp, 400)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorResponse["__type"])
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
	testhelpers.AssertResponseStatus(t, resp, 400)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorResponse["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}
	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp1, 200)

	var output1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &output1)
	require.NoError(t, err)
	revisionId1 := *output1.RevisionId

	// Update policy
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}
	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp2, 200)

	var output2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)
	revisionId2 := *output2.RevisionId

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify policy was updated
	policyKey := "dynamodb:resource-policy:" + resourceArn
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}
	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	var output1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &output1)
	require.NoError(t, err)
	revisionId := *output1.RevisionId

	// Update with correct ExpectedRevisionId
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId,
	}
	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp2, 200)
}

func TestPutResourcePolicy_WrongExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}
	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)

	// Update with incorrect ExpectedRevisionId
	wrongRevisionId := "wrong-revision-id"
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}
	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp2, 400)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "PolicyRevisionMismatchException", errorResponse["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionIdButNoPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	expectedRevisionId := "some-revision-id"

	input := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy,
		ExpectedRevisionId: &expectedRevisionId,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", errorResponse["__type"])
}
