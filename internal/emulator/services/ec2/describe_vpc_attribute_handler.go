package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) describeVpcAttribute(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	attribute, ok := params["Attribute"].(string)
	if !ok || attribute == "" {
		return s.errorResponse(400, "MissingParameter", "Attribute is required"), nil
	}

	// Verify VPC exists
	var vpc Vpc
	if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
	}

	// Get VPC attributes (or use defaults)
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

	// Return the requested attribute
	switch attribute {
	case "enableDnsHostnames":
		return s.successResponse("DescribeVpcAttribute", DescribeVpcAttributeResponse{
			VpcId:              vpcId,
			EnableDnsHostnames: &VpcAttributeBooleanValue{Value: attrs.EnableDnsHostnames},
		})
	case "enableDnsSupport":
		return s.successResponse("DescribeVpcAttribute", DescribeVpcAttributeResponse{
			VpcId:            vpcId,
			EnableDnsSupport: &VpcAttributeBooleanValue{Value: attrs.EnableDnsSupport},
		})
	case "enableNetworkAddressUsageMetrics":
		return s.successResponse("DescribeVpcAttribute", DescribeVpcAttributeResponse{
			VpcId:                            vpcId,
			EnableNetworkAddressUsageMetrics: &VpcAttributeBooleanValue{Value: attrs.EnableNetworkAddressUsageMetrics},
		})
	default:
		return s.errorResponse(400, "InvalidParameterValue", fmt.Sprintf("Invalid attribute '%s'", attribute)), nil
	}
}
