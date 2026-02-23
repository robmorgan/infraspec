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

	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListContributorInsights_WithTableInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights for a table
	tableName := "test-table"
	insightsKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	insightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(insightsKey, insightsConfig)
	require.NoError(t, err)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, tableName, summary["TableName"])
	assert.Equal(t, "ENABLED", summary["ContributorInsightsStatus"])
	assert.NotContains(t, summary, "IndexName")
}

func TestListContributorInsights_WithIndexInsights(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	indexName := "test-index"

	// Create contributor insights for a table and an index
	tableInsightsKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	tableInsightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	}
	err := state.Set(tableInsightsKey, tableInsightsConfig)
	require.NoError(t, err)

	indexInsightsKey := fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName)
	indexInsightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	}
	err = state.Set(indexInsightsKey, indexInsightsConfig)
	require.NoError(t, err)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1 := "table-one"
	table2 := "table-two"

	err := state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table1), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	})
	require.NoError(t, err)

	err = state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table2), map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	})
	require.NoError(t, err)

	// Filter by table1 only
	input := &ListContributorInsightsInput{
		TableName: strPtr(table1),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1, summary["TableName"])
	assert.Equal(t, "ENABLED", summary["ContributorInsightsStatus"])
}

func TestListContributorInsights_DefaultStatus(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Store insights config without status
	tableName := "no-status-table"
	err := state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), map[string]interface{}{})
	require.NoError(t, err)

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "DISABLED", summary["ContributorInsightsStatus"])
}
