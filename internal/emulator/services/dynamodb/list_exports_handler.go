package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports returns a list of DynamoDB table export jobs.
func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// List all exports from state
	keys, err := s.state.List("dynamodb:export:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list exports"), nil
	}

	exportSummaries := []ExportSummary{}

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

		// Build summary
		summary := ExportSummary{}

		if exportArn, ok := exportData["ExportArn"].(string); ok {
			summary.ExportArn = &exportArn
		}

		if status, ok := exportData["ExportStatus"].(string); ok {
			summary.ExportStatus = ExportStatus(status)
		}

		if exportType, ok := exportData["ExportType"].(string); ok {
			summary.ExportType = ExportType(exportType)
		}

		exportSummaries = append(exportSummaries, summary)
	}

	// Apply pagination if specified
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, decode the token to find the start index
		// For simplicity, we'll skip pagination token parsing in the emulator
	}

	// Apply pagination
	endIndex := startIndex + maxResults
	if endIndex > len(exportSummaries) {
		endIndex = len(exportSummaries)
	}

	paginatedSummaries := []ExportSummary{}
	if startIndex < len(exportSummaries) {
		paginatedSummaries = exportSummaries[startIndex:endIndex]
	}

	// Build response
	response := ListExportsOutput{
		ExportSummaries: paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(exportSummaries) {
		nextToken := "next-token-placeholder"
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
