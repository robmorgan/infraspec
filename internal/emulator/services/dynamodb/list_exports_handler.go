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

	summaries := []ExportSummary{}

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
		summary := ExportSummary{}

		if exportArn, ok := exportData["ExportArn"].(string); ok {
			summary.ExportArn = &exportArn
		}

		if exportStatus, ok := exportData["ExportStatus"].(string); ok {
			summary.ExportStatus = ExportStatus(exportStatus)
		}

		if exportType, ok := exportData["ExportType"].(string); ok {
			summary.ExportType = ExportType(exportType)
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// If NextToken is provided, it would contain pagination info
	// For simplicity in this emulator, we'll just use basic pagination

	endIndex := startIndex + maxResults
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := []ExportSummary{}
	if startIndex < len(summaries) {
		paginatedSummaries = summaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ExportSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
