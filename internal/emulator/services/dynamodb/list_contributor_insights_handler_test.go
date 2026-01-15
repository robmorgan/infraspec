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

	// Setup: Create some contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"IndexName":                 "test-index",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "STANDARD",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2:test-index", insight2))

	input := &ListContributorInsightsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ContributorInsightsSummaries, 2)
}

func TestListContributorInsights_WithTableNameFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create some contributor insights data
	insight1 := map[string]interface{}{
		"TableName":                 "test-table-1",
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "STANDARD",
	}
	insight2 := map[string]interface{}{
		"TableName":                 "test-table-2",
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "STANDARD",
	}

	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor-insights:test-table-2", insight2))

	tableName := "test-table-1"
	input := &ListContributorInsightsInput{
		TableName: &tableName,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ContributorInsightsSummaries, 1)
	require.Equal(t, "test-table-1", *output.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_WithPagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Setup: Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		insight := map[string]interface{}{
			"TableName":                 "test-table",
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "STANDARD",
		}
		key := "dynamodb:contributor-insights:test-table"
		if i > 1 {
			key += ":index-" + string(rune('0'+i))
		}
		require.NoError(t, state.Set(key, insight))
	}

	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ContributorInsightsSummaries, 2)
	require.NotNil(t, output.NextToken) // Should have more results
}

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{}
	inputJSON, _ := json.Marshal(input)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-amz-json-1.0"},
		Body:    inputJSON,
		Action:  "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	require.NoError(t, json.Unmarshal(resp.Body, &output))
	require.Len(t, output.ContributorInsightsSummaries, 0)
	require.Nil(t, output.NextToken)
}
