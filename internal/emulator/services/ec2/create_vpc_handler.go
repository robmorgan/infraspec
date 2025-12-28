package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createVpc(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	cidrBlock, ok := params["CidrBlock"].(string)
	if !ok || cidrBlock == "" {
		return s.errorResponse(400, "MissingParameter", "CidrBlock is required"), nil
	}

	// Validate CIDR block format
	if !isValidCIDR(cidrBlock) {
		return s.errorResponse(400, "InvalidParameterValue",
			fmt.Sprintf("Value (%s) for parameter cidrBlock is invalid. This is not a valid CIDR block.", cidrBlock)), nil
	}

	vpcId := fmt.Sprintf("vpc-%s", uuid.New().String()[:8])

	// Parse tags from TagSpecification parameters
	tags := s.parseTagSpecifications(params, "vpc")

	vpc := Vpc{
		VpcId:           &vpcId,
		CidrBlock:       &cidrBlock,
		State:           VpcState("pending"),
		IsDefault:       helpers.BoolPtr(false),
		OwnerId:         helpers.StringPtr("123456789012"),
		InstanceTenancy: Tenancy("default"),
		Tags:            tags,
	}

	if tenancy, ok := params["InstanceTenancy"].(string); ok {
		vpc.InstanceTenancy = Tenancy(tenancy)
	}

	if err := s.state.Set(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store VPC"), nil
	}

	// Also store tags in the separate tag storage for consistency with CreateTags
	if len(tags) > 0 {
		s.state.Set(fmt.Sprintf("ec2:tags:%s", vpcId), tags)
	}

	// Register VPC in the relationship graph
	s.registerResource("vpc", vpcId, map[string]string{
		"cidr": cidrBlock,
	})

	// Create the main route table for this VPC
	rtbId := fmt.Sprintf("rtb-%s", strings.TrimPrefix(vpcId, "vpc-"))
	rtb := RouteTable{
		RouteTableId: helpers.StringPtr(rtbId),
		VpcId:        &vpcId,
		OwnerId:      helpers.StringPtr("123456789012"),
		Routes: []Route{
			{
				DestinationCidrBlock: &cidrBlock,
				GatewayId:            helpers.StringPtr("local"),
				State:                RouteState("active"),
				Origin:               RouteOrigin("CreateRouteTable"),
			},
		},
		Associations: []RouteTableAssociation{
			{
				RouteTableAssociationId: helpers.StringPtr(fmt.Sprintf("rtbassoc-%s", strings.TrimPrefix(vpcId, "vpc-"))),
				RouteTableId:            helpers.StringPtr(rtbId),
				Main:                    helpers.BoolPtr(true),
				AssociationState: &RouteTableAssociationState{
					State: RouteTableAssociationStateCode("associated"),
				},
			},
		},
		Tags: []Tag{},
	}
	s.state.Set(fmt.Sprintf("ec2:route-tables:%s", rtbId), &rtb)

	// Register route table in graph with VPC relationship
	s.registerResource("route-table", rtbId, map[string]string{
		"vpcId": vpcId,
		"main":  "true",
	})
	if err := s.addRelationship("route-table", rtbId, "ec2", "vpc", vpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add route-table-vpc relationship in graph: %v", err)
	}

	// Create the default security group for this VPC (AWS creates one automatically)
	sgId := fmt.Sprintf("sg-%s", strings.TrimPrefix(vpcId, "vpc-"))
	defaultSgName := "default"
	defaultSgDesc := "default VPC security group"
	defaultSG := SecurityGroup{
		GroupId:     helpers.StringPtr(sgId),
		GroupName:   &defaultSgName,
		Description: &defaultSgDesc,
		VpcId:       &vpcId,
		OwnerId:     helpers.StringPtr("123456789012"),
		// Default inbound rule: allow all traffic from resources in this security group
		IpPermissions: []IpPermission{
			{
				IpProtocol: helpers.StringPtr("-1"),
				UserIdGroupPairs: []UserIdGroupPair{
					{
						GroupId: helpers.StringPtr(sgId),
						UserId:  helpers.StringPtr("123456789012"),
					},
				},
			},
		},
		// Default outbound rule: allow all traffic to anywhere
		IpPermissionsEgress: []IpPermission{
			{
				IpProtocol: helpers.StringPtr("-1"),
				IpRanges:   []IpRange{{CidrIp: helpers.StringPtr("0.0.0.0/0")}},
			},
		},
		Tags: []Tag{},
	}
	s.state.Set(fmt.Sprintf("ec2:security-groups:%s", sgId), &defaultSG)

	// Register default security group in graph with VPC relationship
	s.registerResource("security-group", sgId, map[string]string{
		"name":  defaultSgName,
		"vpcId": vpcId,
	})
	if err := s.addRelationship("security-group", sgId, "ec2", "vpc", vpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add security-group-vpc relationship in graph: %v", err)
	}

	// Schedule transition to available
	s.scheduleVpcTransition(vpcId, VpcState("available"), 2*time.Second)

	return s.createVpcResponse(vpc)
}
