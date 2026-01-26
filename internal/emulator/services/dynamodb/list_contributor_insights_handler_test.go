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

	// Create test contributor insights
	tableName := "test-table"

	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	insight1Data := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_EVENTS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:index1", tableName)
	insight2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "index1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_EVENTS",
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Equal(t, 2, len(summaries), "Should have two contributor insights")
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights for different tables
	table1 := "test-table-1"
	table2 := "test-table-2"

	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table1)
	insight1Data := map[string]interface{}{
		"TableName":                 table1,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_EVENTS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table2)
	insight2Data := map[string]interface{}{
		"TableName":                 table2,
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL_EVENTS",
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Equal(t, 1, len(summaries), "Should have one contributor insight for table1")

	// Verify it's the correct table
	if len(summaries) > 0 {
		summary := summaries[0].(map[string]interface{})
		assert.Equal(t, table1, summary["TableName"])
	}
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test contributor insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("test-table-%d", i)
		insightKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
		insightData := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL_EVENTS",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// Request with MaxResults = 2
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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify pagination
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Equal(t, 2, len(summaries), "Should return only 2 results due to MaxResults")

	// Should have a NextToken since there are more results
	assert.Contains(t, responseBody, "NextToken", "Should have NextToken for pagination")
}
