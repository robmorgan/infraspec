package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	summaries := []ContributorInsightsSummary{}

	for _, key := range keys {
		var insightData map[string]interface{}
		if err := s.state.Get(key, &insightData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableNameInData, ok := insightData["TableName"].(string); ok {
				if tableNameInData != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build summary
		summary := ContributorInsightsSummary{}

		if tableName, ok := insightData["TableName"].(string); ok {
			summary.TableName = &tableName
		}

		if indexName, ok := insightData["IndexName"].(string); ok && indexName != "" {
			summary.IndexName = &indexName
		}

		if status, ok := insightData["ContributorInsightsStatus"].(string); ok {
			summary.ContributorInsightsStatus = ContributorInsightsStatus(status)
		}

		if mode, ok := insightData["ContributorInsightsMode"].(string); ok {
			summary.ContributorInsightsMode = ContributorInsightsMode(mode)
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// NextToken would typically be a base64-encoded continuation token
	// For simplicity in the emulator, we'll just return all results

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := []ContributorInsightsSummary{}
	if startIndex < len(summaries) {
		paginatedSummaries = summaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	// In a real implementation, this would be a proper continuation token
	if endIndex < len(summaries) {
		nextToken := "next-page-token"
		response["NextToken"] = nextToken
	}

	return s.jsonResponse(200, response)
}
