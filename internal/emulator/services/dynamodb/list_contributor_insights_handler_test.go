package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Empty(t, output.ContributorInsightsSummaries)
	require.Nil(t, output.NextToken)
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"
	indexName := "test-index"

	insight1 := map[string]interface{}{
		"TableName":                 tableName1,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	insight2 := map[string]interface{}{
		"TableName":                 tableName1,
		"IndexName":                 indexName,
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	insight3 := map[string]interface{}{
		"TableName":                 tableName2,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "TOP_N",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1:test-index", insight2))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2", insight3))

	// List all contributor insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ContributorInsightsSummaries, 3)
}

func TestListContributorInsights_FilterByTable(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"

	insight1 := map[string]interface{}{
		"TableName":                 tableName1,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	insight2 := map[string]interface{}{
		"TableName":                 tableName2,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2", insight2))

	// Filter by table name
	input := &ListContributorInsightsInput{
		TableName: &tableName1,
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ContributorInsightsSummaries, 1)
	require.Equal(t, tableName1, *output.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL_METRICS",
		}
		key := "dynamodb:contributor-insights:test-table"
		if i > 1 {
			key += ":index-" + string(rune('0'+i))
		}
		require.NoError(t, state.Set(key, insight))
	}

	// Request with limit
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)
	require.Len(t, output.ContributorInsightsSummaries, 2)
	require.NotNil(t, output.NextToken)
}
