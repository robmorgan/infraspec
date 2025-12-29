package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeEndpoints_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.DescribeEndpoints",
		},
		Body:   []byte("{}"),
		Action: "DescribeEndpoints",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "Endpoints")
	endpoints, ok := responseBody["Endpoints"].([]interface{})
	require.True(t, ok, "Endpoints should be an array")
	require.Len(t, endpoints, 1, "Should have exactly one endpoint")

	endpoint, ok := endpoints[0].(map[string]interface{})
	require.True(t, ok, "Endpoint should be an object")
	assert.Equal(t, "localhost:3687", endpoint["Address"])
	assert.Equal(t, float64(1440), endpoint["CachePeriodInMinutes"])
}

func TestDescribeEndpoints_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// DescribeEndpoints should work with no parameters (handler takes no input)
	resp, err := service.describeEndpoints(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Contains(t, responseBody, "Endpoints")
}
