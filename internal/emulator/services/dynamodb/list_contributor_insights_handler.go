package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listContributorInsights returns a list of ContributorInsightsSummary for a table
// and all its global secondary indexes.
func (s *DynamoDBService) listContributorInsights(ctx context.Context, input *ListContributorInsightsInput) (*emulator.AWSResponse, error) {
	// List all contributor insights keys from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	summaries := []map[string]interface{}{}

	for _, key := range keys {
		// Parse tableName and optional indexName from state key.
		// Key format: "dynamodb:contributor-insights:tableName"
		// or "dynamodb:contributor-insights:tableName:indexName"
		// Splitting on ":" yields: ["dynamodb", "contributor-insights", "tableName", ...]
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}

		tableName := parts[2]
		indexName := ""
		if len(parts) >= 4 {
			indexName = strings.Join(parts[3:], ":")
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" && tableName != *input.TableName {
			continue
		}

		// Retrieve the contributor insights data
		var insightsData map[string]interface{}
		if err := s.state.Get(key, &insightsData); err != nil {
			continue
		}

		status, _ := insightsData["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}

		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": status,
		}
		if indexName != "" {
			summary["IndexName"] = indexName
		}

		summaries = append(summaries, summary)
	}

	response := map[string]interface{}{
		"ContributorInsightsSummaries": summaries,
	}

	return s.jsonResponse(200, response)
}
