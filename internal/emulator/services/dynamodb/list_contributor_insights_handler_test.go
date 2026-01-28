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

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Empty(t, responseBody.ContributorInsightsSummaries, "Should have no contributor insights initially")
	assert.Nil(t, responseBody.NextToken, "Should have no next token")
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName := "test-table"

	// Insight for the table
	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	insight1Data := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	// Insight for a global secondary index
	indexName := "test-index"
	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName)
	insight2Data := map[string]interface{}{
		"TableName":                 tableName,
		"IndexName":                 indexName,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "TOP_N_KEYS",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List all contributor insights
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, responseBody.ContributorInsightsSummaries, 2, "Should have two contributor insights")

	// Verify summaries contain expected fields
	for _, summary := range responseBody.ContributorInsightsSummaries {
		assert.NotNil(t, summary.TableName)
		assert.Equal(t, tableName, *summary.TableName)
		assert.NotEmpty(t, summary.ContributorInsightsStatus)
		assert.NotEmpty(t, summary.ContributorInsightsMode)
	}
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create insights for different tables
	table1Name := "table1"
	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table1Name)
	insight1Data := map[string]interface{}{
		"TableName":                 table1Name,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	table2Name := "table2"
	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table2Name)
	insight2Data := map[string]interface{}{
		"TableName":                 table2Name,
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL_METRICS",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List insights for table1 only
	input := &ListContributorInsightsInput{
		TableName: &table1Name,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ContributorInsightsSummaries, 1, "Should have only one insight for table1")
	assert.Equal(t, table1Name, *responseBody.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("table%d", i)
		insightKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
		insightData := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL_METRICS",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// List insights with max results
	maxResults := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &maxResults,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Len(t, responseBody.ContributorInsightsSummaries, 2, "Should have only 2 insights due to max results")

	// Should have NextToken for pagination
	assert.NotNil(t, responseBody.NextToken)
}

func TestListContributorInsights_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListContributorInsights should work with no parameters
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.NotNil(t, responseBody.ContributorInsightsSummaries)
}
