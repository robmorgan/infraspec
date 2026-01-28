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
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": "arn:aws:iam::123456789012:user/test"},
			"Action": "dynamodb:GetItem",
			"Resource": "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
		}]
	}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response contains a revision ID
	assert.NotNil(t, responseBody.RevisionId)
	assert.NotEmpty(t, *responseBody.RevisionId)
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17"}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody["__type"], "ValidationException")
	assert.Contains(t, responseBody["message"], "ResourceArn is required")
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody["__type"], "ValidationException")
	assert.Contains(t, responseBody["message"], "Policy is required")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/non-existent-table"
	policy := `{"Version": "2012-10-17"}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody["__type"], "ResourceNotFoundException")
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version": "2012-10-17"}`

	// Put initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var responseBody1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &responseBody1)
	require.NoError(t, err)
	firstRevisionId := *responseBody1.RevisionId

	// Update policy
	newPolicy := `{"Version": "2012-10-17", "Statement": []}`
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &newPolicy,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var responseBody2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)

	// Verify revision ID changed
	assert.NotEqual(t, firstRevisionId, *responseBody2.RevisionId)
}

func TestPutResourcePolicy_ExpectedRevisionIdMismatch(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version": "2012-10-17"}`

	// Put initial policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong expected revision ID
	wrongRevisionId := "wrong-revision-id"
	newPolicy := `{"Version": "2012-10-17", "Statement": []}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody["__type"], "PolicyNotFoundException")
}

func TestPutResourcePolicy_ViaHandleRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Prepare request
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version": "2012-10-17"}`
	requestBody := fmt.Sprintf(`{
		"ResourceArn": "%s",
		"Policy": "%s"
	}`, resourceArn, policy)

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

	var responseBody PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.NotNil(t, responseBody.RevisionId)
}
