package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) describeSecurityGroups(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupIds := s.parseSecurityGroupIds(params)

	// Extract filter values
	vpcIdFilter := s.extractFilterValue(params, "vpc-id")
	groupNameFilter := s.extractFilterValue(params, "group-name")

	var groups []SecurityGroup

	if len(groupIds) > 0 {
		for _, groupId := range groupIds {
			var sg SecurityGroup
			if err := s.state.Get(fmt.Sprintf("ec2:security-groups:%s", groupId), &sg); err != nil {
				return s.errorResponse(400, "InvalidGroup.NotFound", fmt.Sprintf("The security group '%s' does not exist", groupId)), nil
			}
			// When specific IDs are requested with filters, AWS returns NotFound if the SG
			// exists but doesn't match the filter (the resource doesn't exist in that context)
			if vpcIdFilter != "" && (sg.VpcId == nil || *sg.VpcId != vpcIdFilter) {
				return s.errorResponse(400, "InvalidGroup.NotFound", fmt.Sprintf("The security group '%s' does not exist in VPC '%s'", groupId, vpcIdFilter)), nil
			}
			if groupNameFilter != "" && (sg.GroupName == nil || *sg.GroupName != groupNameFilter) {
				return s.errorResponse(400, "InvalidGroup.NotFound", fmt.Sprintf("The security group '%s' does not exist with name '%s'", groupId, groupNameFilter)), nil
			}
			groups = append(groups, sg)
		}
	} else {
		keys, err := s.state.List("ec2:security-groups:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list security groups"), nil
		}

		for _, key := range keys {
			var sg SecurityGroup
			if err := s.state.Get(key, &sg); err == nil {
				// Apply vpc-id filter
				if vpcIdFilter != "" && (sg.VpcId == nil || *sg.VpcId != vpcIdFilter) {
					continue
				}
				// Apply group-name filter
				if groupNameFilter != "" && (sg.GroupName == nil || *sg.GroupName != groupNameFilter) {
					continue
				}
				groups = append(groups, sg)
			}
		}
	}

	return s.describeSecurityGroupsResponse(groups)
}
