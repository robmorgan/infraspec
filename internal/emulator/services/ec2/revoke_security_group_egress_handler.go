package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) revokeSecurityGroupEgress(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupId, ok := params["GroupId"].(string)
	if !ok || groupId == "" {
		return s.errorResponse(400, "MissingParameter", "GroupId is required"), nil
	}

	var sg SecurityGroup
	if err := s.state.Get(fmt.Sprintf("ec2:security-groups:%s", groupId), &sg); err != nil {
		return s.errorResponse(400, "InvalidGroup.NotFound", fmt.Sprintf("The security group '%s' does not exist", groupId)), nil
	}

	// Parse rules to revoke and remove matching ones
	rulesToRevoke := s.parseIpPermissions(params)
	sg.IpPermissionsEgress = s.removeMatchingRules(sg.IpPermissionsEgress, rulesToRevoke)

	if err := s.state.Set(fmt.Sprintf("ec2:security-groups:%s", groupId), &sg); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update security group"), nil
	}

	return s.revokeSecurityGroupEgressResponse()
}
