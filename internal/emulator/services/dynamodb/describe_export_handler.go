package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeExport describes an existing table export.
func (s *DynamoDBService) describeExport(ctx context.Context, input *DescribeExportInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ExportArn == nil || *input.ExportArn == "" {
		return s.errorResponse(400, "ValidationException", "ExportArn is required"), nil
	}

	exportArn := *input.ExportArn

	// Check if export exists in state
	stateKey := fmt.Sprintf("dynamodb:export:%s", exportArn)
	var exportDesc map[string]interface{}
	if err := s.state.Get(stateKey, &exportDesc); err != nil {
		return s.errorResponse(400, "ExportNotFoundException", fmt.Sprintf("Export not found: %s", exportArn)), nil
	}

	// Return export description
	response := map[string]interface{}{
		"ExportDescription": exportDesc,
	}

	return s.jsonResponse(200, response)
}
