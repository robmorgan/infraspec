package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports lists completed exports within the past 90 days.
func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// List all export keys from state
	keys, err := s.state.List("dynamodb:export:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list exports"), nil
	}

	summaries := []map[string]interface{}{}

	for _, key := range keys {
		var exportData map[string]interface{}
		if err := s.state.Get(key, &exportData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, _ := exportData["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build export summary
		summary := map[string]interface{}{}
		if exportArn, ok := exportData["ExportArn"].(string); ok {
			summary["ExportArn"] = exportArn
		}
		if exportStatus, ok := exportData["ExportStatus"].(string); ok {
			summary["ExportStatus"] = exportStatus
		}
		if exportType, ok := exportData["ExportType"].(string); ok {
			summary["ExportType"] = exportType
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination
	pageSize := len(summaries)
	if input.MaxResults != nil && int(*input.MaxResults) > 0 && int(*input.MaxResults) < pageSize {
		pageSize = int(*input.MaxResults)
	}

	startIdx := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// NextToken is the ExportArn of the last returned item
		for i, sm := range summaries {
			if arn, ok := sm["ExportArn"].(string); ok && arn == *input.NextToken {
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
		"ExportSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIdx < len(summaries) && len(paginatedSummaries) > 0 {
		if last, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := last["ExportArn"].(string); ok {
				response["NextToken"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}
