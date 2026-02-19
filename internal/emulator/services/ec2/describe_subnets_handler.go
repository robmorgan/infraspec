package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) describeSubnets(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	subnetIds := s.parseSubnetIds(params)

	var subnets []Subnet

	if len(subnetIds) > 0 {
		for _, subnetId := range subnetIds {
			var subnet Subnet
			if err := s.state.Get(fmt.Sprintf("ec2:subnets:%s", subnetId), &subnet); err != nil {
				return s.errorResponse(400, "InvalidSubnetID.NotFound", fmt.Sprintf("The subnet ID '%s' does not exist", subnetId)), nil
			}
			subnets = append(subnets, subnet)
		}
	} else {
		keys, err := s.state.List("ec2:subnets:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list subnets"), nil
		}

		for _, key := range keys {
			var subnet Subnet
			if err := s.state.Get(key, &subnet); err == nil {
				subnets = append(subnets, subnet)
			}
		}
	}

	return s.describeSubnetsResponse(subnets)
}
