package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s", tableName),
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Create a valid policy document
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
			"Action": "dynamodb:GetItem",
			"Resource": "arn:aws:dynamodb:us-east-1:123456789012:table/test-table"
		}]
	}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	revisionId, ok := response["RevisionId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, revisionId)

	// Verify policy was stored
	var storedPolicy map[string]interface{}
	err = state.Get("dynamodb:policy:arn:aws:dynamodb:us-east-1:123456789012:table/test-table", &storedPolicy)
	require.NoError(t, err)
	require.Equal(t, policy, storedPolicy["Policy"])
	require.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17"}`

	input := &PutResourcePolicyInput{
		Policy: strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	require.Equal(t, "ValidationException", response["__type"])
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	require.Equal(t, "ValidationException", response["__type"])
}

func TestPutResourcePolicy_InvalidPolicyJSON(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:      strPtr("not valid json"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	require.Equal(t, "ValidationException", response["__type"])
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17"}`

	input := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/nonexistent-table"),
		Policy:      strPtr(policy),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	require.Equal(t, "ResourceNotFoundException", response["__type"])
}

func TestPutResourcePolicy_WithExpectedRevisionId_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Put initial policy
	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	var response1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &response1)
	require.NoError(t, err)
	revisionId1 := response1["RevisionId"].(string)

	// Update policy with correct ExpectedRevisionId
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:             strPtr(policy2),
		ExpectedRevisionId: strPtr(revisionId1),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var response2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &response2)
	require.NoError(t, err)
	revisionId2 := response2["RevisionId"].(string)
	require.NotEqual(t, revisionId1, revisionId2)
}

func TestPutResourcePolicy_WithExpectedRevisionId_Mismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Put initial policy
	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:      strPtr(policy1),
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	// Try to update with incorrect ExpectedRevisionId
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:             strPtr(policy2),
		ExpectedRevisionId: strPtr("wrong-revision-id"),
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 409, resp2.StatusCode)

	var response2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &response2)
	require.NoError(t, err)
	require.Equal(t, "PolicyRevisionIdMismatchException", response2["__type"])
}

func TestPutResourcePolicy_WithExpectedRevisionId_PolicyNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
	}
	require.NoError(t, state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc))

	// Try to put policy with ExpectedRevisionId when no policy exists
	policy := `{"Version": "2012-10-17", "Statement": []}`
	input := &PutResourcePolicyInput{
		ResourceArn:        strPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		Policy:             strPtr(policy),
		ExpectedRevisionId: strPtr("some-revision-id"),
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	require.Equal(t, "PolicyNotFoundException", response["__type"])
}
