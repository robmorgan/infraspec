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
	// List all contributor insights entries from state
	keys, err := s.state.List("dynamodb:contributor-insights:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list contributor insights"), nil
	}

	summaries := []interface{}{}

	for _, key := range keys {
		// Filter by table name if specified
		// Key format: "dynamodb:contributor-insights:<tableName>" or
		//             "dynamodb:contributor-insights:<tableName>:<indexName>"
		if input.TableName != nil && *input.TableName != "" {
			prefix := fmt.Sprintf("dynamodb:contributor-insights:%s", *input.TableName)
			if !strings.HasPrefix(key, prefix) {
				continue
			}
		}

		var insightsConfig map[string]interface{}
		if err := s.state.Get(key, &insightsConfig); err != nil {
			continue
		}

		// Extract tableName and optionally indexName from the key
		// key = "dynamodb:contributor-insights:<tableName>" or
		//       "dynamodb:contributor-insights:<tableName>:<indexName>"
		parts := strings.SplitN(key, ":", 4)
		if len(parts) < 3 {
			continue
		}
		tableName := parts[2]

		status, _ := insightsConfig["ContributorInsightsStatus"].(string)
		if status == "" {
			status = "DISABLED"
		}

		summary := map[string]interface{}{
			"TableName":                 tableName,
			"ContributorInsightsStatus": status,
		}

		// Include IndexName if present
		if len(parts) == 4 && parts[3] != "" {
			summary["IndexName"] = parts[3]
		}

		// Include LastUpdateDateTime if present
		if lastUpdateTime, ok := insightsConfig["LastUpdateDateTime"]; ok {
			summary["LastUpdateDateTime"] = lastUpdateTime
		}

		summaries = append(summaries, summary)
	}

	response := map[string]interface{}{
		"ContributorInsightsSummaries": summaries,
	}

	return s.jsonResponse(200, response)
}
