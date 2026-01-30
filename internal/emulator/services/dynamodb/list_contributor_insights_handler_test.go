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

func TestListContributorInsights_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListContributorInsights",
		},
		Body:   []byte("{}"),
		Action: "ListContributorInsights",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ContributorInsightsSummaryList")
	summaries, ok := responseBody["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaryList should be an array")
	assert.Empty(t, summaries, "Should have no contributor insights initially")
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName := "test-table"

	insight1Key := fmt.Sprintf("dynamodb:contributorinsights:%s", tableName)
	insight1Data := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"IndexName":                 nil,
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributorinsights:%s:index1", tableName)
	insight2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 "index1",
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List all contributor insights
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "ContributorInsightsSummaryList")
	summaries, ok := responseBody["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaryList should be an array")
	assert.Len(t, summaries, 2, "Should have 2 contributor insights")
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights for different tables
	table1 := "table-1"
	table2 := "table-2"

	insight1Key := fmt.Sprintf("dynamodb:contributorinsights:%s", table1)
	insight1Data := map[string]interface{}{
		"TableName":                 table1,
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributorinsights:%s", table2)
	insight2Data := map[string]interface{}{
		"TableName":                 table2,
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List contributor insights for table1 only
	input := &ListContributorInsightsInput{
		TableName: &table1,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1, "Should have 1 contributor insight for table-1")

	// Verify it's the correct table
	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, table1, summary["TableName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("table-%d", i)
		insightKey := fmt.Sprintf("dynamodb:contributorinsights:%s", tableName)
		insightData := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// List with pagination (2 results per page)
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 contributor insights in first page")

	// Verify NextToken is present
	assert.Contains(t, responseBody, "NextToken")
	nextToken, ok := responseBody["NextToken"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, nextToken)

	// Fetch next page
	input.NextToken = &nextToken
	resp, err = service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)

	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok = responseBody["ContributorInsightsSummaryList"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2, "Should have 2 contributor insights in second page")
}
