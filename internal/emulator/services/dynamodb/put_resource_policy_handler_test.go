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

	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
				"Action": "dynamodb:*",
				"Resource": "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
			}
		]
	}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	revisionId, ok := result["RevisionId"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, revisionId)

	// Verify policy was stored in state
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resourcepolicy:arn:aws:dynamodb:us-east-1:000000000000:table/test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		Policy: strPtr(`{"Version": "2012-10-17"}`),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
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

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var result1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &result1)
	require.NoError(t, err)
	revisionId1 := result1["RevisionId"].(string)

	// Update the policy
	input2 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:      strPtr(policy2),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)
	revisionId2 := result2["RevisionId"].(string)

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)

	// Verify updated policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:resourcepolicy:arn:aws:dynamodb:us-east-1:000000000000:table/test-table", &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
	assert.Equal(t, revisionId2, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var result1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &result1)
	require.NoError(t, err)
	revisionId1 := result1["RevisionId"].(string)

	// Update with correct ExpectedRevisionId
	input2 := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:             strPtr(policy2),
		ExpectedRevisionId: strPtr(revisionId1),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestPutResourcePolicy_WithWrongExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`

	// Create initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Update with wrong ExpectedRevisionId
	input2 := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:             strPtr(policy2),
		ExpectedRevisionId: strPtr("wrong-revision-id"),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", result["__type"])
}

func TestPutResourcePolicy_ExpectedRevisionIdWithoutExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table"),
		Policy:             strPtr(`{"Version": "2012-10-17"}`),
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
