package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createSecurityGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName, ok := params["GroupName"].(string)
	if !ok || groupName == "" {
		return s.errorResponse(400, "MissingParameter", "GroupName is required"), nil
	}

	description, ok := params["GroupDescription"].(string)
	if !ok || description == "" {
		return s.errorResponse(400, "MissingParameter", "GroupDescription is required"), nil
	}

	vpcId := getStringParamValue(params, "VpcId", "vpc-default")

	groupId := fmt.Sprintf("sg-%s", uuid.New().String()[:8])

	sg := SecurityGroup{
		GroupId:       &groupId,
		GroupName:     &groupName,
		Description:   &description,
		VpcId:         &vpcId,
		OwnerId:       helpers.StringPtr("123456789012"),
		IpPermissions: []IpPermission{},
		IpPermissionsEgress: []IpPermission{
			{
				IpProtocol: helpers.StringPtr("-1"),
				IpRanges:   []IpRange{{CidrIp: helpers.StringPtr("0.0.0.0/0")}},
			},
		},
	}

	stateKey := fmt.Sprintf("ec2:security-groups:%s", groupId)
	if err := s.state.Set(stateKey, &sg); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store security group"), nil
	}

	// Register security group in the relationship graph
	s.registerResource("security-group", groupId, map[string]string{
		"name":  groupName,
		"vpcId": vpcId,
	})

	// Add relationship: security-group -> VPC (security group is contained in VPC)
	if err := s.addRelationship("security-group", groupId, "ec2", "vpc", vpcId, graph.RelContains); err != nil {
		if s.isStrictMode() {
			// Rollback: remove from state and graph
			s.state.Delete(stateKey)
			s.unregisterResource("security-group", groupId)
			return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create security-group-vpc relationship: %v", err)), nil
		}
		log.Printf("Warning: failed to add security-group-vpc relationship in graph: %v", err)
	}

	return s.createSecurityGroupResponse(groupId)
}
