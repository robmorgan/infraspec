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

	summaries := []ImportSummary{}

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

		// Build import summary
		summary := ImportSummary{}

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary.ImportArn = &importArn
		}

		if importStatus, ok := importData["ImportStatus"].(string); ok {
			summary.ImportStatus = ImportStatus(importStatus)
		}

		if tableArn, ok := importData["TableArn"].(string); ok {
			summary.TableArn = &tableArn
		}

		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary.InputFormat = InputFormat(inputFormat)
		}

		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary.CloudWatchLogGroupArn = &cloudWatchLogGroupArn
		}

		// Add S3BucketSource if present
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

		// Add timestamps if present (they would be stored as float64 from JSON)
		// Note: time.Time conversion would happen here if needed

		summaries = append(summaries, summary)
	}

	// Apply pagination
	pageSize := 100 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// If NextToken is provided, it would contain pagination info
	// For simplicity in this emulator, we'll just use basic pagination

	endIndex := startIndex + pageSize
	if endIndex > len(summaries) {
		endIndex = len(summaries)
	}

	paginatedSummaries := []ImportSummary{}
	if startIndex < len(summaries) {
		paginatedSummaries = summaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(summaries) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}
