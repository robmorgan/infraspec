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

	// Create some contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributor-insights:test-table-1", insight1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-2:test-index", insight2)
	require.NoError(t, err)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListContributorInsights_WithTableNameFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributor-insights:test-table-1", insight1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributor-insights:test-table-2", insight2)
	require.NoError(t, err)

	tableName := "test-table-1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "test-table-1", summary["TableName"])
}

func TestListContributorInsights_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 0)
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set("dynamodb:contributor-insights:test-table:"+string(rune(i)), insight)
		require.NoError(t, err)
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 2)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	assert.True(t, hasNextToken)
}
