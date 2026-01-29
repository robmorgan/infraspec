package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listExports returns a list of completed exports within the past 90 days.
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
	// Note: NextToken would be used for pagination in a real implementation
	// For simplicity, we'll just apply the max results limit

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
		nextToken := "has-more-results" // Simplified token
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
