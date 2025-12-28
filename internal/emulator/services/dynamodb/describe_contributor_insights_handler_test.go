package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeContributorInsights_Success_Disabled(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":   uuid.New().String(),
	}
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)

	input := &DescribeContributorInsightsInput{
		TableName: strPtr(tableName),
	}

	resp, err := service.describeContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, tableName, result["TableName"])
	assert.Equal(t, "DISABLED", result["ContributorInsightsStatus"])
}

func TestDescribeContributorInsights_Success_Enabled(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":   uuid.New().String(),
	}
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)

	// Create contributor insights configuration
	insightsConfig := map[string]interface{}{
		"ContributorInsightsStatus": "ENABLED",
		"LastUpdateDateTime":        float64(1234567890),
	}
	state.Set(fmt.Sprintf("dynamodb:contributor-insights:%s", tableName), insightsConfig)

	input := &DescribeContributorInsightsInput{
		TableName: strPtr(tableName),
	}

	resp, err := service.describeContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, tableName, result["TableName"])
	assert.Equal(t, "ENABLED", result["ContributorInsightsStatus"])
	assert.Equal(t, float64(1234567890), result["LastUpdateDateTime"])
}

func TestDescribeContributorInsights_WithIndexName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":   uuid.New().String(),
	}
	state.Set(fmt.Sprintf("dynamodb:table:%s", tableName), tableDesc)

	input := &DescribeContributorInsightsInput{
		TableName: strPtr(tableName),
		IndexName: strPtr("test-index"),
	}

	resp, err := service.describeContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	assert.Equal(t, tableName, result["TableName"])
	assert.Equal(t, "test-index", result["IndexName"])
	assert.Equal(t, "DISABLED", result["ContributorInsightsStatus"])
}

func TestDescribeContributorInsights_MissingTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DescribeContributorInsightsInput{}

	resp, err := service.describeContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "TableName is required")
}

func TestDescribeContributorInsights_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DescribeContributorInsightsInput{
		TableName: strPtr("nonexistent-table"),
	}

	resp, err := service.describeContributorInsights(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", result["__type"])
}
