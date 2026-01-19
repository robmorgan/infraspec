package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights configurations from state
	keys, err := s.state.List("dynamodb:contributorinsights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	summaries := []ContributorInsightsSummary{}

	for _, key := range keys {
		var config map[string]interface{}
		if err := s.state.Get(key, &config); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableName, ok := config["TableName"].(string); ok {
				if tableName != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := ContributorInsightsSummary{}

		if tableName, ok := config["TableName"].(string); ok {
			summary.TableName = &tableName
		}

		if indexName, ok := config["IndexName"].(string); ok && indexName != "" {
			summary.IndexName = &indexName
		}

		if status, ok := config["ContributorInsightsStatus"].(string); ok {
			summary.ContributorInsightsStatus = ContributorInsightsStatus(status)
		}

		if mode, ok := config["ContributorInsightsMode"].(string); ok {
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
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded to determine start position
		// For simplicity, we'll use the token value as-is
		// This is a simplified pagination implementation
		for i, summary := range summaries {
			if summary.TableName != nil {
				key := *summary.TableName
				if summary.IndexName != nil {
					key += ":" + *summary.IndexName
				}
				if key == *input.NextToken {
					startIndex = i + 1
					break
				}
			}
		}
	}

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
	if endIndex < len(summaries) {
		if nextSummary := summaries[endIndex]; nextSummary.TableName != nil {
			nextToken := *nextSummary.TableName
			if nextSummary.IndexName != nil {
				nextToken += ":" + *nextSummary.IndexName
			}
			response["NextToken"] = nextToken
		}
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract table name from contributor insights key
func extractTableNameFromContributorInsightsKey(key string) string {
	// Key format: "dynamodb:contributorinsights:tablename" or "dynamodb:contributorinsights:tablename:indexname"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
