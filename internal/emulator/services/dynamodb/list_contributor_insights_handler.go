package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// Optional: Validate table exists if TableName is provided
	if input.TableName != nil && *input.TableName != "" {
		tableName := *input.TableName
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}
	}

	// List all contributor insights from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	contributorInsightsSummaries := []interface{}{}

	for _, key := range keys {
		var insightsConfig map[string]interface{}
		if err := s.state.Get(key, &insightsConfig); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if configTableName, ok := insightsConfig["TableName"].(string); ok {
				if configTableName != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Build contributor insights summary
		summary := map[string]interface{}{
			"ContributorInsightsStatus": "DISABLED",
		}

		if tableName, ok := insightsConfig["TableName"].(string); ok {
			summary["TableName"] = tableName
		}

		if indexName, ok := insightsConfig["IndexName"].(string); ok {
			summary["IndexName"] = indexName
		}

		if status, ok := insightsConfig["ContributorInsightsStatus"].(string); ok {
			summary["ContributorInsightsStatus"] = status
		}

		contributorInsightsSummaries = append(contributorInsightsSummaries, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// Note: For simplicity, we're not implementing NextToken pagination in this basic implementation
	// In a real implementation, NextToken would encode pagination state

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
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
