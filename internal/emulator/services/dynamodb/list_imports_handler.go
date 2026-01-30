package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all imports from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	importSummaries := []interface{}{}

	for _, key := range keys {
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if tableArnInImport, ok := importData["TableArn"].(string); ok {
				if tableArnInImport != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		importSummaries = append(importSummaries, importData)
	}

	// Apply pagination if specified
	pageSize := 25 // Default page size for ListImports
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// For simplicity, NextToken represents the start index as a string
		// In a production system, this would be a properly encoded token
		var tokenIndex int
		if _, err := fmt.Sscanf(*input.NextToken, "%d", &tokenIndex); err == nil {
			startIndex = tokenIndex
		}
	}

	// Apply pagination
	endIndex := startIndex + pageSize
	if endIndex > len(importSummaries) {
		endIndex = len(importSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(importSummaries) {
		paginatedSummaries = importSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(importSummaries) {
		response["NextToken"] = fmt.Sprintf("%d", endIndex)
	}

	return s.jsonResponse(200, response)
}
