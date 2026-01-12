package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// Optional table name filter
	var tableName string
	if input.TableName != nil {
		tableName = *input.TableName
	}

	// List all contributor insights configurations
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	var summaries []interface{}

	for _, key := range keys {
		var insightsConfig map[string]interface{}
		if err := s.state.Get(key, &insightsConfig); err != nil {
			continue
		}

		// Extract table name from insights config
		configTableName, _ := insightsConfig["TableName"].(string)

		// Filter by table name if specified
		if tableName != "" && configTableName != tableName {
			continue
		}

		// Build contributor insights summary
		summary := map[string]interface{}{}

		if configTableName != "" {
			summary["TableName"] = configTableName
		}

		// Add index name if present
		if indexName, ok := insightsConfig["IndexName"].(string); ok && indexName != "" {
			summary["IndexName"] = indexName
		}

		// Add status
		status, _ := insightsConfig["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}
		summary["ContributorInsightsStatus"] = status

		summaries = append(summaries, summary)
	}

	// Build response with pagination support
	response := map[string]interface{}{
		"ContributorInsightsSummaries": summaries,
	}

	// Handle pagination if MaxResults is specified
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults := int(*input.MaxResults)
		startIndex := 0

		// If NextToken is provided, decode it to get start index
		// For simplicity, we'll use a simple index-based pagination
		if input.NextToken != nil && *input.NextToken != "" {
			// In a real implementation, NextToken would be decoded
			// For now, we'll keep it simple and not implement pagination offset
		}

		// Apply pagination
		if len(summaries) > maxResults {
			response["ContributorInsightsSummaries"] = summaries[startIndex:maxResults]
			response["NextToken"] = "next-page-token" // Simplified token
		}
	}

	return s.jsonResponse(200, response)
}
