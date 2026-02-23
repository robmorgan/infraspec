package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all import keys from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	summaries := []map[string]interface{}{}

	for _, key := range keys {
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, _ := importData["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build import summary from stored data
		summary := map[string]interface{}{}
		if importArn, ok := importData["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		}
		if importStatus, ok := importData["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		}
		if tableArn, ok := importData["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
		}
		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
		}
		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination
	pageSize := len(summaries)
	if input.PageSize != nil && int(*input.PageSize) > 0 && int(*input.PageSize) < pageSize {
		pageSize = int(*input.PageSize)
	}

	startIdx := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// NextToken is the ImportArn of the last returned item
		for i, sm := range summaries {
			if arn, ok := sm["ImportArn"].(string); ok && arn == *input.NextToken {
				startIdx = i + 1
				break
			}
		}
	}

	endIdx := startIdx + pageSize
	if endIdx > len(summaries) {
		endIdx = len(summaries)
	}

	paginatedSummaries := []interface{}{}
	if startIdx < len(summaries) {
		for _, sm := range summaries[startIdx:endIdx] {
			paginatedSummaries = append(paginatedSummaries, sm)
		}
	}

	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIdx < len(summaries) && len(paginatedSummaries) > 0 {
		if last, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := last["ImportArn"].(string); ok {
				response["NextToken"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}
