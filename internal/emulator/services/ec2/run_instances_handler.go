package ec2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) runInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	imageId, ok := params["ImageId"].(string)
	if !ok || imageId == "" {
		return s.errorResponse(400, "MissingParameter", "ImageId is required"), nil
	}

	// Get instance count
	minCount := getIntParam(params, "MinCount", 1)
	maxCount := getIntParam(params, "MaxCount", minCount)
	if maxCount < minCount {
		maxCount = minCount
	}

	instanceType := getStringParamValue(params, "InstanceType", "t2.micro")
	keyName := getStringParamValue(params, "KeyName", "")

	// Get subnet ID - check if explicitly provided
	subnetId := getStringParamValue(params, "SubnetId", "")
	var subnet Subnet
	vpcId := "vpc-default"

	if subnetId != "" {
		// Subnet was explicitly specified - validate it exists
		if err := s.state.Get(fmt.Sprintf("ec2:subnets:%s", subnetId), &subnet); err != nil {
			return s.errorResponse(400, "InvalidSubnetID.NotFound",
				fmt.Sprintf("The subnet ID '%s' does not exist", subnetId)), nil
		}
		if subnet.VpcId != nil {
			vpcId = *subnet.VpcId
		}
	} else {
		subnetId = "subnet-default"
	}

	// Parse security group IDs
	securityGroupIds := s.parseSecurityGroupIds(params)
	if len(securityGroupIds) == 0 {
		securityGroupIds = []string{"sg-default"}
	}

	// Parse tags from TagSpecification parameters for instances
	instanceTags := s.parseTagSpecifications(params, "instance")

	// Create reservation
	reservationId := fmt.Sprintf("r-%s", uuid.New().String()[:8])
	instances := make([]Instance, 0, maxCount)

	for i := 0; i < maxCount; i++ {
		instanceId := fmt.Sprintf("i-%s", uuid.New().String()[:17])
		privateIp := fmt.Sprintf("172.31.%d.%d", (i/256)%256, i%256+10)

		instance := Instance{
			InstanceId:       &instanceId,
			ImageId:          &imageId,
			InstanceType:     InstanceType(instanceType),
			PrivateIpAddress: &privateIp,
			PrivateDnsName:   helpers.StringPtr(fmt.Sprintf("ip-%s.ec2.internal", strings.ReplaceAll(privateIp, ".", "-"))),
			SubnetId:         &subnetId,
			VpcId:            &vpcId,
			State: &InstanceState{
				Code: helpers.Int32Ptr(0),
				Name: InstanceStateName("pending"),
			},
			LaunchTime:         helpers.TimePtr(time.Now()),
			Architecture:       ArchitectureValues("x86_64"),
			RootDeviceType:     DeviceType("ebs"),
			RootDeviceName:     helpers.StringPtr("/dev/xvda"),
			Hypervisor:         HypervisorType("xen"),
			VirtualizationType: VirtualizationType("hvm"),
			Tags:               instanceTags,
		}

		if keyName != "" {
			instance.KeyName = &keyName
		}

		// Add security groups
		for _, sgId := range securityGroupIds {
			var sg SecurityGroup
			if err := s.state.Get(fmt.Sprintf("ec2:security-groups:%s", sgId), &sg); err == nil {
				instance.SecurityGroups = append(instance.SecurityGroups, GroupIdentifier{
					GroupId:   sg.GroupId,
					GroupName: sg.GroupName,
				})
			}
		}

		// Store instance
		if err := s.state.Set(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to store instance"), nil
		}

		// Also store tags in the separate tag storage for consistency with CreateTags
		if len(instanceTags) > 0 {
			s.state.Set(fmt.Sprintf("ec2:tags:%s", instanceId), instanceTags)
		}

		instances = append(instances, instance)

		// Schedule transition to running after delay
		s.scheduleInstanceTransition(instanceId, InstanceStateName("running"), 5*time.Second)
	}

	// Store reservation
	reservation := Reservation{
		ReservationId: &reservationId,
		OwnerId:       helpers.StringPtr("123456789012"),
		Instances:     instances,
	}
	s.state.Set(fmt.Sprintf("ec2:reservations:%s", reservationId), &reservation)

	return s.runInstancesResponse(reservation)
}
