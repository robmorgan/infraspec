package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports lists completed exports within the past 90 days.
func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// List all export keys from state
	keys, err := s.state.List("dynamodb:export:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list exports"), nil
	}

	exportSummaries := []interface{}{}

	for _, key := range keys {
		var exportData map[string]interface{}
		if err := s.state.Get(key, &exportData); err != nil {
			continue
		}

		// Filter by TableArn if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, _ := exportData["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build export summary with required fields
		summary := map[string]interface{}{}

		if exportArn, ok := exportData["ExportArn"].(string); ok {
			summary["ExportArn"] = exportArn
		} else {
			// Derive ARN from the state key
			// Key format: dynamodb:export:<exportArn>
			exportArn := strings.TrimPrefix(key, "dynamodb:export:")
			if exportArn != "" {
				summary["ExportArn"] = exportArn
			}
		}

		if exportStatus, ok := exportData["ExportStatus"].(string); ok {
			summary["ExportStatus"] = exportStatus
		}

		if exportType, ok := exportData["ExportType"].(string); ok {
			summary["ExportType"] = exportType
		}

		exportSummaries = append(exportSummaries, summary)
	}

	// Apply MaxResults pagination if specified
	maxResults := len(exportSummaries)
	if input.MaxResults != nil && *input.MaxResults > 0 && int(*input.MaxResults) < maxResults {
		maxResults = int(*input.MaxResults)
	}

	paginatedSummaries := exportSummaries
	var nextToken *string
	if maxResults < len(exportSummaries) {
		paginatedSummaries = exportSummaries[:maxResults]
		// Indicate more results available
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if arn, ok := lastSummary["ExportArn"].(string); ok {
				nextToken = &arn
			}
		}
	}

	response := map[string]interface{}{
		"ExportSummaries": paginatedSummaries,
	}

	if nextToken != nil {
		response["NextToken"] = *nextToken
	}

	return s.jsonResponse(200, response)
}
