package ec2

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// describeNetworkInterfaces returns network interfaces matching the given filters.
// For the emulator, we return an empty list since we don't fully simulate ENIs.
func (s *EC2Service) describeNetworkInterfaces(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Return an empty list - we don't simulate network interfaces in detail
	// This is sufficient for Terraform cleanup operations that check for ENIs
	return s.successResponse("DescribeNetworkInterfaces", NetworkInterfaceSetResponse{
		NetworkInterfaces: []NetworkInterface{},
	})
}

func (s *EC2Service) describeNetworkAcls(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Extract VPC filter if present (Terraform uses this to find default NACL)
	vpcFilter := s.extractFilterValue(params, "vpc-id")
	defaultFilter := s.extractFilterValue(params, "default")

	var networkAcls []NetworkAcl

	// List all VPCs
	vpcKeys, err := s.state.List("ec2:vpcs:")
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to list VPCs"), nil
	}

	for _, vpcKey := range vpcKeys {
		var vpc Vpc
		if err := s.state.Get(vpcKey, &vpc); err != nil {
			continue
		}

		// Apply VPC filter if present
		if vpcFilter != "" && (vpc.VpcId == nil || *vpc.VpcId != vpcFilter) {
			continue
		}

		// Create a default network ACL for this VPC
		naclId := fmt.Sprintf("acl-%s", strings.TrimPrefix(*vpc.VpcId, "vpc-"))
		isDefault := true

		// Apply default filter if present
		if defaultFilter != "" {
			if defaultFilter == "true" && !isDefault {
				continue
			}
			if defaultFilter == "false" && isDefault {
				continue
			}
		}

		// Create default ACL entries (allow all inbound and outbound)
		entries := []NetworkAclEntry{
			{
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
				Egress:     helpers.BoolPtr(false),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("allow"),
				RuleNumber: helpers.Int32Ptr(100),
			},
			{
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
				Egress:     helpers.BoolPtr(false),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("deny"),
				RuleNumber: helpers.Int32Ptr(32767),
			},
			{
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
				Egress:     helpers.BoolPtr(true),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("allow"),
				RuleNumber: helpers.Int32Ptr(100),
			},
			{
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
				Egress:     helpers.BoolPtr(true),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("deny"),
				RuleNumber: helpers.Int32Ptr(32767),
			},
		}

		nacl := NetworkAcl{
			NetworkAclId: helpers.StringPtr(naclId),
			VpcId:        vpc.VpcId,
			IsDefault:    helpers.BoolPtr(isDefault),
			OwnerId:      helpers.StringPtr("123456789012"),
			Entries:      entries,
			Associations: []NetworkAclAssociation{},
			Tags:         []Tag{},
		}

		networkAcls = append(networkAcls, nacl)
	}

	return s.successResponse("DescribeNetworkAcls", NetworkAclSetResponse{NetworkAcls: networkAcls})
}

func (s *EC2Service) describeRouteTables(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Extract filters
	vpcFilter := s.extractFilterValue(params, "vpc-id")
	mainFilter := s.extractFilterValue(params, "association.main")

	var routeTables []RouteTable

	// List all route tables from state
	rtbKeys, err := s.state.List("ec2:route-tables:")
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to list route tables"), nil
	}

	for _, rtbKey := range rtbKeys {
		var rtb RouteTable
		if err := s.state.Get(rtbKey, &rtb); err != nil {
			continue
		}

		// Apply VPC filter if present
		if vpcFilter != "" && (rtb.VpcId == nil || *rtb.VpcId != vpcFilter) {
			continue
		}

		// Apply main filter if present
		if mainFilter != "" {
			isMain := false
			for _, assoc := range rtb.Associations {
				if assoc.Main != nil && *assoc.Main {
					isMain = true
					break
				}
			}
			if mainFilter == "true" && !isMain {
				continue
			}
			if mainFilter == "false" && isMain {
				continue
			}
		}

		routeTables = append(routeTables, rtb)
	}

	return s.successResponse("DescribeRouteTables", RouteTableSetResponse{RouteTables: routeTables})
}
