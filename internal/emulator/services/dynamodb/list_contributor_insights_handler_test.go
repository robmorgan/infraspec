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
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	require.Empty(t, summaries)
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some contributor insights to state
	insight1 := map[string]interface{}{
		"TableName":                 "TestTable1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set("dynamodb:contributor-insights:TestTable1", insight1)
	require.NoError(t, err)

	insight2 := map[string]interface{}{
		"TableName":                 "TestTable2",
		"IndexName":                 "TestIndex",
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set("dynamodb:contributor-insights:TestTable2:TestIndex", insight2)
	require.NoError(t, err)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 2)
}

func TestListContributorInsights_FilterByTable(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add some contributor insights to state
	insight1 := map[string]interface{}{
		"TableName":                 "TestTable1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set("dynamodb:contributor-insights:TestTable1", insight1)
	require.NoError(t, err)

	insight2 := map[string]interface{}{
		"TableName":                 "TestTable2",
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set("dynamodb:contributor-insights:TestTable2", insight2)
	require.NoError(t, err)

	tableName := "TestTable1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "TestTable1", summary["TableName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Add multiple contributor insights to state
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "TestTable" + string(rune('0'+i)),
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set("dynamodb:contributor-insights:TestTable"+string(rune('0'+i)), insight)
		require.NoError(t, err)
	}

	maxResults := int32(3)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body, &response)
	require.NoError(t, err)

	summaries, ok := response["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 3)

	// Should have NextToken since we have more results
	_, hasNextToken := response["NextToken"]
	require.True(t, hasNextToken)
}
