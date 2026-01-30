package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights from state
	keys, err := s.state.List("dynamodb:contributorinsights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	contributorInsightsSummaries := []interface{}{}

	for _, key := range keys {
		var insightData map[string]interface{}
		if err := s.state.Get(key, &insightData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableNameInInsight, ok := insightData["TableName"].(string); ok {
				if tableNameInInsight != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		contributorInsightsSummaries = append(contributorInsightsSummaries, insightData)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// For simplicity, NextToken represents the start index as a string
		// In a production system, this would be a properly encoded token
		var tokenIndex int
		if _, err := fmt.Sscanf(*input.NextToken, "%d", &tokenIndex); err == nil {
			startIndex = tokenIndex
		}
	}

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
		"ContributorInsightsSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(contributorInsightsSummaries) {
		response["NextToken"] = fmt.Sprintf("%d", endIndex)
	}

	return s.jsonResponse(200, response)
}
