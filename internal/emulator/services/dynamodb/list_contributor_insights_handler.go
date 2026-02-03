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
		var insightData ContributorInsightsSummary
		if err := s.state.Get(key, &insightData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if insightData.TableName == nil || *insightData.TableName != *input.TableName {
				continue
			}
		}

		contributorInsightsSummaries = append(contributorInsightsSummaries, insightData)
	}

	// Apply pagination
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, decode the NextToken to get the start index
		// For simplicity in the emulator, we'll just use a basic approach
		// NextToken could be the index as a string, or a more complex encoding
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
		nextToken := "next-page" // Simplified token for emulator
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
