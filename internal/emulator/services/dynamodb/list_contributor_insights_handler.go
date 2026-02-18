package dynamodb

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// Build prefix for listing contributor insights
	var prefix string
	if input.TableName != nil && *input.TableName != "" {
		tableName := *input.TableName

		// Verify table exists
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}

		prefix = fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
	} else {
		prefix = "dynamodb:contributor-insights:"
	}

	// List all contributor insights entries
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	summaries := []interface{}{}

	for _, key := range keys {
		var insightsConfig map[string]interface{}
		if err := s.state.Get(key, &insightsConfig); err != nil {
			continue
		}

		// Extract table name and optional index name from state key
		// Key format: dynamodb:contributor-insights:<tableName> or dynamodb:contributor-insights:<tableName>:<indexName>
		keyWithoutPrefix := strings.TrimPrefix(key, "dynamodb:contributor-insights:")
		parts := strings.SplitN(keyWithoutPrefix, ":", 2)

		tableName := parts[0]
		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": "DISABLED",
		}

		if len(parts) == 2 && parts[1] != "" {
			summary["IndexName"] = parts[1]
		}

		if status, ok := insightsConfig["ContributorInsightsStatus"].(string); ok && status != "" {
			summary["ContributorInsightsStatus"] = status
		}

		summaries = append(summaries, summary)
	}

	// Apply MaxResults pagination
	maxResults := 100
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	// Handle NextToken pagination (simple offset-based)
	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// Find index matching the token (token is the last seen key)
		for i, summary := range summaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				tableN, _ := summaryMap["TableName"].(string)
				indexN, _ := summaryMap["IndexName"].(string)
				tokenKey := tableN
				if indexN != "" {
					tokenKey = tokenKey + ":" + indexN
				}
				if tokenKey == *input.NextToken {
					startIndex = i + 1
					break
				}
			}
		}
	}

	endIndex := startIndex + maxResults
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(summaries) {
		paginatedSummaries = summaries[startIndex:endIndex]
	}

	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	// Set NextToken if there are more results
	if endIndex < len(summaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			tableN, _ := lastSummary["TableName"].(string)
			indexN, _ := lastSummary["IndexName"].(string)
			nextToken := tableN
			if indexN != "" {
				nextToken = nextToken + ":" + indexN
			}
			response["NextToken"] = nextToken
		}
	}

	return s.jsonResponse(200, response)
}
