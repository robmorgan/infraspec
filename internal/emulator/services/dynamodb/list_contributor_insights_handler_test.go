package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/testing/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContributorInsights_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - contributor insights for a table
	insightsData1 := map[string]interface{}{
		"TableName":                 "test-table",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_ATTRIBUTES",
	}
	err := state.Set("dynamodb:contributor-insights:test-table", insightsData1)
	require.NoError(t, err)

	// Create test data - contributor insights for a table with index
	insightsData2 := map[string]interface{}{
		"TableName":                 "test-table",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "SPECIFIC_ATTRIBUTES",
	}
	err = state.Set("dynamodb:contributor-insights:test-table:test-index", insightsData2)
	require.NoError(t, err)

	// Create test data for another table
	insightsData3 := map[string]interface{}{
		"TableName":                 "other-table",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL_ATTRIBUTES",
	}
	err = state.Set("dynamodb:contributor-insights:other-table", insightsData3)
	require.NoError(t, err)

	// Test listing all contributor insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 3)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data
	insightsData1 := map[string]interface{}{
		"TableName":                 "test-table",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_ATTRIBUTES",
	}
	err := state.Set("dynamodb:contributor-insights:test-table", insightsData1)
	require.NoError(t, err)

	insightsData2 := map[string]interface{}{
		"TableName":                 "test-table",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "SPECIFIC_ATTRIBUTES",
	}
	err = state.Set("dynamodb:contributor-insights:test-table:test-index", insightsData2)
	require.NoError(t, err)

	insightsData3 := map[string]interface{}{
		"TableName":                 "other-table",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL_ATTRIBUTES",
	}
	err = state.Set("dynamodb:contributor-insights:other-table", insightsData3)
	require.NoError(t, err)

	// Test filtering by table name
	tableName := "test-table"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 2)
	for _, summary := range output.ContributorInsightsSummaries {
		assert.Equal(t, "test-table", *summary.TableName)
	}
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Empty(t, output.ContributorInsightsSummaries)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test data - 5 insights
	for i := 1; i <= 5; i++ {
		insightsData := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL_ATTRIBUTES",
		}
		err := state.Set("dynamodb:contributor-insights:test-table-"+string(rune('0'+i)), insightsData)
		require.NoError(t, err)
	}

	// Request with max results of 2
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 2)
	assert.NotNil(t, output.NextToken)
}
