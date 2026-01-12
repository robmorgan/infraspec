package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports lists completed exports within the past 90 days.
func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// Optional table ARN filter
	var tableArn string
	if input.TableArn != nil {
		tableArn = *input.TableArn
	}

	// List all exports from state
	keys, err := s.state.List("dynamodb:export:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list exports"), nil
	}

	var exportSummaries []interface{}

	for _, key := range keys {
		var exportData map[string]interface{}
		if err := s.state.Get(key, &exportData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if tableArn != "" {
			if exportTableArn, ok := exportData["TableArn"].(string); ok {
				if exportTableArn != tableArn {
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

	// Build response with pagination support
	response := map[string]interface{}{
		"ExportSummaries": exportSummaries,
	}

	// Handle pagination if MaxResults is specified
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults := int(*input.MaxResults)
		startIndex := 0

		// If NextToken is provided, decode it to get start index
		// For simplicity, we'll use a simple index-based pagination
		if input.NextToken != nil && *input.NextToken != "" {
			// In a real implementation, NextToken would be decoded
			// For now, we'll keep it simple and not implement pagination offset
		}

		// Apply pagination
		if len(exportSummaries) > maxResults {
			response["ExportSummaries"] = exportSummaries[startIndex:maxResults]
			response["NextToken"] = "next-page-token" // Simplified token
		}
	}

	return s.jsonResponse(200, response)
}
