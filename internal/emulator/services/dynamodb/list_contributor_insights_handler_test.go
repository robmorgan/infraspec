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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Empty(t, summaries, "Should have no contributor insights initially")
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights configs
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"

	// Table 1 contributor insights
	insights1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName1)
	insights1Data := map[string]interface{}{
		"TableName":                 tableName1,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insights1Key, insights1Data)
	require.NoError(t, err)

	// Table 1 index contributor insights
	insights2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:gsi-1", tableName1)
	insights2Data := map[string]interface{}{
		"TableName":                 tableName1,
		"IndexName":                 "gsi-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(insights2Key, insights2Data)
	require.NoError(t, err)

	// Table 2 contributor insights
	insights3Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName2)
	insights3Data := map[string]interface{}{
		"TableName":                 tableName2,
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insights3Key, insights3Data)
	require.NoError(t, err)

	// Test without filter
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 3, "Should have 3 contributor insights")
}

func TestListContributorInsights_WithTableFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights configs
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"

	// Table 1 contributor insights
	insights1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName1)
	insights1Data := map[string]interface{}{
		"TableName":                 tableName1,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insights1Key, insights1Data)
	require.NoError(t, err)

	// Table 1 index contributor insights
	insights2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:gsi-1", tableName1)
	insights2Data := map[string]interface{}{
		"TableName":                 tableName1,
		"IndexName":                 "gsi-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(insights2Key, insights2Data)
	require.NoError(t, err)

	// Table 2 contributor insights
	insights3Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName2)
	insights3Data := map[string]interface{}{
		"TableName":                 tableName2,
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insights3Key, insights3Data)
	require.NoError(t, err)

	// Test with TableName filter
	reqBody := fmt.Sprintf(`{"TableName": "%s"}`, tableName1)
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 contributor insights for table-1")

	// Verify all summaries are for table-1
	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})
		assert.Equal(t, tableName1, summaryMap["TableName"])
	}
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights configs
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("test-table-%d", i)
		insightsKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
		insightsData := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(insightsKey, insightsData)
		require.NoError(t, err)
	}

	// First page with MaxResults=2
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte(`{"MaxResults": 2}`),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in first page")

	// Verify NextToken is present
	nextToken, hasNext := responseBody["NextToken"].(string)
	assert.True(t, hasNext, "Should have NextToken for more results")

	// Second page using NextToken
	reqBody := fmt.Sprintf(`{"MaxResults": 2, "NextToken": "%s"}`, nextToken)
	req.Body = []byte(reqBody)

	resp, err = service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 summaries in second page")
}
