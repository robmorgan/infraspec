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
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:index1", tableName)
	insight2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "index1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List all contributor insights
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have two contributor insights")

	// Verify insights contain expected fields
	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok, "Each summary should be an object")
		assert.Contains(t, summaryMap, "TableName")
		assert.Contains(t, summaryMap, "ContributorInsightsStatus")
		assert.Equal(t, tableName, summaryMap["TableName"])
	}
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create insights for different tables
	table1Name := "table1"
	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table1Name)
	insight1Data := map[string]interface{}{
		"TableName":                 table1Name,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	table2Name := "table2"
	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table2Name)
	insight2Data := map[string]interface{}{
		"TableName":                 table2Name,
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List insights for table1 only
	input := &ListContributorInsightsInput{
		TableName: &table1Name,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Len(t, summaries, 1, "Should have only one contributor insight for table1")

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Name, summaryMap["TableName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple insights
	tableName := "test-table-paginated"

	for i := 1; i <= 5; i++ {
		insightKey := fmt.Sprintf("dynamodb:contributor-insights:%s:index%d", tableName, i)
		insightData := map[string]interface{}{
			"TableName":                 tableName,
			"IndexName":                 fmt.Sprintf("index%d", i),
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// List insights with limit
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have only 2 insights due to limit")

	// Should have NextToken for pagination
	assert.Contains(t, responseBody, "NextToken")
}

func TestListContributorInsights_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListContributorInsights should work with no parameters
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
}
