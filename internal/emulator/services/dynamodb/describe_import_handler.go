package dynamodb

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeImport represents the properties of the import.
func (s *DynamoDBService) describeImport(ctx context.Context, input *DescribeImportInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ImportArn == nil || *input.ImportArn == "" {
		return s.errorResponse(400, "ValidationException", "ImportArn is required"), nil
	}

	importArn := *input.ImportArn

	// Check if import exists in state
	stateKey := fmt.Sprintf("dynamodb:import:%s", importArn)
	var importDesc map[string]interface{}
	if err := s.state.Get(stateKey, &importDesc); err != nil {
		return s.errorResponse(400, "ImportNotFoundException", fmt.Sprintf("Import not found: %s", importArn)), nil
	}

	// Return import description
	response := map[string]interface{}{
		"ImportTableDescription": importDesc,
	}

	return s.jsonResponse(200, response)
}
