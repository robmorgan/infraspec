package dynamodb

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
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

	// List all imports
	// If TableArn is specified, filter by table; otherwise return all
	prefix := "dynamodb:import:"
	var filterTableName string
	if input.TableArn != nil && *input.TableArn != "" {
		parts := strings.Split(*input.TableArn, "/")
		if len(parts) >= 2 {
			filterTableName = parts[len(parts)-1]
		}
	}

	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	// Build list of import summaries
	summaries := []map[string]interface{}{}
	for _, key := range keys {
		var importDesc map[string]interface{}
		if err := s.state.Get(key, &importDesc); err != nil {
			continue
		}

		// Filter by table if TableArn is specified
		if filterTableName != "" {
			importTableArn, _ := importDesc["TableArn"].(string)
			if !strings.Contains(importTableArn, "/"+filterTableName) {
				continue
			}
		}

		// Build import summary
		summary := map[string]interface{}{}

		if importArn, ok := importDesc["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		}

		if importStatus, ok := importDesc["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		} else {
			summary["ImportStatus"] = "COMPLETED"
		}

		if inputFormat, ok := importDesc["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
		}

		if cloudWatchLogGroupArn, ok := importDesc["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}

		if endTime, ok := importDesc["EndTime"].(float64); ok {
			summary["EndTime"] = endTime
		}

		if startTime, ok := importDesc["StartTime"].(float64); ok {
			summary["StartTime"] = startTime
		}

		if tableArn, ok := importDesc["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
		}

		summaries = append(summaries, summary)
	}

	// Apply pagination if needed
	pageSize := 100 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// Simple pagination: NextToken is the start index as a string
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, you would parse the token
		// For simplicity, we'll just return all results
	}

	endIndex := startIndex + pageSize
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := summaries[startIndex:endIndex]

	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		response["NextToken"] = fmt.Sprintf("%d", endIndex)
	}

	return s.jsonResponse(200, response)
}
