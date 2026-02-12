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

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte("{}"),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, responseBody.ContributorInsightsSummaries)
	assert.Empty(t, responseBody.ContributorInsightsSummaries, "Should have no contributor insights initially")
	assert.Nil(t, responseBody.NextToken)
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName := "test-table"

	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s:", tableName)
	insight1Data := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_OPERATIONS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:index1", tableName)
	insight2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "index1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "KEY_ONLY",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte("{}"),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ContributorInsightsSummaries, 2)

	// Check first insight
	assert.Equal(t, tableName, *responseBody.ContributorInsightsSummaries[0].TableName)
	assert.Equal(t, ContributorInsightsStatus("ENABLED"), responseBody.ContributorInsightsSummaries[0].ContributorInsightsStatus)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"

	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s:", table1)
	insight1Data := map[string]interface{}{
		"TableName":                 table1,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_OPERATIONS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:", table2)
	insight2Data := map[string]interface{}{
		"TableName":                 table2,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_OPERATIONS",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// Filter by table1
	reqBody := fmt.Sprintf(`{"TableName": "%s"}`, table1)
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte(reqBody),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should only return insights for table1
	assert.Len(t, responseBody.ContributorInsightsSummaries, 1)
	assert.Equal(t, table1, *responseBody.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test contributor insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("test-table-%d", i)
		insightKey := fmt.Sprintf("dynamodb:contributor-insights:%s:", tableName)
		insightData := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL_OPERATIONS",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// Request with MaxResults
	reqBody := `{"MaxResults": 2}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte(reqBody),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Should return only 2 results
	assert.Len(t, responseBody.ContributorInsightsSummaries, 2)
	// Should have NextToken for more results
	assert.NotNil(t, responseBody.NextToken)
}
