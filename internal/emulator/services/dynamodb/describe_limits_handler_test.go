package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeLimits_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.DescribeLimits",
		},
		Body:   []byte("{}"),
		Action: "DescribeLimits",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure contains all required fields
	assert.Contains(t, responseBody, "AccountMaxReadCapacityUnits")
	assert.Contains(t, responseBody, "AccountMaxWriteCapacityUnits")
	assert.Contains(t, responseBody, "TableMaxReadCapacityUnits")
	assert.Contains(t, responseBody, "TableMaxWriteCapacityUnits")

	// Verify values are reasonable
	assert.Equal(t, float64(80000), responseBody["AccountMaxReadCapacityUnits"])
	assert.Equal(t, float64(80000), responseBody["AccountMaxWriteCapacityUnits"])
	assert.Equal(t, float64(40000), responseBody["TableMaxReadCapacityUnits"])
	assert.Equal(t, float64(40000), responseBody["TableMaxWriteCapacityUnits"])
}

func TestDescribeLimits_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// DescribeLimits should work with no parameters (handler takes no input)
	resp, err := service.describeLimits(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify all expected fields are present
	assert.Contains(t, responseBody, "AccountMaxReadCapacityUnits")
	assert.Contains(t, responseBody, "AccountMaxWriteCapacityUnits")
	assert.Contains(t, responseBody, "TableMaxReadCapacityUnits")
	assert.Contains(t, responseBody, "TableMaxWriteCapacityUnits")
}

func TestDescribeLimits_EmptyRequest(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with empty JSON body
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.DescribeLimits",
		},
		Body:   []byte("{}"),
		Action: "DescribeLimits",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should still return valid limits
	assert.NotEmpty(t, responseBody)
	assert.Contains(t, responseBody, "AccountMaxReadCapacityUnits")
}
