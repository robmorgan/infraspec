package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/require"
)

func TestListContributorInsights_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some contributor insights entries
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-1:test-index", insight2))
	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-2", insight3))

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 3)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some contributor insights entries
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
	}
	insight3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-1:test-index", insight2))
	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-2", insight3))

	input := &ListContributorInsightsInput{
		TableName: strPtr("test-table-1"),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Verify both summaries are for test-table-1
	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})
		require.Equal(t, "test-table-1", summaryMap["TableName"])
	}
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create several contributor insights entries
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table-" + string(rune('0'+i)),
			"ContributorInsightsStatus": "ENABLED",
		}
		require.NoError(t, state.Set("dynamodb:contributorinsights:test-table-"+string(rune('0'+i)), insight))
	}

	// Request with MaxResults = 2
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)

	// Should have NextToken since there are more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}

func TestListContributorInsights_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 0)
}

func TestListContributorInsights_DefaultStatus(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create an insight without status
	insight := map[string]interface{}{
		"TableName": "test-table",
	}
	require.NoError(t, state.Set("dynamodb:contributorinsights:test-table", insight))

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "DISABLED", summary["ContributorInsightsStatus"])
}
