package dynamodb

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// Optional: Validate table exists if TableName is specified
	if input.TableName != nil && *input.TableName != "" {
		tableName := *input.TableName
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}
	}

	// List all contributor insights configurations
	// If TableName is specified, filter by table; otherwise return all
	prefix := "dynamodb:contributor-insights:"
	if input.TableName != nil && *input.TableName != "" {
		prefix = fmt.Sprintf("dynamodb:contributor-insights:%s", *input.TableName)
	}

	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	// Build list of summaries
	summaries := []map[string]interface{}{}
	for _, key := range keys {
		var insightsConfig map[string]interface{}
		if err := s.state.Get(key, &insightsConfig); err != nil {
			continue
		}

		// Extract table name and index name from key
		// Key format: "dynamodb:contributor-insights:tableName" or "dynamodb:contributor-insights:tableName:indexName"
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}

		tableName := parts[2]
		var indexName *string
		if len(parts) >= 4 {
			idx := strings.Join(parts[3:], ":")
			indexName = &idx
		}

		// Get status from config
		status, _ := insightsConfig["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}

		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": status,
		}

		if indexName != nil {
			summary["IndexName"] = *indexName
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination if needed
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// Simple pagination: NextToken is the start index as a string
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, you would parse the token
		// For simplicity, we'll just return all results
	}

	endIndex := startIndex + maxResults
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := summaries[startIndex:endIndex]

	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		response["NextToken"] = fmt.Sprintf("%d", endIndex)
	}

	return s.jsonResponse(200, response)
}
