package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) describeVpcs(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	vpcIds := s.parseVpcIds(params)

	var vpcs []Vpc

	if len(vpcIds) > 0 {
		for _, vpcId := range vpcIds {
			var vpc Vpc
			if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
				return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
			}
			// Merge in tags from separate tag storage
			s.mergeResourceTags(&vpc.Tags, vpcId)
			vpcs = append(vpcs, vpc)
		}
	} else {
		keys, err := s.state.List("ec2:vpcs:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list VPCs"), nil
		}

		for _, key := range keys {
			var vpc Vpc
			if err := s.state.Get(key, &vpc); err == nil {
				// Merge in tags from separate tag storage
				if vpc.VpcId != nil {
					s.mergeResourceTags(&vpc.Tags, *vpc.VpcId)
				}
				vpcs = append(vpcs, vpc)
			}
		}
	}

	return s.describeVpcsResponse(vpcs)
}

// mergeResourceTags retrieves tags from the separate tag storage and merges them into the target slice
func (s *EC2Service) mergeResourceTags(target *[]Tag, resourceId string) {
	var storedTags []Tag
	if err := s.state.Get(fmt.Sprintf("ec2:tags:%s", resourceId), &storedTags); err == nil {
		// Create a map to merge tags (stored tags override existing)
		tagMap := make(map[string]string)
		if target != nil {
			for _, tag := range *target {
				if tag.Key != nil && tag.Value != nil {
					tagMap[*tag.Key] = *tag.Value
				}
			}
		}
		for _, tag := range storedTags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Convert back to tag slice
		mergedTags := make([]Tag, 0, len(tagMap))
		for k, v := range tagMap {
			key := k
			value := v
			mergedTags = append(mergedTags, Tag{Key: &key, Value: &value})
		}
		*target = mergedTags
	}
}
