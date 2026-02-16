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
			if tableArnInExport, ok := exportData["TableArn"].(string); ok {
				if tableArnInExport != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		// Build export summary
		summary := map[string]interface{}{}

		// Add optional fields if present
		if exportArn, ok := exportData["ExportArn"]; ok {
			summary["ExportArn"] = exportArn
		}
		if exportStatus, ok := exportData["ExportStatus"]; ok {
			summary["ExportStatus"] = exportStatus
		}
		if exportType, ok := exportData["ExportType"]; ok {
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
	// For NextToken pagination, we would decode it to get the start index
	// For simplicity, we'll use a basic implementation

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
		// In a real implementation, this would be an encoded token
		response["NextToken"] = "nextPageToken"
	}

	return s.jsonResponse(200, response)
}
