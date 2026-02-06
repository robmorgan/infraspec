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

	// Create a test table
	tableName := "test-table"
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Create contributor insights config
	insightsKey := "dynamodb:contributor-insights:" + tableName
	insightsConfig := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(insightsKey, insightsConfig)
	require.NoError(t, err)

	// Test ListContributorInsights
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(summaries), 1)

	if len(summaries) > 0 {
		summary := summaries[0].(map[string]interface{})
		assert.Equal(t, tableName, summary["TableName"])
		assert.Equal(t, "ENABLED", summary["ContributorInsightsStatus"])
	}
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test without any contributor insights
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListContributorInsights_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "non-existent-table"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "not found")
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights entries
	for i := 1; i <= 3; i++ {
		tableName := "test-table-" + string(rune('0'+i))
		insightsKey := "dynamodb:contributor-insights:" + tableName
		insightsConfig := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(insightsKey, insightsConfig)
		require.NoError(t, err)
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(summaries), 2)
}
