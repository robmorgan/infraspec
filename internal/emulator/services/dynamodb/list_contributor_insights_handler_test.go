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
	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Empty(t, summaries, "Should have no contributor insights initially")
}

func TestListContributorInsights_WithInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test contributor insights
	tableName := "test-table"
	indexName := "test-index"

	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	tableNameCopy := tableName
	insight1Data := ContributorInsightsSummary{
		TableName:                 &tableNameCopy,
		ContributorInsightsStatus: "ENABLED",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName)
	indexNameCopy := indexName
	insight2Data := ContributorInsightsSummary{
		TableName:                 &tableNameCopy,
		IndexName:                 &indexNameCopy,
		ContributorInsightsStatus: "ENABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List all contributor insights
	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, output.ContributorInsightsSummaries, 2, "Should have two contributor insights")

	// Verify summaries contain expected fields
	for _, summary := range output.ContributorInsightsSummaries {
		assert.NotNil(t, summary.TableName)
		assert.Equal(t, tableName, *summary.TableName)
		assert.Equal(t, ContributorInsightsStatus("ENABLED"), summary.ContributorInsightsStatus)
	}
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights for different tables
	table1Name := "table1"
	table1NameCopy := table1Name
	insight1Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table1Name)
	insight1Data := ContributorInsightsSummary{
		TableName:                 &table1NameCopy,
		ContributorInsightsStatus: "ENABLED",
	}
	err := state.Set(insight1Key, insight1Data)
	require.NoError(t, err)

	table2Name := "table2"
	table2NameCopy := table2Name
	insight2Key := fmt.Sprintf("dynamodb:contributor-insights:%s", table2Name)
	insight2Data := ContributorInsightsSummary{
		TableName:                 &table2NameCopy,
		ContributorInsightsStatus: "DISABLED",
	}
	err = state.Set(insight2Key, insight2Data)
	require.NoError(t, err)

	// List contributor insights for table1 only
	input := &ListContributorInsightsInput{
		TableName: &table1Name,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.Len(t, output.ContributorInsightsSummaries, 1, "Should have only one contributor insight for table1")
	assert.Equal(t, table1Name, *output.ContributorInsightsSummaries[0].TableName)
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple contributor insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("table%d", i)
		tableNameCopy := tableName
		insightKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
		insightData := ContributorInsightsSummary{
			TableName:                 &tableNameCopy,
			ContributorInsightsStatus: "ENABLED",
		}
		err := state.Set(insightKey, insightData)
		require.NoError(t, err)
	}

	// List contributor insights with limit
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

	assert.Len(t, output.ContributorInsightsSummaries, 2, "Should have only 2 contributor insights due to limit")

	// Should have NextToken for pagination
	assert.NotNil(t, output.NextToken, "Should have NextToken when there are more results")
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

	var output ListContributorInsightsOutput
	err = json.Unmarshal(resp.Body, &output)
	require.NoError(t, err)

	assert.NotNil(t, output.ContributorInsightsSummaries)
}
