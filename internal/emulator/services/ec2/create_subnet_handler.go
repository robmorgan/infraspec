package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createSubnet(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	cidrBlock, ok := params["CidrBlock"].(string)
	if !ok || cidrBlock == "" {
		return s.errorResponse(400, "MissingParameter", "CidrBlock is required"), nil
	}

	// Validate CIDR block format
	if !isValidCIDR(cidrBlock) {
		return s.errorResponse(400, "InvalidParameterValue",
			fmt.Sprintf("Value (%s) for parameter cidrBlock is invalid. This is not a valid CIDR block.", cidrBlock)), nil
	}

	// Verify VPC exists
	var vpc Vpc
	if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
	}

	subnetId := fmt.Sprintf("subnet-%s", uuid.New().String()[:8])
	az := getStringParamValue(params, "AvailabilityZone", "us-east-1a")

	subnet := Subnet{
		SubnetId:                &subnetId,
		VpcId:                   &vpcId,
		CidrBlock:               &cidrBlock,
		AvailabilityZone:        &az,
		AvailabilityZoneId:      helpers.StringPtr("use1-az1"),
		State:                   SubnetState("pending"),
		DefaultForAz:            helpers.BoolPtr(false),
		MapPublicIpOnLaunch:     helpers.BoolPtr(false),
		AvailableIpAddressCount: helpers.Int32Ptr(251),
		OwnerId:                 helpers.StringPtr("123456789012"),
	}

	stateKey := fmt.Sprintf("ec2:subnets:%s", subnetId)
	if err := s.state.Set(stateKey, &subnet); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store subnet"), nil
	}

	// Register subnet in the relationship graph
	s.registerResource("subnet", subnetId, map[string]string{
		"cidr":  cidrBlock,
		"vpcId": vpcId,
	})

	// Add relationship: subnet -> VPC (subnet is contained in VPC)
	if err := s.addRelationship("subnet", subnetId, "ec2", "vpc", vpcId, graph.RelContains); err != nil {
		if s.isStrictMode() {
			// Rollback: remove from state and graph
			s.state.Delete(stateKey)
			s.unregisterResource("subnet", subnetId)
			return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create subnet-vpc relationship: %v", err)), nil
		}
		log.Printf("Warning: failed to add subnet-vpc relationship in graph: %v", err)
	}

	// Schedule transition to available
	s.scheduleSubnetTransition(subnetId, SubnetState("available"), 2*time.Second)

	return s.createSubnetResponse(subnet)
}
