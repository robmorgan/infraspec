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

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableDesc))

	// Put resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.NotNil(t, output.RevisionId)
	require.NotEmpty(t, *output.RevisionId)
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
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Contains(t, errorResponse["__type"], "ValidationException")
	require.Contains(t, errorResponse["message"], "ResourceArn is required")
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
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Contains(t, errorResponse["__type"], "ValidationException")
	require.Contains(t, errorResponse["message"], "Policy is required")
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Contains(t, errorResponse["__type"], "ResourceNotFoundException")
}

func TestPutResourcePolicy_UpdateWithExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableDesc))

	// Put initial resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	firstRevisionId := *output.RevisionId

	// Update policy with correct expected revision ID
	newPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &firstRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var output2 PutResourcePolicyOutput
	err = json.Unmarshal(resp2.Body, &output2)
	require.NoError(t, err)
	require.NotNil(t, output2.RevisionId)
	require.NotEqual(t, firstRevisionId, *output2.RevisionId)
}

func TestPutResourcePolicy_UpdateWithWrongExpectedRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableDesc))

	// Put initial resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Try to update with wrong expected revision ID
	wrongRevisionId := "wrong-revision-id"
	newPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:*","Resource":"*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &newPolicy,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 400, resp2.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorResponse)
	require.NoError(t, err)
	require.Contains(t, errorResponse["__type"], "PolicyRevisionMismatchException")
}

func TestPutResourcePolicy_WithConfirmRemoveSelfAccess(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	require.NoError(t, state.Set("dynamodb:table:"+tableName, tableDesc))

	// Put resource policy with confirm flag
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	confirmRemove := true
	input := &PutResourcePolicyInput{
		ResourceArn:                     &resourceArn,
		Policy:                          &policy,
		ConfirmRemoveSelfResourceAccess: &confirmRemove,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output PutResourcePolicyOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.NotNil(t, output.RevisionId)

	// Verify the policy was stored with the confirm flag
	policyKey := "dynamodb:resource-policy:" + resourceArn
	var storedPolicy map[string]interface{}
	err = state.Get(policyKey, &storedPolicy)
	require.NoError(t, err)
	require.Equal(t, true, storedPolicy["ConfirmRemoveSelfResourceAccess"])
}
