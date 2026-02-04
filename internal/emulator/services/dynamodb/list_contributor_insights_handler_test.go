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

func TestListContributorInsights_WithEntries(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Seed contributor insights entries
	table1 := "table-alpha"
	table2 := "table-beta"

	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table1), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL",
	})
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table2), map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL",
	})

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

	for _, s := range summaries {
		sMap, ok := s.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, sMap, "TableName")
		assert.Contains(t, sMap, "ContributorInsightsStatus")
		assert.Contains(t, sMap, "ContributorInsightsMode")
	}
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "target-table"

	// Create the table so validation passes
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), map[string]interface{}{
		"TableName": tableName,
	})

	// Seed insights for two tables
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL",
	})
	state.Set("dynamodb:contributor-insights:other-table", map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
		"ContributorInsightsMode":   "ALL",
	})

	input := &ListContributorInsightsInput{
		TableName: strPtr(tableName),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 1)

	sMap := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName, sMap["TableName"])
}

func TestListContributorInsights_FilterByTableName_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &ListContributorInsightsInput{
		TableName: strPtr("nonexistent-table"),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Equal(t, "ResourceNotFoundException", responseBody["__type"])
}

func TestListContributorInsights_WithIndexName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "my-table"
	indexName := "gsi-index"

	// Seed an entry with index
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"ContributorInsightsMode":   "ALL",
	})

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

	sMap := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName, sMap["TableName"])
	assert.Equal(t, indexName, sMap["IndexName"])
}

func TestListContributorInsights_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Seed 5 entries
	for i := 1; i <= 5; i++ {
		state.Set(fmt.Sprintf("dynamodb:contributor-insights:table-%d", i), map[string]interface{}{
			"ContributorInsightsStatus": "ENABLED",
			"ContributorInsightsMode":   "ALL",
		})
	}

	limit := int32(2)
	input := &ListContributorInsightsInput{
		MaxResults: &limit,
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)

	// Should have NextToken when there are more results
	assert.Contains(t, responseBody, "NextToken")
}
