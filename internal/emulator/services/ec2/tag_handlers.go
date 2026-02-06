package ec2

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) createTags(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	resourceIds := s.parseResourceIds(params)
	if len(resourceIds) == 0 {
		return s.errorResponse(400, "MissingParameter", "ResourceId is required"), nil
	}

	tags := s.parseTags(params)
	if len(tags) == 0 {
		return s.errorResponse(400, "MissingParameter", "Tags are required"), nil
	}

	for _, resourceId := range resourceIds {
		var existingTags []Tag
		s.state.Get(fmt.Sprintf("ec2:tags:%s", resourceId), &existingTags)

		// Merge tags
		tagMap := make(map[string]string)
		for _, tag := range existingTags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}
		for _, tag := range tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Convert back to tag list
		mergedTags := make([]Tag, 0)
		for k, v := range tagMap {
			key := k
			value := v
			mergedTags = append(mergedTags, Tag{Key: &key, Value: &value})
		}

		s.state.Set(fmt.Sprintf("ec2:tags:%s", resourceId), mergedTags)
	}

	return s.createTagsResponse()
}

func (s *EC2Service) describeTags(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Parse filters from request
	resourceIdFilters := s.parseFilterValues(params, "resource-id")
	resourceTypeFilters := s.parseFilterValues(params, "resource-type")
	keyFilters := s.parseFilterValues(params, "key")
	valueFilters := s.parseFilterValues(params, "value")

	var allTags []TagDescription

	keys, err := s.state.List("ec2:tags:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list tags"), nil
	}

	for _, key := range keys {
		resourceId := strings.TrimPrefix(key, "ec2:tags:")

		// Apply resource-id filter
		if len(resourceIdFilters) > 0 && !containsString(resourceIdFilters, resourceId) {
			continue
		}

		resourceType := getResourceType(resourceId)

		// Apply resource-type filter
		if len(resourceTypeFilters) > 0 && !containsString(resourceTypeFilters, string(resourceType)) {
			continue
		}

		var tags []Tag
		if err := s.state.Get(key, &tags); err == nil {
			for _, tag := range tags {
				// Apply key filter
				if len(keyFilters) > 0 && (tag.Key == nil || !containsString(keyFilters, *tag.Key)) {
					continue
				}

				// Apply value filter
				if len(valueFilters) > 0 && (tag.Value == nil || !containsString(valueFilters, *tag.Value)) {
					continue
				}

				allTags = append(allTags, TagDescription{
					ResourceId:   &resourceId,
					ResourceType: resourceType,
					Key:          tag.Key,
					Value:        tag.Value,
				})
			}
		}
	}

	return s.describeTagsResponse(allTags)
}

// parseFilterValues extracts filter values for a given filter name from params
func (s *EC2Service) parseFilterValues(params map[string]interface{}, filterName string) []string {
	var values []string

	// Check for Filter.N.Name / Filter.N.Value.M format
	for i := 1; i <= 20; i++ {
		nameKey := fmt.Sprintf("Filter.%d.Name", i)
		name, hasName := params[nameKey].(string)
		if !hasName {
			break
		}

		if name == filterName {
			// Get all values for this filter
			for j := 1; j <= 100; j++ {
				valueKey := fmt.Sprintf("Filter.%d.Value.%d", i, j)
				value, hasValue := params[valueKey].(string)
				if !hasValue {
					break
				}
				values = append(values, value)
			}
		}
	}

	return values
}

// containsString checks if a slice contains a string
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (s *EC2Service) deleteTags(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	resourceIds := s.parseResourceIds(params)
	if len(resourceIds) == 0 {
		return s.errorResponse(400, "MissingParameter", "ResourceId is required"), nil
	}

	tagsToDelete := s.parseTags(params)

	for _, resourceId := range resourceIds {
		var existingTags []Tag
		s.state.Get(fmt.Sprintf("ec2:tags:%s", resourceId), &existingTags)

		if len(tagsToDelete) == 0 {
			// Delete all tags
			s.state.Delete(fmt.Sprintf("ec2:tags:%s", resourceId))
		} else {
			// Delete specific tags
			deleteKeys := make(map[string]bool)
			for _, tag := range tagsToDelete {
				if tag.Key != nil {
					deleteKeys[*tag.Key] = true
				}
			}

			remainingTags := make([]Tag, 0)
			for _, tag := range existingTags {
				if tag.Key != nil && !deleteKeys[*tag.Key] {
					remainingTags = append(remainingTags, tag)
				}
			}

			if len(remainingTags) > 0 {
				s.state.Set(fmt.Sprintf("ec2:tags:%s", resourceId), remainingTags)
			} else {
				s.state.Delete(fmt.Sprintf("ec2:tags:%s", resourceId))
			}
		}
	}

	return s.deleteTagsResponse()
}
