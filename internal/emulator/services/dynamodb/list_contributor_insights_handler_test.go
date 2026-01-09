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

	// Create some test contributor insights
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"

	insight1 := map[string]interface{}{
		"TableName": tableName1,
		"IndexName": "",
		"Status":    "ENABLED",
		"Mode":      "ALL_ATTRIBUTES",
	}

	insight2 := map[string]interface{}{
		"TableName": tableName1,
		"IndexName": "test-index",
		"Status":    "ENABLED",
		"Mode":      "BASIC",
	}

	insight3 := map[string]interface{}{
		"TableName": tableName2,
		"IndexName": "",
		"Status":    "DISABLED",
		"Mode":      "ALL_ATTRIBUTES",
	}

	require.NoError(t, state.Set("dynamodb:contributor_insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor_insights:test-table-1:test-index", insight2))
	require.NoError(t, state.Set("dynamodb:contributor_insights:test-table-2", insight3))

	// Test without filter
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(summaries))
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName1 := "test-table-1"
	tableName2 := "test-table-2"

	insight1 := map[string]interface{}{
		"TableName": tableName1,
		"Status":    "ENABLED",
		"Mode":      "ALL_ATTRIBUTES",
	}

	insight2 := map[string]interface{}{
		"TableName": tableName2,
		"Status":    "DISABLED",
		"Mode":      "ALL_ATTRIBUTES",
	}

	require.NoError(t, state.Set("dynamodb:contributor_insights:test-table-1", insight1))
	require.NoError(t, state.Set("dynamodb:contributor_insights:test-table-2", insight2))

	// Test with table name filter
	input := &ListContributorInsightsInput{
		TableName: &tableName1,
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(summaries))

	// Verify it's the correct table
	firstSummary := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName1, firstSummary["TableName"])
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(summaries))
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple insights
	for i := 0; i < 5; i++ {
		insight := map[string]interface{}{
			"TableName": "test-table",
			"Status":    "ENABLED",
			"Mode":      "ALL_ATTRIBUTES",
		}
		key := "dynamodb:contributor_insights:test-table"
		if i > 0 {
			key = key + "-" + string(rune('a'+i))
		}
		require.NoError(t, state.Set(key, insight))
	}

	// Test with max results
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
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(summaries))

	// Should have NextToken since there are more results
	_, hasNextToken := result["NextToken"]
	assert.True(t, hasNextToken)
}
