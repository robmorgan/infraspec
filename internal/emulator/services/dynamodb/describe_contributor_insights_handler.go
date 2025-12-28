package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) describeContributorInsights(ctx context.Context, input *DescribeContributorInsightsInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}

	tableName := *input.TableName

	// Extract optional IndexName parameter
	var indexName string
	if input.IndexName != nil {
		indexName = *input.IndexName
	}

	// Verify table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException",
			fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Build state key for contributor insights
	var insightsKey string
	if indexName != "" {
		insightsKey = fmt.Sprintf("dynamodb:contributor-insights:%s:%s", tableName, indexName)
	} else {
		insightsKey = fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	}

	// Try to get existing contributor insights configuration
	var insightsConfig map[string]interface{}
	if err := s.state.Get(insightsKey, &insightsConfig); err != nil {
		// If not configured, return default DISABLED status
		insightsConfig = map[string]interface{}{
			"ContributorInsightsStatus": "DISABLED",
		}
	}

	// Get status from config or default to DISABLED
	status, _ := insightsConfig["ContributorInsightsStatus"].(string)
	if status == "" {
		status = "DISABLED"
	}

	// Build response
	response := map[string]interface{}{
		"TableName":                 tableName,
		"ContributorInsightsStatus": status,
	}

	// Add IndexName if specified
	if indexName != "" {
		response["IndexName"] = indexName
	}

	// Add failure exception if status is FAILED
	if status == "FAILED" {
		if failureException, ok := insightsConfig["FailureException"].(map[string]interface{}); ok {
			response["FailureException"] = failureException
		}
	}

	// Add last update time if available
	if lastUpdateTime, ok := insightsConfig["LastUpdateDateTime"].(float64); ok {
		response["LastUpdateDateTime"] = lastUpdateTime
	} else if status != "DISABLED" {
		// Default to current time for enabled insights
		response["LastUpdateDateTime"] = float64(time.Now().Unix())
	}

	return s.jsonResponse(200, response)
}
