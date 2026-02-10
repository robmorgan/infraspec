package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
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
	require.NoError(t, state.Set(tableKey, tableDesc))

	// Create some contributor insights configurations
	insights1 := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"LastUpdateDateTime":        float64(1234567890),
	}
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table", insights1))

	insights2 := map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	}
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table:test-index", insights2))

	// Test listing all contributor insights for the table
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Verify first summary
	summary1 := summaries[0].(map[string]interface{})
	require.Equal(t, "test-table", summary1["TableName"])
	require.Equal(t, "ENABLED", summary1["ContributorInsightsStatus"])

	// Verify second summary
	summary2 := summaries[1].(map[string]interface{})
	require.Equal(t, "test-table", summary2["TableName"])
	require.Equal(t, "DISABLED", summary2["ContributorInsightsStatus"])
	require.Equal(t, "test-index", summary2["IndexName"])
}

func TestListContributorInsights_AllTables(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights for multiple tables
	insights1 := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	}
	require.NoError(t, state.Set("dynamodb:contributor-insights:table1", insights1))

	insights2 := map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	}
	require.NoError(t, state.Set("dynamodb:contributor-insights:table2", insights2))

	// Test listing all contributor insights (no table filter)
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a test table but no contributor insights
	tableName := "test-table"
	tableKey := "dynamodb:table:" + tableName
	tableDesc := map[string]interface{}{
		"TableName":   tableName,
		"TableStatus": "ACTIVE",
	}
	require.NoError(t, state.Set(tableKey, tableDesc))

	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListContributorInsights_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "nonexistent-table"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 400)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Contains(t, output["message"], "not found")
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		insights := map[string]interface{}{
			"ContributorInsightsStatus": "ENABLED",
		}
		key := "dynamodb:contributor-insights:table" + string(rune('0'+i))
		require.NoError(t, state.Set(key, insights))
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &output))

	summaries, ok := output["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.LessOrEqual(t, len(summaries), 2)
}
