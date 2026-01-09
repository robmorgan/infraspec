package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all imports from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	importSummaries := []ImportSummary{}

	for _, key := range keys {
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if tableArn, ok := importData["TableArn"].(string); ok {
				if tableArn != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		// Build summary
		summary := ImportSummary{}

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary.ImportArn = &importArn
		}

		if status, ok := importData["ImportStatus"].(string); ok {
			summary.ImportStatus = ImportStatus(status)
		}

		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary.InputFormat = InputFormat(inputFormat)
		}

		if tableArn, ok := importData["TableArn"].(string); ok {
			summary.TableArn = &tableArn
		}

		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary.CloudWatchLogGroupArn = &cloudWatchLogGroupArn
		}

		// Handle S3BucketSource if present
		if s3BucketSource, ok := importData["S3BucketSource"].(map[string]interface{}); ok {
			s3Source := &S3BucketSource{}
			if s3Bucket, ok := s3BucketSource["S3Bucket"].(string); ok {
				s3Source.S3Bucket = &s3Bucket
			}
			if s3KeyPrefix, ok := s3BucketSource["S3KeyPrefix"].(string); ok {
				s3Source.S3KeyPrefix = &s3KeyPrefix
			}
			if s3BucketOwner, ok := s3BucketSource["S3BucketOwner"].(string); ok {
				s3Source.S3BucketOwner = &s3BucketOwner
			}
			summary.S3BucketSource = s3Source
		}

		importSummaries = append(importSummaries, summary)
	}

	// Handle pagination
	pageSize := 100 // Default
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// For simplicity, we're not implementing full NextToken logic
	// In a production emulator, you would need to properly handle pagination tokens

	endIndex := startIndex + pageSize
	if endIndex > len(importSummaries) {
		endIndex = len(importSummaries)
	}

	paginatedSummaries := []ImportSummary{}
	if startIndex < len(importSummaries) {
		paginatedSummaries = importSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(importSummaries) {
		response["NextToken"] = "has-more-results"
	}

	return s.jsonResponse(200, response)
}
