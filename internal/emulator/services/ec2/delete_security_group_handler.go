package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) deleteSecurityGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupId, ok := params["GroupId"].(string)
	if !ok || groupId == "" {
		return s.errorResponse(400, "MissingParameter", "GroupId is required"), nil
	}

	if groupId == "sg-default" {
		return s.errorResponse(400, "CannotDelete", "Cannot delete the default security group"), nil
	}

	var sg SecurityGroup
	if err := s.state.Get(fmt.Sprintf("ec2:security-groups:%s", groupId), &sg); err != nil {
		return s.errorResponse(400, "InvalidGroup.NotFound", fmt.Sprintf("The security group '%s' does not exist", groupId)), nil
	}

	// Check graph-based dependencies (instances, ENIs, etc.)
	if canDelete, dependents := s.canDeleteResource("security-group", groupId); !canDelete {
		return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("The security group '%s' has dependencies and cannot be deleted: %v", groupId, dependents)), nil
	}

	// Unregister from graph
	if err := s.unregisterResource("security-group", groupId); err != nil {
		return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("Cannot delete security group: %v", err)), nil
	}

	s.state.Delete(fmt.Sprintf("ec2:security-groups:%s", groupId))

	return s.deleteSecurityGroupResponse()
}
