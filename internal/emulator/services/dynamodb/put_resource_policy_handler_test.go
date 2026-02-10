package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

func TestPutResourcePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  tableArn,
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Put a resource policy
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &tableArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	// Verify RevisionId is returned
	revisionId, ok := output["RevisionId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, revisionId)

	// Verify policy is stored in state
	policyKey := "dynamodb:resource-policy:" + tableArn
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get(policyKey, &storedPolicy))
	require.Equal(t, policy, storedPolicy["Policy"])
	require.Equal(t, revisionId, storedPolicy["RevisionId"])
}

func TestPutResourcePolicy_UpdateExistingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  tableArn,
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Put initial policy
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &tableArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp1, 200)

	var output1 map[string]interface{}
	require.NoError(t, json.Unmarshal(resp1.Body, &output1))
	revisionId1 := output1["RevisionId"].(string)

	// Update policy with expected revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:PutItem","Resource":"*"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &tableArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp2, 200)

	var output2 map[string]interface{}
	require.NoError(t, json.Unmarshal(resp2.Body, &output2))
	revisionId2 := output2["RevisionId"].(string)

	// Verify revision IDs are different
	require.NotEqual(t, revisionId1, revisionId2)

	// Verify new policy is stored
	policyKey := "dynamodb:resource-policy:" + tableArn
	var storedPolicy map[string]interface{}
	require.NoError(t, state.Get(policyKey, &storedPolicy))
	require.Equal(t, policy2, storedPolicy["Policy"])
	require.Equal(t, revisionId2, storedPolicy["RevisionId"])
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

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "ResourceArn is required")
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table"
	input := &PutResourcePolicyInput{
		ResourceArn: &tableArn,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "Policy is required")
}

func TestPutResourcePolicy_ResourceNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/nonexistent-table"
	policy := `{"Version":"2012-10-17","Statement":[]}`
	input := &PutResourcePolicyInput{
		ResourceArn: &tableArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "not found")
}

func TestPutResourcePolicy_InvalidRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  tableArn,
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Put initial policy
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:GetItem","Resource":"*"}]}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &tableArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp1, 200)

	// Try to update with wrong revision ID
	policy2 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"dynamodb:PutItem","Resource":"*"}]}`
	wrongRevisionId := "wrong-revision-id"
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &tableArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp2, 400)

	var output2 map[string]interface{}
	require.NoError(t, json.Unmarshal(resp2.Body, &output2))
	require.Contains(t, output2["message"], "revision")
}

func TestPutResourcePolicy_ExpectedRevisionIdNoPolicyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-east-1:000000000000:table/" + tableName
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  tableArn,
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Try to put policy with expected revision ID when no policy exists
	policy := `{"Version":"2012-10-17","Statement":[]}`
	revisionId := "some-revision-id"
	input := &PutResourcePolicyInput{
		ResourceArn:        &tableArn,
		Policy:             &policy,
		ExpectedRevisionId: &revisionId,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "does not exist")
}
