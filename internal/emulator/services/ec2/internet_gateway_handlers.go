package ec2

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createInternetGateway(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	igwId := fmt.Sprintf("igw-%s", uuid.New().String()[:8])

	igw := InternetGateway{
		InternetGatewayId: &igwId,
		OwnerId:           helpers.StringPtr("123456789012"),
		Attachments:       []InternetGatewayAttachment{},
	}

	if err := s.state.Set(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store internet gateway"), nil
	}

	return s.createInternetGatewayResponse(igw)
}

func (s *EC2Service) describeInternetGateways(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	igwIds := s.parseInternetGatewayIds(params)

	var gateways []InternetGateway

	if len(igwIds) > 0 {
		for _, igwId := range igwIds {
			var igw InternetGateway
			if err := s.state.Get(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
				return s.errorResponse(400, "InvalidInternetGatewayID.NotFound", fmt.Sprintf("The internet gateway ID '%s' does not exist", igwId)), nil
			}
			gateways = append(gateways, igw)
		}
	} else {
		keys, err := s.state.List("ec2:internet-gateways:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list internet gateways"), nil
		}

		for _, key := range keys {
			var igw InternetGateway
			if err := s.state.Get(key, &igw); err == nil {
				gateways = append(gateways, igw)
			}
		}
	}

	return s.describeInternetGatewaysResponse(gateways)
}

func (s *EC2Service) attachInternetGateway(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	igwId, ok := params["InternetGatewayId"].(string)
	if !ok || igwId == "" {
		return s.errorResponse(400, "MissingParameter", "InternetGatewayId is required"), nil
	}

	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	// Acquire per-resource lock for atomic operation
	resourceKey := "igw:" + igwId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var igw InternetGateway
	if err := s.state.Get(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(400, "InvalidInternetGatewayID.NotFound", fmt.Sprintf("The internet gateway ID '%s' does not exist", igwId)), nil
	}

	var vpc Vpc
	if err := s.state.Get(fmt.Sprintf("ec2:vpcs:%s", vpcId), &vpc); err != nil {
		return s.errorResponse(400, "InvalidVpcID.NotFound", fmt.Sprintf("The vpc ID '%s' does not exist", vpcId)), nil
	}

	// Check if IGW is already attached to a VPC
	for _, att := range igw.Attachments {
		if att.VpcId != nil && att.State == AttachmentStatus("available") {
			return s.errorResponse(400, "Resource.AlreadyAssociated",
				fmt.Sprintf("Internet gateway '%s' is already attached to a VPC", igwId)), nil
		}
	}

	igw.Attachments = append(igw.Attachments, InternetGatewayAttachment{
		VpcId: &vpcId,
		State: AttachmentStatus("available"),
	})

	if err := s.state.Set(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update internet gateway"), nil
	}

	return s.attachInternetGatewayResponse()
}

func (s *EC2Service) detachInternetGateway(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	igwId, ok := params["InternetGatewayId"].(string)
	if !ok || igwId == "" {
		return s.errorResponse(400, "MissingParameter", "InternetGatewayId is required"), nil
	}

	vpcId, ok := params["VpcId"].(string)
	if !ok || vpcId == "" {
		return s.errorResponse(400, "MissingParameter", "VpcId is required"), nil
	}

	// Acquire per-resource lock for atomic operation
	resourceKey := "igw:" + igwId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var igw InternetGateway
	if err := s.state.Get(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(400, "InvalidInternetGatewayID.NotFound", fmt.Sprintf("The internet gateway ID '%s' does not exist", igwId)), nil
	}

	// Validate IGW is attached to the specified VPC
	found := false
	for _, att := range igw.Attachments {
		if att.VpcId != nil && *att.VpcId == vpcId && att.State == AttachmentStatus("available") {
			found = true
			break
		}
	}
	if !found {
		return s.errorResponse(400, "Gateway.NotAttached",
			fmt.Sprintf("Internet gateway '%s' is not attached to VPC '%s'", igwId, vpcId)), nil
	}

	// Remove the attachment
	newAttachments := make([]InternetGatewayAttachment, 0)
	for _, att := range igw.Attachments {
		if att.VpcId != nil && *att.VpcId != vpcId {
			newAttachments = append(newAttachments, att)
		}
	}
	igw.Attachments = newAttachments

	if err := s.state.Set(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update internet gateway"), nil
	}

	return s.detachInternetGatewayResponse()
}

func (s *EC2Service) deleteInternetGateway(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	igwId, ok := params["InternetGatewayId"].(string)
	if !ok || igwId == "" {
		return s.errorResponse(400, "MissingParameter", "InternetGatewayId is required"), nil
	}

	// Acquire per-resource lock for atomic operation
	resourceKey := "igw:" + igwId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var igw InternetGateway
	if err := s.state.Get(fmt.Sprintf("ec2:internet-gateways:%s", igwId), &igw); err != nil {
		return s.errorResponse(400, "InvalidInternetGatewayID.NotFound", fmt.Sprintf("The internet gateway ID '%s' does not exist", igwId)), nil
	}

	if len(igw.Attachments) > 0 {
		return s.errorResponse(400, "DependencyViolation", "The internet gateway is still attached to a VPC"), nil
	}

	s.state.Delete(fmt.Sprintf("ec2:internet-gateways:%s", igwId))

	// Clean up state machine entry
	s.stateMachine.RemoveResourceState(resourceKey)

	return s.deleteInternetGatewayResponse()
}
