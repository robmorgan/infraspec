package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create a table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:test-table", tableData))

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": "arn:aws:iam::123456789012:user/testuser"},
			"Action": "dynamodb:GetItem",
			"Resource": "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
		}]
	}`

	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.NotNil(t, output.RevisionId)
	require.NotEmpty(t, *output.RevisionId)

	// Verify the policy was stored
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy))
	require.Equal(t, resourceArn, storedPolicy["ResourceArn"])
	require.Equal(t, policy, storedPolicy["Policy"])
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17"}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorResp))
	require.Equal(t, "ValidationException", errorResp["__type"])
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorResp))
	require.Equal(t, "ValidationException", errorResp["__type"])
}

func TestPutResourcePolicy_ResourceNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version": "2012-10-17"}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "PutResourcePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &errorResp))
	require.Equal(t, "ResourceNotFoundException", errorResp["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create a table
	tableName := "test-table"
	tableData := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:test-table", tableData))

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy1 := `{"Version": "2012-10-17", "Statement": []}`

	// First, create a policy
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}
	inputJSON1, _ := json.Marshal(input1)
	req1 := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON1,
		Action:  "PutResourcePolicy",
	}

	resp1, err := service.HandleRequest(context.Background(), req1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	var output1 PutResourcePolicyOutput
	require.NoError(t, json.Unmarshal(resp1.Body, &output1))
	firstRevisionId := *output1.RevisionId

	// Now, update the policy
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy2,
	}
	inputJSON2, _ := json.Marshal(input2)
	req2 := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON2,
		Action:  "PutResourcePolicy",
	}

	resp2, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var output2 PutResourcePolicyOutput
	require.NoError(t, json.Unmarshal(resp2.Body, &output2))
	secondRevisionId := *output2.RevisionId

	// Verify the revision IDs are different
	require.NotEqual(t, firstRevisionId, secondRevisionId)

	// Verify the updated policy was stored
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get("dynamodb:resource-policy:table:test-table", &storedPolicy))
	require.Equal(t, policy2, storedPolicy["Policy"])
	require.Equal(t, secondRevisionId, storedPolicy["RevisionId"])
}
