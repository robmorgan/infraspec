package dynamodb

import (
	"context"
	"strings"

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

	summaries := []interface{}{}

	for _, key := range keys {
		var insightData map[string]interface{}
		if err := s.state.Get(key, &insightData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableName, ok := insightData["TableName"].(string); ok {
				if tableName != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insight summary
		summary := map[string]interface{}{}

		if tableName, ok := insightData["TableName"].(string); ok {
			summary["TableName"] = tableName
		}

		if indexName, ok := insightData["IndexName"].(string); ok && indexName != "" {
			summary["IndexName"] = indexName
		}

		if status, ok := insightData["ContributorInsightsStatus"].(string); ok {
			summary["ContributorInsightsStatus"] = status
		} else {
			summary["ContributorInsightsStatus"] = "DISABLED"
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
		// In a real implementation, NextToken would be decoded to find the start position
		// For the emulator, we'll use a simple approach
		for i, summary := range summaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				tableName := ""
				indexName := ""
				if tn, ok := summaryMap["TableName"].(string); ok {
					tableName = tn
				}
				if in, ok := summaryMap["IndexName"].(string); ok {
					indexName = in
				}
				token := tableName
				if indexName != "" {
					token = tableName + ":" + indexName
				}
				if token == *input.NextToken {
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

	paginatedSummaries := []interface{}{}
	if startIndex < len(summaries) {
		paginatedSummaries = summaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			tableName := ""
			indexName := ""
			if tn, ok := lastSummary["TableName"].(string); ok {
				tableName = tn
			}
			if in, ok := lastSummary["IndexName"].(string); ok {
				indexName = in
			}
			token := tableName
			if indexName != "" {
				token = tableName + ":" + indexName
			}
			response["NextToken"] = token
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
