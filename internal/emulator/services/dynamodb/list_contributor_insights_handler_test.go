package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContributorInsights_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights configs
	config1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	config2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
	}
	config3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributor-insights:test-table-1", config1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-1:test-index", config2)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-2", config3)
	require.NoError(t, err)

	// Test list all contributor insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ContributorInsightsSummaries"].([]interface{})
	assert.Equal(t, 3, len(summaries))
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights configs
	config1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	config2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
	}
	config3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributor-insights:test-table-1", config1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-1:test-index", config2)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-2", config3)
	require.NoError(t, err)

	// Test filter by table name
	tableName := "test-table-1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ContributorInsightsSummaries"].([]interface{})
	assert.Equal(t, 2, len(summaries))

	// Verify both summaries are for test-table-1
	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})
		assert.Equal(t, "test-table-1", summaryMap["TableName"])
	}
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no contributor insights configured
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ContributorInsightsSummaries"].([]interface{})
	assert.Equal(t, 0, len(summaries))
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test contributor insights configs
	for i := 1; i <= 5; i++ {
		config := map[string]interface{}{
			"TableName":                 "test-table-1",
			"ContributorInsightsStatus": "ENABLED",
		}
		key := "dynamodb:contributor-insights:test-table-1"
		if i > 1 {
			config["IndexName"] = "index-" + string(rune('0'+i))
			key = key + ":index-" + string(rune('0'+i))
		}
		err := state.Set(key, config)
		require.NoError(t, err)
	}

	// Test with MaxResults limit
	maxResults := int32(3)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries := responseData["ContributorInsightsSummaries"].([]interface{})
	assert.Equal(t, 3, len(summaries))

	// Should have NextToken since there are more results
	_, hasNextToken := responseData["NextToken"]
	assert.True(t, hasNextToken)
}
