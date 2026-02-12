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

	// Create a table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	// Put resource policy
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`
	reqBody := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy)

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

	var responseBody PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, responseBody.RevisionId)
	assert.NotEmpty(t, *responseBody.RevisionId)
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Request without ResourceArn
	policy := `{"Version":"2012-10-17","Statement":[]}`
	reqBody := fmt.Sprintf(`{"Policy":"%s"}`, policy)

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

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["message"], "ResourceArn is required")
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Request without Policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	reqBody := fmt.Sprintf(`{"ResourceArn":"%s"}`, resourceArn)

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

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["message"], "Policy is required")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Request for non-existent table
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	reqBody := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy)

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

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["message"], "not found")
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`

	// Put initial policy
	reqBody1 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy1)
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

	var responseBody1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &responseBody1)
	require.NoError(t, err)
	revisionId1 := *responseBody1.RevisionId

	// Update policy
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`
	reqBody2 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy2)
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

	var responseBody2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &responseBody2)
	require.NoError(t, err)
	revisionId2 := *responseBody2.RevisionId

	// Revision IDs should be different
	assert.NotEqual(t, revisionId1, revisionId2)
}

func TestPutResourcePolicy_WithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[]}`

	// Put initial policy
	reqBody1 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy)
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

	var responseBody1 PutResourcePolicyOutput
	err = json.Unmarshal(resp1.Body, &responseBody1)
	require.NoError(t, err)
	revisionId := *responseBody1.RevisionId

	// Update with correct ExpectedRevisionId
	reqBody2 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s","ExpectedRevisionId":"%s"}`, resourceArn, policy, revisionId)
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
}

func TestPutResourcePolicy_WithWrongExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableData)
	require.NoError(t, err)

	resourceArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
	policy := `{"Version":"2012-10-17","Statement":[]}`

	// Put initial policy
	reqBody1 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s"}`, resourceArn, policy)
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

	// Update with wrong ExpectedRevisionId
	wrongRevisionId := "wrong-revision-id"
	reqBody2 := fmt.Sprintf(`{"ResourceArn":"%s","Policy":"%s","ExpectedRevisionId":"%s"}`, resourceArn, policy, wrongRevisionId)
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

	var errorBody map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorBody)
	require.NoError(t, err)
	assert.Contains(t, errorBody["message"], "revision ID")
}
