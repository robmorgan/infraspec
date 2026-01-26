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
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	// Put a resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": "*", "Action": "dynamodb:*"}]}`

	reqBody := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s"}`, resourceArn, escapeJSON(policy))
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "RevisionId")
	revisionId, ok := responseBody["RevisionId"].(string)
	require.True(t, ok, "RevisionId should be a string")
	assert.NotEmpty(t, revisionId, "RevisionId should not be empty")

	// Verify policy was stored
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	assert.Equal(t, policy, storedPolicy["Policy"])
	assert.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17", "Statement": []}`
	reqBody := fmt.Sprintf(`{"Policy": "%s"}`, escapeJSON(policy))

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
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
	reqBody := fmt.Sprintf(`{"ResourceArn": "%s"}`, resourceArn)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", responseBody["__type"])
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version": "2012-10-17", "Statement": []}`

	reqBody := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s"}`, resourceArn, escapeJSON(policy))
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody),
		Action: "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", responseBody["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Put initial policy
	policy1 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	reqBody1 := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s"}`, resourceArn, escapeJSON(policy1))
	req1 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody1),
		Action: "PutResourcePolicy",
	}

	resp1, err := service.HandleRequest(context.Background(), req1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	var responseBody1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &responseBody1)
	require.NoError(t, err)
	revisionId1 := responseBody1["RevisionId"].(string)

	// Update policy with expected revision ID
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Deny"}]}`
	reqBody2 := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s", "ExpectedRevisionId": "%s"}`,
		resourceArn, escapeJSON(policy2), revisionId1)
	req2 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody2),
		Action: "PutResourcePolicy",
	}

	resp2, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var responseBody2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)
	revisionId2 := responseBody2["RevisionId"].(string)

	// Verify revision ID changed
	assert.NotEqual(t, revisionId1, revisionId2, "Revision ID should change on update")

	// Verify policy was updated
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	assert.Equal(t, policy2, storedPolicy["Policy"])
}

func TestPutResourcePolicy_WrongRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)

	// Put initial policy
	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	reqBody1 := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s"}`, resourceArn, escapeJSON(policy1))
	req1 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody1),
		Action: "PutResourcePolicy",
	}

	resp1, err := service.HandleRequest(context.Background(), req1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Deny"}]}`
	wrongRevisionId := "wrong-revision-id"
	reqBody2 := fmt.Sprintf(`{"ResourceArn": "%s", "Policy": "%s", "ExpectedRevisionId": "%s"}`,
		resourceArn, escapeJSON(policy2), wrongRevisionId)
	req2 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.PutResourcePolicy",
		},
		Body:   []byte(reqBody2),
		Action: "PutResourcePolicy",
	}

	resp2, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode)

	var responseBody2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)
	assert.Equal(t, "PolicyNotFoundException", responseBody2["__type"])
}

// Helper function to escape JSON strings for embedding in JSON
func escapeJSON(s string) string {
	result := ""
	for _, char := range s {
		if char == '"' {
			result += "\\\""
		} else if char == '\\' {
			result += "\\\\"
		} else {
			result += string(char)
		}
	}
	return result
}
