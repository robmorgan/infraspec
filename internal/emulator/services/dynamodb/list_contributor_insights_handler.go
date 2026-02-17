package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights from state
	keys, err := s.state.List("dynamodb:contributorinsights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	contributorInsightsSummaries := []interface{}{}

	for _, key := range keys {
		var insightsData map[string]interface{}
		if err := s.state.Get(key, &insightsData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableNameInInsights, ok := insightsData["TableName"].(string); ok {
				if tableNameInInsights != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := map[string]interface{}{}

		if tableName, ok := insightsData["TableName"].(string); ok {
			summary["TableName"] = tableName
		}
		if indexName, ok := insightsData["IndexName"].(string); ok {
			summary["IndexName"] = indexName
		}
		if status, ok := insightsData["ContributorInsightsStatus"].(string); ok {
			summary["ContributorInsightsStatus"] = status
		}

		contributorInsightsSummaries = append(contributorInsightsSummaries, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default limit
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// For NextToken, we could implement pagination logic here if needed
	// For now, simple implementation without token parsing

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(contributorInsightsSummaries) {
		endIndex = len(contributorInsightsSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(contributorInsightsSummaries) {
		paginatedSummaries = contributorInsightsSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(contributorInsightsSummaries) {
		response["NextToken"] = "has-more-results" // Simplified token
	}

	return s.jsonResponse(200, response)
}
