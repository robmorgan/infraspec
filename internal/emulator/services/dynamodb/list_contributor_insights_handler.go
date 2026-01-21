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

	summaries := []ContributorInsightsSummary{}

	for _, key := range keys {
		var insight map[string]interface{}
		if err := s.state.Get(key, &insight); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableName, ok := insight["TableName"].(string); ok {
				if tableName != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := ContributorInsightsSummary{}

		if tableName, ok := insight["TableName"].(string); ok {
			summary.TableName = &tableName
		}

		if indexName, ok := insight["IndexName"].(string); ok && indexName != "" {
			summary.IndexName = &indexName
		}

		if status, ok := insight["ContributorInsightsStatus"].(string); ok {
			summary.ContributorInsightsStatus = ContributorInsightsStatus(status)
		}

		if mode, ok := insight["ContributorInsightsMode"].(string); ok {
			summary.ContributorInsightsMode = ContributorInsightsMode(mode)
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// If NextToken is provided, it would contain pagination info
	// For simplicity in this emulator, we'll just use basic pagination

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
	if endIndex < len(summaries) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
