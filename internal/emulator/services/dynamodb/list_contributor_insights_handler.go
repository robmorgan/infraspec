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

	contributorInsightsSummaries := []ContributorInsightsSummary{}

	for _, key := range keys {
		var insightsData map[string]interface{}
		if err := s.state.Get(key, &insightsData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableName, ok := insightsData["TableName"].(string); ok {
				if tableName != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := ContributorInsightsSummary{}

		if tableName, ok := insightsData["TableName"].(string); ok {
			summary.TableName = &tableName
		}

		if indexName, ok := insightsData["IndexName"].(string); ok && indexName != "" {
			summary.IndexName = &indexName
		}

		if status, ok := insightsData["ContributorInsightsStatus"].(string); ok {
			summary.ContributorInsightsStatus = ContributorInsightsStatus(status)
		}

		if mode, ok := insightsData["ContributorInsightsMode"].(string); ok {
			summary.ContributorInsightsMode = ContributorInsightsMode(mode)
		}

		contributorInsightsSummaries = append(contributorInsightsSummaries, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded to get the start index
		// For emulator purposes, we'll keep it simple
		// NextToken could be implemented as base64(index) in production
		startIndex = 0 // Simplified for emulator
	}

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(contributorInsightsSummaries) {
		endIndex = len(contributorInsightsSummaries)
	}

	paginatedSummaries := []ContributorInsightsSummary{}
	if startIndex < len(contributorInsightsSummaries) {
		paginatedSummaries = contributorInsightsSummaries[startIndex:endIndex]
	}

	// Build response
	response := ListContributorInsightsOutput{
		ContributorInsightsSummaries: paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(contributorInsightsSummaries) {
		nextToken := "next-token" // Simplified token for emulator
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
