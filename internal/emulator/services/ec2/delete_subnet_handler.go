package ec2

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) deleteSubnet(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	subnetId, ok := params["SubnetId"].(string)
	if !ok || subnetId == "" {
		return s.errorResponse(400, "MissingParameter", "SubnetId is required"), nil
	}

	if subnetId == "subnet-default" {
		return s.errorResponse(400, "OperationNotPermitted", "Cannot delete the default subnet"), nil
	}

	var subnet Subnet
	if err := s.state.Get(fmt.Sprintf("ec2:subnets:%s", subnetId), &subnet); err != nil {
		return s.errorResponse(400, "InvalidSubnetID.NotFound", fmt.Sprintf("The subnet ID '%s' does not exist", subnetId)), nil
	}

	// Check graph-based dependencies (instances, NAT gateways, etc.)
	if canDelete, dependents := s.canDeleteResource("subnet", subnetId); !canDelete {
		return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("The subnet '%s' has dependencies and cannot be deleted: %v", subnetId, dependents)), nil
	}

	// Unregister from graph
	if err := s.unregisterResource("subnet", subnetId); err != nil {
		return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("Cannot delete subnet: %v", err)), nil
	}

	s.state.Delete(fmt.Sprintf("ec2:subnets:%s", subnetId))

	return s.deleteSubnetResponse()
}
