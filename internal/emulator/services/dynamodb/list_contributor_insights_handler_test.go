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

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok, "ContributorInsightsSummaries should be an array")
	assert.Empty(t, summaries)
}

func TestListContributorInsights_WithTableName_TableExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)
	require.NoError(t, err)

	// Set contributor insights to ENABLED for the table
	insightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	}
	err = state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), insightsConfig)
	require.NoError(t, err)

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
	require.Len(t, summaries, 1)

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, tableName, summaryMap["TableName"])
	assert.Equal(t, "ENABLED", summaryMap["ContributorInsightsStatus"])
}

func TestListContributorInsights_WithTableName_DefaultDisabled(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)
	require.NoError(t, err)

	// No contributor insights configured - should default to DISABLED
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
	require.Len(t, summaries, 1)

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, tableName, summaryMap["TableName"])
	assert.Equal(t, "DISABLED", summaryMap["ContributorInsightsStatus"])
}

func TestListContributorInsights_TableNotFound(t *testing.T) {
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

func TestListContributorInsights_WithMaxResults(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple tables with contributor insights
	for i := 1; i <= 5; i++ {
		tableName := fmt.Sprintf("table-%d", i)
		tableDesc := map[string]interface{}{
			"TableName": tableName,
		}
		err := state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)
		require.NoError(t, err)

		insightsConfig := map[string]interface{}{
			"ContributorInsightsStatus": "ENABLED",
		}
		err = state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), insightsConfig)
		require.NoError(t, err)
	}

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

	summaries, ok := responseBody["ContributorInsightsSummaries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, summaries, 2)
	assert.Contains(t, responseBody, "NextToken")
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
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)
	assert.Contains(t, responseBody, "ContributorInsightsSummaries")
}
