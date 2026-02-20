package dynamodb

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table and all its
// global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	summaries := []interface{}{}

	// If a specific table name is provided, only return insights for that table
	if input.TableName != nil && *input.TableName != "" {
		tableName := *input.TableName

		// Verify the table exists
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}

		// Get contributor insights for the table itself
		insightsKey := fmt.Sprintf("dynamodb:contributor-insights:%s", tableName)
		var insightsConfig map[string]interface{}
		if err := s.state.Get(insightsKey, &insightsConfig); err != nil {
			insightsConfig = map[string]interface{}{
				"ContributorInsightsStatus": "DISABLED",
			}
		}

		status, _ := insightsConfig["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}

		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": status,
		}
		summaries = append(summaries, summary)

		// Get contributor insights for all GSIs of the table
		gsiKeys, err := s.state.List(fmt.Sprintf("dynamodb:contributor-insights:%s:", tableName))
		if err == nil {
			for _, key := range gsiKeys {
				// Extract the index name from the key
				// Key format: dynamodb:contributor-insights:<tableName>:<indexName>
				prefix := fmt.Sprintf("dynamodb:contributor-insights:%s:", tableName)
				indexName := strings.TrimPrefix(key, prefix)
				if indexName == "" {
					continue
				}

				var gsiConfig map[string]interface{}
				if err := s.state.Get(key, &gsiConfig); err != nil {
					continue
				}

				gsiStatus, _ := gsiConfig["ContributorInsightsStatus"].(string)
				if gsiStatus == "" {
					gsiStatus = "DISABLED"
				}

				gsiSummary := map[string]interface{}{
					"TableName":                 tableName,
					"IndexName":                 indexName,
					"ContributorInsightsStatus": gsiStatus,
				}
				summaries = append(summaries, gsiSummary)
			}
		}
	} else {
		// No table name filter - list insights for all tables
		tableKeys, err := s.state.List("dynamodb:contributor-insights:")
		if err != nil {
			return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
		}

		for _, key := range tableKeys {
			// Skip GSI-specific keys (they have an extra colon segment)
			suffix := strings.TrimPrefix(key, "dynamodb:contributor-insights:")
			parts := strings.Split(suffix, ":")
			if len(parts) > 1 {
				// This is a GSI-specific entry; skip unless we find it from table enumeration
				continue
			}

			tableName := parts[0]

			var insightsConfig map[string]interface{}
			if err := s.state.Get(key, &insightsConfig); err != nil {
				continue
			}

			status, _ := insightsConfig["ContributorInsightsStatus"].(string)
			if status == "" {
				status = "DISABLED"
			}

			summary := map[string]interface{}{
				"TableName":                 tableName,
				"ContributorInsightsStatus": status,
			}
			summaries = append(summaries, summary)
		}
	}

	// Apply MaxResults pagination if specified
	maxResults := len(summaries)
	if input.MaxResults != nil && *input.MaxResults > 0 && int(*input.MaxResults) < maxResults {
		maxResults = int(*input.MaxResults)
	}

	paginatedSummaries := summaries
	var nextToken *string
	if maxResults < len(summaries) {
		paginatedSummaries = summaries[:maxResults]
		// Return a next token indicating more results are available
		token := fmt.Sprintf("page:%d", maxResults)
		nextToken = &token
	}

	response := map[string]interface{}{
		"ContributorInsightsSummaries": paginatedSummaries,
	}

	if nextToken != nil {
		response["NextToken"] = *nextToken
	}

	return s.jsonResponse(200, response)
}
