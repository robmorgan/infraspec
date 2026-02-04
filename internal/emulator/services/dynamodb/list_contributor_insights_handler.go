package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// If TableName is specified, verify the table exists
	if input.TableName != nil && *input.TableName != "" {
		tableName := *input.TableName
		tableKey := "dynamodb:table:" + tableName
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				"Requested resource not found: Table: "+tableName+" not found"), nil
		}
	}

	// List all contributor insights entries from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
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
		// Key format: "dynamodb:contributor-insights:<tableName>" or "dynamodb:contributor-insights:<tableName>:<indexName>"
		parts := strings.SplitN(key, ":", 4)
		if len(parts) < 3 {
			continue
		}
		tableName := parts[2]
		var indexName string
		if len(parts) == 4 {
			indexName = parts[3]
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" && tableName != *input.TableName {
			continue
		}

		status, _ := insightsConfig["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}

		mode, _ := insightsConfig["ContributorInsightsMode"].(string)
		if mode == "" {
			mode = "ALL"
		}

		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": status,
			"ContributorInsightsMode":   mode,
		}

		if indexName != "" {
			summary["IndexName"] = indexName
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination
	limit := 100
	if input.MaxResults != nil && *input.MaxResults > 0 {
		limit = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// NextToken is the index to resume from
		for i, summary := range summaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				token := summaryMap["TableName"].(string)
				if indexName, ok := summaryMap["IndexName"].(string); ok {
					token += ":" + indexName
				}
				if token == *input.NextToken {
					startIndex = i + 1
					break
				}
			}
		}
	}

	endIndex := startIndex + limit
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

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			token := lastSummary["TableName"].(string)
			if indexName, ok := lastSummary["IndexName"].(string); ok {
				token += ":" + indexName
			}
			response["NextToken"] = token
		}
	}

	return s.jsonResponse(200, response)
}
