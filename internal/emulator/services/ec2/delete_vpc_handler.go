package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

func (s *EC2Service) deleteVpc(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	if vpcId == "vpc-default" {
		return s.errorResponse(400, "OperationNotPermitted", "Cannot delete the default VPC"), nil
	}

	var vpc Vpc
	if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
	}

	// Check graph-based dependencies (subnets, etc.)
	// Note: Route tables and default security groups are automatically deleted with the VPC,
	// so they're not blocking dependencies
	if canDelete, dependents := s.canDeleteResource("vpc", vpcId); !canDelete {
		// Filter out auto-deleted resources from dependents
		var filteredDependents []graph.ResourceID
		for _, dep := range dependents {
			// Route tables are auto-deleted
			if dep.Type == "route-table" {
				continue
			}
			// Default security group is auto-deleted (check if it's the default SG for this VPC)
			if dep.Type == "security-group" {
				var sg SecurityGroup
				if err := s.state.Get(fmt.Sprintf("ec2:security-groups:%s", dep.ID), &sg); err == nil {
					if sg.GroupName != nil && *sg.GroupName == "default" && sg.VpcId != nil && *sg.VpcId == vpcId {
						continue // Skip default security group
					}
				}
			}
			filteredDependents = append(filteredDependents, dep)
		}
		if len(filteredDependents) > 0 {
			return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("The vpc '%s' has dependencies and cannot be deleted: %v", vpcId, filteredDependents)), nil
		}
	}

	// Collect resources to delete (for two-phase delete: graph first, then state)
	var rtbsToDelete []string // route table keys
	var rtbIds []string       // route table IDs for graph
	var sgToDelete string     // default security group key
	var sgId string           // default security group ID for graph

	rtbKeys, _ := s.state.List("ec2:route-tables:")
	for _, rtbKey := range rtbKeys {
		var rtb RouteTable
		if err := s.state.Get(rtbKey, &rtb); err == nil {
			if rtb.VpcId != nil && *rtb.VpcId == vpcId {
				rtbsToDelete = append(rtbsToDelete, rtbKey)
				rtbIds = append(rtbIds, *rtb.RouteTableId)
			}
		}
	}

	sgKeys, _ := s.state.List("ec2:security-groups:")
	for _, sgKey := range sgKeys {
		var sg SecurityGroup
		if err := s.state.Get(sgKey, &sg); err == nil {
			if sg.VpcId != nil && *sg.VpcId == vpcId && sg.GroupName != nil && *sg.GroupName == "default" {
				sgToDelete = sgKey
				sgId = *sg.GroupId
				break
			}
		}
	}

	// Phase 1: Unregister all resources from graph (fail-fast, no state mutations yet)
	// Unregister children first, then parent
	for _, rtbId := range rtbIds {
		s.unregisterResource("route-table", rtbId)
	}
	if sgId != "" {
		s.unregisterResource("security-group", sgId)
	}
	if err := s.unregisterResource("vpc", vpcId); err != nil {
		return s.errorResponse(400, "DependencyViolation", fmt.Sprintf("Cannot delete VPC: %v", err)), nil
	}

	// Phase 2: Delete all resources from state (graph is already consistent)
	for _, rtbKey := range rtbsToDelete {
		s.state.Delete(rtbKey)
	}
	if sgToDelete != "" {
		s.state.Delete(sgToDelete)
	}
	s.state.Delete(fmt.Sprintf("ec2:vpcs:%s", vpcId))

	return s.deleteVpcResponse()
}
