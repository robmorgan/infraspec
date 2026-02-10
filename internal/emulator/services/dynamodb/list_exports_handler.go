package dynamodb

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) listExports(ctx context.Context, input *ListExportsInput) (*emulator.AWSResponse, error) {
	// Optional: Validate table exists if TableArn is specified
	if input.TableArn != nil && *input.TableArn != "" {
		// Extract table name from ARN
		// ARN format: arn:aws:dynamodb:region:account:table/tablename
		parts := strings.Split(*input.TableArn, "/")
		if len(parts) < 2 {
			return s.errorResponse(400, "ValidationException", "Invalid TableArn format"), nil
		}
		tableName := parts[len(parts)-1]
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}
	}

	// List all exports
	// If TableArn is specified, filter by table; otherwise return all
	prefix := "dynamodb:export:"
	var filterTableName string
	if input.TableArn != nil && *input.TableArn != "" {
		parts := strings.Split(*input.TableArn, "/")
		if len(parts) >= 2 {
			filterTableName = parts[len(parts)-1]
		}
	}

	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list exports"), nil
	}

	// Build list of export summaries
	summaries := []map[string]interface{}{}
	for _, key := range keys {
		var exportDesc map[string]interface{}
		if err := s.state.Get(key, &exportDesc); err != nil {
			continue
		}

		// Filter by table if TableArn is specified
		if filterTableName != "" {
			exportTableArn, _ := exportDesc["TableArn"].(string)
			if !strings.Contains(exportTableArn, "/"+filterTableName) {
				continue
			}
		}

		// Build export summary
		summary := map[string]interface{}{}

		if exportArn, ok := exportDesc["ExportArn"].(string); ok {
			summary["ExportArn"] = exportArn
		}

		if exportStatus, ok := exportDesc["ExportStatus"].(string); ok {
			summary["ExportStatus"] = exportStatus
		} else {
			summary["ExportStatus"] = "COMPLETED"
		}

		if exportType, ok := exportDesc["ExportType"].(string); ok {
			summary["ExportType"] = exportType
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination if needed
	maxResults := 100 // Default max results
	if input.MaxResults != nil && *input.MaxResults > 0 {
		maxResults = int(*input.MaxResults)
	}

	startIndex := 0
	// Simple pagination: NextToken is the start index as a string
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, you would parse the token
		// For simplicity, we'll just return all results
	}

	endIndex := startIndex + maxResults
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := summaries[startIndex:endIndex]

	response := map[string]interface{}{
		"ExportSummaries": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		response["NextToken"] = fmt.Sprintf("%d", endIndex)
	}

	return s.jsonResponse(200, response)
}
