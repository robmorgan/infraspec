package dynamodb

import (
	"context"
	"strings"

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
	maxResults := 25 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded to find the start position
		// For the emulator, we'll use a simple approach
		for i, summary := range exportSummaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if exportArn, ok := summaryMap["ExportArn"].(string); ok {
					if exportArn == *input.NextToken {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

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
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if exportArn, ok := lastSummary["ExportArn"].(string); ok {
				response["NextToken"] = exportArn
			}
		}
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract export ID from export key
func extractExportIDFromKey(key string) string {
	// Key format: "dynamodb:export:exportid"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
