package ec2

import (
	"encoding/xml"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// ==================== Response Builders ====================

// errorResponse builds an EC2 error response
func (s *EC2Service) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildEC2ErrorResponse(statusCode, code, message)
}

// successResponse builds an EC2 response using the standard response builder pattern.
// Note: EC2 uses a different XML format than Query Protocol services (RDS, IAM, STS).
// EC2 responses do NOT have an <ActionResult> wrapper - data is placed directly
// inside the <ActionResponse> element with a <RequestId> element at the end.
func (s *EC2Service) successResponse(action string, data interface{}) (*emulator.AWSResponse, error) {
	// Use EC2-specific response builder (no <ActionResult> wrapper)
	resp, err := emulator.BuildEC2Response(data, emulator.ResponseBuilderConfig{
		ServiceName: "ec2",
		Version:     "2016-11-15",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build response: %w", err)
	}

	// Validate response headers and structure
	if err := s.responseValidator.ValidateResponseHeaders(resp, "ec2"); err != nil {
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	return resp, nil
}

// ==================== Instance Responses ====================

func (s *EC2Service) runInstancesResponse(reservation Reservation) (*emulator.AWSResponse, error) {
	type RunInstancesResult struct {
		XMLName       xml.Name   `xml:"RunInstancesResponse"`
		ReservationId string     `xml:"reservationId"`
		OwnerId       string     `xml:"ownerId"`
		InstancesSet  []Instance `xml:"instancesSet>item"`
	}

	result := RunInstancesResult{
		InstancesSet: reservation.Instances,
	}
	if reservation.ReservationId != nil {
		result.ReservationId = *reservation.ReservationId
	}
	if reservation.OwnerId != nil {
		result.OwnerId = *reservation.OwnerId
	}

	return s.successResponse("RunInstances", result)
}

func (s *EC2Service) describeInstancesResponse(reservations []Reservation) (*emulator.AWSResponse, error) {
	type ReservationItem struct {
		ReservationId string     `xml:"reservationId"`
		OwnerId       string     `xml:"ownerId"`
		InstancesSet  []Instance `xml:"instancesSet>item"`
	}

	type DescribeInstancesResult struct {
		XMLName        xml.Name          `xml:"DescribeInstancesResponse"`
		ReservationSet []ReservationItem `xml:"reservationSet>item"`
	}

	items := make([]ReservationItem, 0)
	for _, r := range reservations {
		item := ReservationItem{
			InstancesSet: r.Instances,
		}
		if r.ReservationId != nil {
			item.ReservationId = *r.ReservationId
		}
		if r.OwnerId != nil {
			item.OwnerId = *r.OwnerId
		}
		items = append(items, item)
	}

	result := DescribeInstancesResult{
		ReservationSet: items,
	}

	return s.successResponse("DescribeInstances", result)
}

func (s *EC2Service) instanceStateChangeResponse(action string, changes []InstanceStateChange) (*emulator.AWSResponse, error) {
	type StateChangeItem struct {
		InstanceId   string `xml:"instanceId"`
		CurrentState struct {
			Code int32  `xml:"code"`
			Name string `xml:"name"`
		} `xml:"currentState"`
		PreviousState struct {
			Code int32  `xml:"code"`
			Name string `xml:"name"`
		} `xml:"previousState"`
	}

	type StateChangeResult struct {
		XMLName      xml.Name          `xml:"instancesSet"`
		InstancesSet []StateChangeItem `xml:"item"`
	}

	items := make([]StateChangeItem, 0)
	for _, c := range changes {
		item := StateChangeItem{}
		if c.InstanceId != nil {
			item.InstanceId = *c.InstanceId
		}
		if c.CurrentState != nil {
			if c.CurrentState.Code != nil {
				item.CurrentState.Code = *c.CurrentState.Code
			}
			item.CurrentState.Name = string(c.CurrentState.Name)
		}
		if c.PreviousState != nil {
			if c.PreviousState.Code != nil {
				item.PreviousState.Code = *c.PreviousState.Code
			}
			item.PreviousState.Name = string(c.PreviousState.Name)
		}
		items = append(items, item)
	}

	result := StateChangeResult{
		InstancesSet: items,
	}

	return s.successResponse(action, result)
}

// ==================== VPC Responses ====================

func (s *EC2Service) createVpcResponse(vpc Vpc) (*emulator.AWSResponse, error) {
	return s.successResponse("CreateVpc", CreateVpcResult{Vpc: &vpc})
}

func (s *EC2Service) describeVpcsResponse(vpcs []Vpc) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeVpcs", VpcSetResponse{VpcSet: vpcs})
}

func (s *EC2Service) deleteVpcResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteVpc", DeleteVpcResponse{Return: true})
}

func (s *EC2Service) modifyVpcAttributeResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("ModifyVpcAttribute", ModifyVpcAttributeResponse{Return: true})
}

// ==================== Subnet Responses ====================

func (s *EC2Service) createSubnetResponse(subnet Subnet) (*emulator.AWSResponse, error) {
	return s.successResponse("CreateSubnet", SubnetResponse{Subnet: subnet})
}

func (s *EC2Service) describeSubnetsResponse(subnets []Subnet) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeSubnets", SubnetSetResponse{SubnetSet: subnets})
}

func (s *EC2Service) deleteSubnetResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteSubnet", DeleteSubnetResponse{Return: true})
}

// ==================== Security Group Responses ====================

func (s *EC2Service) createSecurityGroupResponse(groupId string) (*emulator.AWSResponse, error) {
	return s.successResponse("CreateSecurityGroup", CreateSecurityGroupResponse{GroupId: groupId, Return: true})
}

func (s *EC2Service) describeSecurityGroupsResponse(groups []SecurityGroup) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeSecurityGroups", SecurityGroupInfoResponse{SecurityGroupInfo: groups})
}

func (s *EC2Service) deleteSecurityGroupResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteSecurityGroup", DeleteSecurityGroupResponse{Return: true})
}

func (s *EC2Service) authorizeSecurityGroupIngressResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("AuthorizeSecurityGroupIngress", AuthorizeSecurityGroupIngressResponse{Return: true})
}

func (s *EC2Service) authorizeSecurityGroupEgressResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("AuthorizeSecurityGroupEgress", AuthorizeSecurityGroupEgressResponse{Return: true})
}

func (s *EC2Service) revokeSecurityGroupIngressResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("RevokeSecurityGroupIngress", RevokeSecurityGroupIngressResponse{Return: true})
}

func (s *EC2Service) revokeSecurityGroupEgressResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("RevokeSecurityGroupEgress", RevokeSecurityGroupEgressResponse{Return: true})
}

// ==================== Internet Gateway Responses ====================

func (s *EC2Service) createInternetGatewayResponse(igw InternetGateway) (*emulator.AWSResponse, error) {
	// Use generated CreateInternetGatewayResult which has XMLName for correct root element
	return s.successResponse("CreateInternetGateway", CreateInternetGatewayResult{InternetGateway: &igw})
}

func (s *EC2Service) describeInternetGatewaysResponse(gateways []InternetGateway) (*emulator.AWSResponse, error) {
	// Use generated DescribeInternetGatewaysResult which has XMLName for correct root element
	return s.successResponse("DescribeInternetGateways", DescribeInternetGatewaysResult{InternetGateways: gateways})
}

func (s *EC2Service) attachInternetGatewayResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("AttachInternetGateway", AttachInternetGatewayResponse{Return: true})
}

func (s *EC2Service) detachInternetGatewayResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DetachInternetGateway", DetachInternetGatewayResponse{Return: true})
}

func (s *EC2Service) deleteInternetGatewayResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteInternetGateway", DeleteInternetGatewayResponse{Return: true})
}

// ==================== AMI Responses ====================

func (s *EC2Service) describeImagesResponse(images []Image) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeImages", ImagesSetResponse{ImagesSet: images})
}

// ==================== Volume Responses ====================

func (s *EC2Service) createVolumeResponse(volume Volume) (*emulator.AWSResponse, error) {
	// Pass Volume directly - BuildEC2Response adds the CreateVolumeResponse wrapper
	return s.successResponse("CreateVolume", volume)
}

func (s *EC2Service) describeVolumesResponse(volumes []Volume) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeVolumes", VolumeSetResponse{VolumeSet: volumes})
}

func (s *EC2Service) attachVolumeResponse(attachment VolumeAttachment) (*emulator.AWSResponse, error) {
	// Pass VolumeAttachment directly - BuildEC2Response adds the AttachVolumeResponse wrapper
	return s.successResponse("AttachVolume", attachment)
}

func (s *EC2Service) detachVolumeResponse(attachment VolumeAttachment) (*emulator.AWSResponse, error) {
	// Pass VolumeAttachment directly - BuildEC2Response adds the DetachVolumeResponse wrapper
	return s.successResponse("DetachVolume", attachment)
}

func (s *EC2Service) deleteVolumeResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteVolume", DeleteVolumeResponse{Return: true})
}

// ==================== Key Pair Responses ====================

func (s *EC2Service) createKeyPairResponse(keyPairId, keyName, fingerprint, privateKey string) (*emulator.AWSResponse, error) {
	return s.successResponse("CreateKeyPair", KeyPairResponse{
		KeyPairId:      keyPairId,
		KeyName:        keyName,
		KeyFingerprint: fingerprint,
		KeyMaterial:    privateKey,
	})
}

func (s *EC2Service) describeKeyPairsResponse(keyPairs []KeyPairInfo) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeKeyPairs", KeySetResponse{KeySet: keyPairs})
}

func (s *EC2Service) deleteKeyPairResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteKeyPair", DeleteKeyPairResponse{Return: true})
}

func (s *EC2Service) importKeyPairResponse(keyPairId, keyName, fingerprint string) (*emulator.AWSResponse, error) {
	return s.successResponse("ImportKeyPair", KeyPairImportResponse{
		KeyPairId:      keyPairId,
		KeyName:        keyName,
		KeyFingerprint: fingerprint,
	})
}

// ==================== Launch Template Responses ====================

func (s *EC2Service) createLaunchTemplateResponse(template LaunchTemplate) (*emulator.AWSResponse, error) {
	return s.successResponse("CreateLaunchTemplate", LaunchTemplateResponse{LaunchTemplate: template})
}

func (s *EC2Service) describeLaunchTemplatesResponse(templates []LaunchTemplate) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeLaunchTemplates", LaunchTemplatesResponse{LaunchTemplates: templates})
}

func (s *EC2Service) deleteLaunchTemplateResponse(template LaunchTemplate) (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteLaunchTemplate", LaunchTemplateResponse{LaunchTemplate: template})
}

// ==================== Tag Responses ====================

func (s *EC2Service) createTagsResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("CreateTags", CreateTagsResponse{Return: true})
}

func (s *EC2Service) describeTagsResponse(tags []TagDescription) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeTags", TagSetResponse{TagSet: tags})
}

func (s *EC2Service) deleteTagsResponse() (*emulator.AWSResponse, error) {
	return s.successResponse("DeleteTags", DeleteTagsResponse{Return: true})
}

// ==================== Instance Type Responses ====================

func (s *EC2Service) describeInstanceTypesResponse(instanceTypes []InstanceTypeInfo) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeInstanceTypes", InstanceTypeSetResponse{InstanceTypes: instanceTypes})
}

// ==================== Instance Attribute Responses ====================

func (s *EC2Service) describeInstanceAttributeResponse(attr InstanceAttribute) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeInstanceAttribute", attr)
}

// ==================== Instance Credit Specifications Responses ====================

func (s *EC2Service) describeInstanceCreditSpecificationsResponse(creditSpecs []InstanceCreditSpecification) (*emulator.AWSResponse, error) {
	return s.successResponse("DescribeInstanceCreditSpecifications", DescribeInstanceCreditSpecificationsResult{
		InstanceCreditSpecifications: creditSpecs,
	})
}
