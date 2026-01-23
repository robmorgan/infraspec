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

	// Create some test contributor insights
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
		"IndexName":                 "test-index",
	}

	state.Set("dynamodb:contributor-insights:test-table-1", insight1)
	state.Set("dynamodb:contributor-insights:test-table-2", insight2)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some test contributor insights
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	state.Set("dynamodb:contributor-insights:test-table-1", insight1)
	state.Set("dynamodb:contributor-insights:test-table-2", insight2)

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

	summaries, ok := responseData["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "test-table-1", summary["TableName"])
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListContributorInsights_WithMaxResults(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple test contributor insights
	for i := 0; i < 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
		}
		state.Set("dynamodb:contributor-insights:test-table:"+string(rune(i)), insight)
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	summaries, ok := responseData["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := responseData["NextToken"]
	assert.True(t, hasNextToken)
}
