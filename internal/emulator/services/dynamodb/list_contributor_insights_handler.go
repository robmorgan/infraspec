package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	contributorInsights := []interface{}{}

	for _, key := range keys {
		var insight map[string]interface{}
		if err := s.state.Get(key, &insight); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableNameInInsight, ok := insight["TableName"].(string); ok {
				if tableNameInInsight != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := map[string]interface{}{}

		if tableName, ok := insight["TableName"].(string); ok {
			summary["TableName"] = tableName
		}

		if indexName, ok := insight["IndexName"].(string); ok {
			summary["IndexName"] = indexName
		}

		if status, ok := insight["ContributorInsightsStatus"].(string); ok {
			summary["ContributorInsightsStatus"] = status
		}

		contributorInsights = append(contributorInsights, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded to determine start position
		// For simplicity, we'll start from the beginning
		startIndex = 0
	}

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(contributorInsights) {
		endIndex = len(contributorInsights)
	}

	paginatedInsights := []interface{}{}
	if startIndex < len(contributorInsights) {
		paginatedInsights = contributorInsights[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedInsights,
	}

	// Add NextToken if there are more results
	if endIndex < len(contributorInsights) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
