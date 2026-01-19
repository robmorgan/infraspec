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

	// Create some contributor insights configurations in state
	config1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	state.Set("dynamodb:contributorinsights:test-table-1", config1)

	config2 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"IndexName":                 "gsi-1",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	state.Set("dynamodb:contributorinsights:test-table-1:gsi-1", config2)

	config3 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "THROTTLE_ONLY",
	}
	state.Set("dynamodb:contributorinsights:test-table-2", config3)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 3, len(summaries))
}

func TestListContributorInsights_FilterByTable(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights configurations for multiple tables
	config1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	state.Set("dynamodb:contributorinsights:test-table-1", config1)

	config2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	state.Set("dynamodb:contributorinsights:test-table-2", config2)

	tableName := "test-table-1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 1, len(summaries))

	// Verify the returned summary is for test-table-1
	summary := summaries[0].(map[string]interface{})
	require.Equal(t, "test-table-1", summary["TableName"])
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 0, len(summaries))
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights configurations
	for i := 1; i <= 5; i++ {
		config := map[string]interface{}{
			"TableName":                 "test-table-" + string(rune('0'+i)),
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "STANDARD",
		}
		state.Set("dynamodb:contributorinsights:test-table-"+string(rune('0'+i)), config)
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(summaries))

	// Verify NextToken is present
	_, hasNextToken := result["NextToken"]
	require.True(t, hasNextToken)
}
