package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	service, state := setupTestService(t)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Test putting a resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`
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
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	service, _ := setupTestService(t)

	// Test with missing ResourceArn
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResp)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", errorResp["__type"])
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	service, state := setupTestService(t)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Test with missing Policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResp)
	require.NoError(t, err)

	assert.Equal(t, "ValidationException", errorResp["__type"])
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	service, _ := setupTestService(t)

	// Test with non-existent table
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/non-existent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResp)
	require.NoError(t, err)

	assert.Equal(t, "ResourceNotFoundException", errorResp["__type"])
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	service, state := setupTestService(t)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Put initial policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem"}]}`
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
	revisionId1 := *output1.RevisionId

	// Update policy with expected revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var output2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)

	// Should have a different revision ID
	assert.NotEqual(t, revisionId1, *output2.RevisionId)
}

func TestPutResourcePolicy_InvalidRevisionId(t *testing.T) {
	service, state := setupTestService(t)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Put initial policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17","Statement":[]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	wrongRevisionId := "wrong-revision-id"
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var errorResp map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorResp)
	require.NoError(t, err)

	assert.Equal(t, "PolicyNotFoundException", errorResp["__type"])
}

func TestPutResourcePolicy_WithConfirmRemoveSelfAccess(t *testing.T) {
	service, state := setupTestService(t)

	// Create a test table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Test with ConfirmRemoveSelfResourceAccess
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"dynamodb:*"}]}`
	confirmRemove := true
	input := &PutResourcePolicyInput{
		ResourceArn:                     &resourceArn,
		Policy:                          &policy,
		ConfirmRemoveSelfResourceAccess: &confirmRemove,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.RevisionId)
}
