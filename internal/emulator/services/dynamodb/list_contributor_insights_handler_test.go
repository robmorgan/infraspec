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

	insights1Key := fmt.Sprintf("dynamodb:contributorinsights:%s", tableName)
	insights1Data := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insights1Key, insights1Data)
	require.NoError(t, err)

	insights2Key := fmt.Sprintf("dynamodb:contributorinsights:%s:index1", tableName)
	insights2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "index1",
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insights2Key, insights2Data)
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

	// Verify summaries contain expected fields
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

	// Create contributor insights for different tables
	table1Name := "table1"
	insights1Key := fmt.Sprintf("dynamodb:contributorinsights:%s", table1Name)
	insights1Data := map[string]interface{}{
		"TableName":                 table1Name,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insights1Key, insights1Data)
	require.NoError(t, err)

	table2Name := "table2"
	insights2Key := fmt.Sprintf("dynamodb:contributorinsights:%s", table2Name)
	insights2Data := map[string]interface{}{
		"TableName":                 table2Name,
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insights2Key, insights2Data)
	require.NoError(t, err)

	// List contributor insights for table1 only
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
	assert.Len(t, summaries, 1, "Should have only one contributor insights for table1")

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Name, summaryMap["TableName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	tableName := "test-table-paginated"

	for i := 1; i <= 5; i++ {
		insightsKey := fmt.Sprintf("dynamodb:contributorinsights:%s:index%d", tableName, i)
		insightsData := map[string]interface{}{
			"TableName":                 tableName,
			"IndexName":                 fmt.Sprintf("index%d", i),
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(insightsKey, insightsData)
		require.NoError(t, err)
	}

	// List contributor insights with limit
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
	assert.Len(t, summaries, 2, "Should have only 2 contributor insights due to limit")

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
