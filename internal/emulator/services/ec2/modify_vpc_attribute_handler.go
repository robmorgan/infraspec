package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) modifyVpcAttribute(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	// Verify VPC exists
	var vpc Vpc
	if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
	}

	// Get or create VPC attributes (stored separately from the VPC resource)
	attrKey := fmt.Sprintf("ec2:vpc-attributes:%s", vpcId)
	var attrs VpcAttributes
	if err := s.state.Get(attrKey, &attrs); err != nil {
		// Initialize defaults - AWS defaults EnableDnsSupport to true
		attrs = VpcAttributes{
			EnableDnsHostnames:               false,
			EnableDnsSupport:                 true,
			EnableNetworkAddressUsageMetrics: false,
		}
	}

	// Handle EnableDnsHostnames attribute
	if enableDnsHostnames, ok := params["EnableDnsHostnames.Value"].(string); ok {
		attrs.EnableDnsHostnames = enableDnsHostnames == "true"
	}

	// Handle EnableDnsSupport attribute
	if enableDnsSupport, ok := params["EnableDnsSupport.Value"].(string); ok {
		attrs.EnableDnsSupport = enableDnsSupport == "true"
	}

	// Handle EnableNetworkAddressUsageMetrics attribute
	if enableNetworkAddrUsage, ok := params["EnableNetworkAddressUsageMetrics.Value"].(string); ok {
		attrs.EnableNetworkAddressUsageMetrics = enableNetworkAddrUsage == "true"
	}

	// Save the updated VPC attributes
	if err := s.state.Set(attrKey, &attrs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update VPC attributes"), nil
	}

	return s.modifyVpcAttributeResponse()
}
