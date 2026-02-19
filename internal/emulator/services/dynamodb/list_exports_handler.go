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

	exportSummaries := []interface{}{}

	for _, key := range keys {
		var exportDesc map[string]interface{}
		if err := s.state.Get(key, &exportDesc); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, ok := exportDesc["TableArn"].(string)
			if !ok || tableArn != *input.TableArn {
				continue
			}
		}

		// Build export summary with key fields
		summary := map[string]interface{}{}

		if exportArn, ok := exportDesc["ExportArn"]; ok {
			summary["ExportArn"] = exportArn
		}
		if exportStatus, ok := exportDesc["ExportStatus"]; ok {
			summary["ExportStatus"] = exportStatus
		}
		if tableArn, ok := exportDesc["TableArn"]; ok {
			summary["TableArn"] = tableArn
		}
		if exportTime, ok := exportDesc["ExportTime"]; ok {
			summary["ExportTime"] = exportTime
		}

		exportSummaries = append(exportSummaries, summary)
	}

	response := map[string]interface{}{
		"ExportSummaries": exportSummaries,
	}

	return s.jsonResponse(200, response)
}
