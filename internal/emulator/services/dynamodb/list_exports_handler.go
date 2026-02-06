package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports lists completed exports within the past 90 days.
func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// List all exports from state
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

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if tableArn, ok := exportData["TableArn"].(string); ok {
				if tableArn != *input.TableArn {
					continue
				}
			} else {
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

		exportSummaries = append(exportSummaries, summary)
	}

	// Apply pagination if specified
	maxResults := 25 // Default max results for ListExports
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// Note: For simplicity, we're not implementing NextToken pagination in this basic implementation
	// In a real implementation, NextToken would encode pagination state

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(exportSummaries) {
		endIndex = len(exportSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(exportSummaries) {
		paginatedSummaries = exportSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ExportSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(exportSummaries) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
