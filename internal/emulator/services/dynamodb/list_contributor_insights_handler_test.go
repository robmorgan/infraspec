package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
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

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, summaries)
}

func TestListContributorInsights_WithTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":   uuid.New().String(),
	}
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)

	// Store contributor insights for the table
	insightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
	}
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), insightsConfig)

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
	assert.True(t, ok)
	assert.Len(t, summaries, 1)

	summary := summaries[0].(map[string]interface{})
	assert.Equal(t, tableName, summary["TableName"])
	assert.Equal(t, "ENABLED", summary["ContributorInsightsStatus"])
}

func TestListContributorInsights_WithIndex(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	tableName := "test-table"
	indexName := "my-gsi"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":   uuid.New().String(),
	}
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)

	// Store contributor insights for the table and an index
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), map[string]interface{}{
		"ContributorInsightsStatus": "DISABLED",
	})
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName), map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
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
	assert.True(t, ok)
	assert.Len(t, summaries, 2)
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

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", result["__type"])
}

func TestListContributorInsights_NoTableFilter(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create two tables with contributor insights
	for _, name := range []string{"table-a", "table-b"} {
		state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", name), map[string]interface{}{
			"ContributorInsightsStatus": "DISABLED",
		})
	}

	input := &ListContributorInsightsInput{}

	resp, err := service.listContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	summaries, ok := result["ContributorInsightsSummaries"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, summaries, 2)
}
