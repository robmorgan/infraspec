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

	// Create test data - contributor insights for a table
	tableName := "test-table"
	insight1 := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "test-gsi",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight3 := map[string]interface{}{
		"TableName":                 "other-table",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table:", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table:test-gsi", insight2))
	require.NoError(t, state.Set("dynamodb:contributor-insights:other-table:", insight3))

	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 3)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	tableName := "test-table"
	insight1 := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "other-table",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table:", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:other-table:", insight2))

	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	// Verify it's the correct table
	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName, summary["TableName"])
}

func TestListContributorInsights_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	for i := 0; i < 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "RULE_BASED",
		}
		require.NoError(t, state.Set("dynamodb:contributor-insights:test-table:"+string(rune(i)), insight))
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := output["NextToken"]
	assert.True(t, hasNextToken)
}
