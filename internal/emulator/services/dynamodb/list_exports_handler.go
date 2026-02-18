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
		var exportDesc map[string]interface{}
		if err := s.state.Get(key, &exportDesc); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, _ := exportDesc["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build export summary from export description
		summary := map[string]interface{}{}

		if exportArn, ok := exportDesc["ExportArn"].(string); ok {
			summary["ExportArn"] = exportArn
		}
		if exportStatus, ok := exportDesc["ExportStatus"].(string); ok {
			summary["ExportStatus"] = exportStatus
		}
		if exportType, ok := exportDesc["ExportType"].(string); ok {
			summary["ExportType"] = exportType
		}
		if tableArn, ok := exportDesc["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
		}
		if exportTime, ok := exportDesc["ExportTime"]; ok {
			summary["ExportTime"] = exportTime
		}

		exportSummaries = append(exportSummaries, summary)
	}

	// Apply MaxResults pagination
	maxResults := 25 // DynamoDB default for ListExports
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	// Handle NextToken pagination
	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		for i, summary := range exportSummaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if arn, ok := summaryMap["ExportArn"].(string); ok {
					// Token is the export ARN prefix (last part after "export:")
					arnSuffix := strings.TrimPrefix(arn, "arn:aws:dynamodb:")
					if arnSuffix == *input.NextToken || arn == *input.NextToken {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

	endIndex := startIndex + maxResults
	if endIndex > len(exportSummaries) {
		endIndex = len(exportSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(exportSummaries) {
		paginatedSummaries = exportSummaries[startIndex:endIndex]
	}

	response := map[string]interface{}{
		"ExportSummaries": paginatedSummaries,
	}

	// Set NextToken if there are more results
	if endIndex < len(exportSummaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := lastSummary["ExportArn"].(string); ok {
				response["NextToken"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}
