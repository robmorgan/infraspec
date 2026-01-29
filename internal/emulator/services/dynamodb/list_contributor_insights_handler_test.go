package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) (*DynamoDBService, emulator.StateManager) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)
	return service, state
}

func TestListContributorInsights_Success(t *testing.T) {
	service, state := setupTestService(t)

	// Create test contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "test-index-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1:test-index-1", insight2))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2", insight3))

	// Test listing all contributor insights
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 3)
}

func TestListContributorInsights_FilterByTable(t *testing.T) {
	service, state := setupTestService(t)

	// Create test contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "RULE_BASED",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2", insight2))

	// Test listing contributor insights for specific table
	tableName := "test-table-1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 1)
	assert.Equal(t, "test-table-1", *output.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	service, _ := setupTestService(t)

	// Test listing when no contributor insights exist
	input := &ListContributorInsightsInput{}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 0)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	service, state := setupTestService(t)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "RULE_BASED",
		}
		key := "dynamodb:contributor-insights:test-table:" + string(rune('a'+i))
		require.NoError(t, state.Set(key, insight))
	}

	// Test pagination with MaxResults
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	resp, err := service.listContributorInsights(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 2)
	assert.NotNil(t, output.NextToken)
}
