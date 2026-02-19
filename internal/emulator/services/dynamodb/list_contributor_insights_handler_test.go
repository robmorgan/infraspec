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

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "ContributorInsightsSummaries")
	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListContributorInsights_WithEntries(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create contributor insights entries for two tables
	table1 := "table-one"
	table2 := "table-two"

	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table1), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"LastUpdateDateTime":        float64(1000000),
	})
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table2), map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	})

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
}

func TestListContributorInsights_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	table1 := "my-table"
	table2 := "other-table"

	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table1), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	})
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", table2), map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	})
	// Add an index-level entry for table1
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s:my-index", table1), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	})

	input := &ListContributorInsightsInput{
		TableName: strPtr(table1),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	// Should return table-level and index-level entries for table1 only
	assert.Len(t, summaries, 2)
	for _, s := range summaries {
		m := s.(map[string]interface{})
		assert.Equal(t, table1, m["TableName"])
	}
}

func TestListContributorInsights_WithIndexEntry(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	indexName := "test-index"

	// Store index-level contributor insights
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"LastUpdateDateTime":        float64(9999999),
	})

	input := &ListContributorInsightsInput{
		TableName: strPtr(tableName),
	}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	require.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName, summary["TableName"])
	assert.Equal(t, indexName, summary["IndexName"])
	assert.Equal(t, "ENABLED", summary["ContributorInsightsStatus"])
	assert.Equal(t, float64(9999999), summary["LastUpdateDateTime"])
}

func TestListContributorInsights_ViaHandleRequest(t *testing.T) {
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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "ContributorInsightsSummaries")
}
