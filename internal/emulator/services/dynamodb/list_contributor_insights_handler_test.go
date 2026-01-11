package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContributorInsights_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some contributor insights entries in state
	insights1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insights2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributorinsights:test-table-1", insights1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributorinsights:test-table-2:test-index", insights2)
	require.NoError(t, err)

	// Test listing all insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create some contributor insights entries in state
	insights1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
	}
	insights2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
	}

	err := state.Set("dynamodb:contributorinsights:test-table-1", insights1)
	require.NoError(t, err)
	err = state.Set("dynamodb:contributorinsights:test-table-2", insights2)
	require.NoError(t, err)

	// Test filtering by table name
	input := &ListContributorInsightsInput{
		TableName: strPtr("test-table-1"),
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, "test-table-1", summary["TableName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights entries
	for i := 1; i <= 5; i++ {
		insights := map[string]interface{}{
			"TableName":                 fmt.Sprintf("test-table-%d", i),
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(fmt.Sprintf("dynamodb:contributorinsights:test-table-%d", i), insights)
		require.NoError(t, err)
	}

	// Test first page
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken
	nextToken, ok := result["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Test second page
	input2 := &ListContributorInsightsInput{
		MaxResults: &maxResults,
		NextToken:  &nextToken,
	}
	resp2, err := service.listContributorInsights(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.Unmarshal(resp2.Body, &result2)
	require.NoError(t, err)

	summaries2, ok := result2["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries2, 2)
}

func TestListContributorInsights_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Test with no insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, summaries)
}
