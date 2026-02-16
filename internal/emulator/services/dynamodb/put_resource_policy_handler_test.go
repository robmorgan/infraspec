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
	tableName := "TestTable"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	policy := `{"Version": "2012-10-17", "Statement": []}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	revisionId, ok := response["RevisionId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, revisionId)
}

func TestPutResourcePolicy_MissingResourceArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	policy := `{"Version": "2012-10-17", "Statement": []}`
	input := &PutResourcePolicyInput{
		Policy: &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Equal(t, "ValidationException", errorResponse["__type"])
}

func TestPutResourcePolicy_MissingPolicy(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Equal(t, "ValidationException", errorResponse["__type"])
}

func TestPutResourcePolicy_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/NonExistentTable"
	policy := `{"Version": "2012-10-17", "Statement": []}`
	input := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy,
	}

	resp, err := service.putResourcePolicy(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorResponse)
	require.NoError(t, err)
	require.Equal(t, "ResourceNotFoundException", errorResponse["__type"])
}

func TestPutResourcePolicy_UpdateExisting(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "TestTable"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put initial resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	var response1 map[string]interface{}
	err = json.Unmarshal(resp1.Body, &response1)
	require.NoError(t, err)
	revisionId1 := response1["RevisionId"].(string)

	// Update the policy with the correct revision ID
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &revisionId1,
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

func TestPutResourcePolicy_WrongRevisionId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "TestTable"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Put initial resource policy
	resourceArn := "arn:aws:dynamodb:us-east-1:000000000000:table/TestTable"
	policy1 := `{"Version": "2012-10-17", "Statement": []}`
	input1 := &PutResourcePolicyInput{
		ResourceArn: &resourceArn,
		Policy:      &policy1,
	}

	resp1, err := service.putResourcePolicy(context.Background(), input1)
	require.NoError(t, err)
	require.Equal(t, 200, resp1.StatusCode)

	// Try to update with wrong revision ID
	policy2 := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow"}]}`
	wrongRevisionId := "wrong-revision-id"
	input2 := &PutResourcePolicyInput{
		ResourceArn:        &resourceArn,
		Policy:             &policy2,
		ExpectedRevisionId: &wrongRevisionId,
	}

	resp2, err := service.putResourcePolicy(context.Background(), input2)
	require.NoError(t, err)
	require.Equal(t, 400, resp2.StatusCode)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(resp2.Body, &errorResponse)
	require.NoError(t, err)
	require.Equal(t, "PolicyNotFoundException", errorResponse["__type"])
}
