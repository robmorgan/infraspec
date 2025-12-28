package ec2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// EC2Service implements the EC2 service emulator
type EC2Service struct {
	state             emulator.StateManager
	validator         emulator.Validator
	stateMachine      *ResourceStateManager
	shutdownCtx       context.Context
	shutdownCancel    context.CancelFunc
	responseValidator *emulator.ResponseValidator
	resourceManager   *graph.ResourceManager
}

// NewEC2Service creates a new EC2 service instance
func NewEC2Service(state emulator.StateManager, validator emulator.Validator) *EC2Service {
	ctx, cancel := context.WithCancel(context.Background())

	// Create response registry and validator for EC2
	responseRegistry := emulator.NewResponseRegistry()
	responseValidator := emulator.NewResponseValidator(responseRegistry)

	svc := &EC2Service{
		state:             state,
		validator:         validator,
		stateMachine:      NewResourceStateManager(),
		shutdownCtx:       ctx,
		shutdownCancel:    cancel,
		responseValidator: responseValidator,
	}
	svc.initializeDefaults()
	return svc
}

// NewEC2ServiceWithGraph creates a new EC2 service instance with ResourceManager for relationship tracking
func NewEC2ServiceWithGraph(state emulator.StateManager, validator emulator.Validator, rm *graph.ResourceManager) *EC2Service {
	ctx, cancel := context.WithCancel(context.Background())

	// Create response registry and validator for EC2
	responseRegistry := emulator.NewResponseRegistry()
	responseValidator := emulator.NewResponseValidator(responseRegistry)

	svc := &EC2Service{
		state:             state,
		validator:         validator,
		stateMachine:      NewResourceStateManager(),
		shutdownCtx:       ctx,
		shutdownCancel:    cancel,
		responseValidator: responseValidator,
		resourceManager:   rm,
	}
	svc.initializeDefaults()
	return svc
}

// Shutdown gracefully stops the EC2 service, cancelling all pending transitions
func (s *EC2Service) Shutdown() {
	s.shutdownCancel()
}

// ServiceName returns the service name
func (s *EC2Service) ServiceName() string {
	return "ec2"
}

// SupportedActions returns the list of AWS API actions this service handles.
// Used by the router to determine which service handles a given Query Protocol request.
func (s *EC2Service) SupportedActions() []string {
	return []string{
		// Instance operations
		"RunInstances",
		"DescribeInstances",
		"DescribeInstanceTypes",
		"DescribeInstanceAttribute",
		"DescribeInstanceCreditSpecifications",
		"TerminateInstances",
		"StartInstances",
		"StopInstances",
		// VPC operations
		"CreateVpc",
		"DescribeVpcs",
		"DeleteVpc",
		"ModifyVpcAttribute",
		"DescribeVpcAttribute",
		// Subnet operations
		"CreateSubnet",
		"DescribeSubnets",
		"DeleteSubnet",
		// Security Group operations
		"CreateSecurityGroup",
		"DescribeSecurityGroups",
		"DeleteSecurityGroup",
		"AuthorizeSecurityGroupIngress",
		"AuthorizeSecurityGroupEgress",
		"RevokeSecurityGroupIngress",
		"RevokeSecurityGroupEgress",
		// Internet Gateway operations
		"CreateInternetGateway",
		"DescribeInternetGateways",
		"AttachInternetGateway",
		"DetachInternetGateway",
		"DeleteInternetGateway",
		// AMI operations
		"DescribeImages",
		// Volume operations
		"CreateVolume",
		"DescribeVolumes",
		"AttachVolume",
		"DetachVolume",
		"DeleteVolume",
		// Key Pair operations
		"CreateKeyPair",
		"DescribeKeyPairs",
		"DeleteKeyPair",
		"ImportKeyPair",
		// Launch Template operations
		"CreateLaunchTemplate",
		"DescribeLaunchTemplates",
		"DeleteLaunchTemplate",
		// Tag operations
		"CreateTags",
		"DescribeTags",
		"DeleteTags",
		// Network Interface operations
		"DescribeNetworkInterfaces",
		// Network ACL operations
		"DescribeNetworkAcls",
		// Route Table operations
		"DescribeRouteTables",
	}
}

// initializeDefaults sets up default VPC, subnet, security group, and AMIs
func (s *EC2Service) initializeDefaults() {
	// Create default VPC
	defaultVpcId := "vpc-default"
	defaultVpc := Vpc{
		VpcId:           &defaultVpcId,
		CidrBlock:       helpers.StringPtr("172.31.0.0/16"),
		State:           VpcState("available"),
		IsDefault:       helpers.BoolPtr(true),
		OwnerId:         helpers.StringPtr("123456789012"),
		InstanceTenancy: Tenancy("default"),
	}
	s.state.Set(fmt.Sprintf("ec2:vpcs:%s", defaultVpcId), &defaultVpc)

	// Register default VPC in graph
	s.registerResource("vpc", defaultVpcId, map[string]string{
		"cidr":    "172.31.0.0/16",
		"default": "true",
	})

	// Create default subnet
	defaultSubnetId := "subnet-default"
	defaultSubnet := Subnet{
		SubnetId:                &defaultSubnetId,
		VpcId:                   &defaultVpcId,
		CidrBlock:               helpers.StringPtr("172.31.0.0/20"),
		AvailabilityZone:        helpers.StringPtr("us-east-1a"),
		AvailabilityZoneId:      helpers.StringPtr("use1-az1"),
		State:                   SubnetState("available"),
		DefaultForAz:            helpers.BoolPtr(true),
		MapPublicIpOnLaunch:     helpers.BoolPtr(true),
		AvailableIpAddressCount: helpers.Int32Ptr(4091),
		OwnerId:                 helpers.StringPtr("123456789012"),
	}
	s.state.Set(fmt.Sprintf("ec2:subnets:%s", defaultSubnetId), &defaultSubnet)

	// Register default subnet in graph and add relationship to VPC
	s.registerResource("subnet", defaultSubnetId, map[string]string{
		"cidr":    "172.31.0.0/20",
		"vpcId":   defaultVpcId,
		"default": "true",
	})
	if err := s.addRelationship("subnet", defaultSubnetId, "ec2", "vpc", defaultVpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add default subnet-vpc relationship in graph: %v", err)
	}

	// Create default security group
	defaultSgId := "sg-default"
	defaultSgName := "default"
	defaultSg := SecurityGroup{
		GroupId:     &defaultSgId,
		GroupName:   &defaultSgName,
		Description: helpers.StringPtr("default VPC security group"),
		VpcId:       &defaultVpcId,
		OwnerId:     helpers.StringPtr("123456789012"),
		IpPermissions: []IpPermission{
			{
				IpProtocol: helpers.StringPtr("-1"),
				UserIdGroupPairs: []UserIdGroupPair{
					{GroupId: &defaultSgId},
				},
			},
		},
		IpPermissionsEgress: []IpPermission{
			{
				IpProtocol: helpers.StringPtr("-1"),
				IpRanges:   []IpRange{{CidrIp: helpers.StringPtr("0.0.0.0/0")}},
			},
		},
	}
	s.state.Set(fmt.Sprintf("ec2:security-groups:%s", defaultSgId), &defaultSg)

	// Register default security group in graph and add relationship to VPC
	s.registerResource("security-group", defaultSgId, map[string]string{
		"name":    defaultSgName,
		"vpcId":   defaultVpcId,
		"default": "true",
	})
	if err := s.addRelationship("security-group", defaultSgId, "ec2", "vpc", defaultVpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add default security-group-vpc relationship in graph: %v", err)
	}

	// Create default network ACL for default VPC
	defaultNaclId := "acl-default"
	defaultNacl := NetworkAcl{
		NetworkAclId: helpers.StringPtr(defaultNaclId),
		VpcId:        &defaultVpcId,
		IsDefault:    helpers.BoolPtr(true),
		OwnerId:      helpers.StringPtr("123456789012"),
		Entries: []NetworkAclEntry{
			{
				RuleNumber: helpers.Int32Ptr(100),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("allow"),
				Egress:     helpers.BoolPtr(false),
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
			},
			{
				RuleNumber: helpers.Int32Ptr(32767),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("deny"),
				Egress:     helpers.BoolPtr(false),
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
			},
			{
				RuleNumber: helpers.Int32Ptr(100),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("allow"),
				Egress:     helpers.BoolPtr(true),
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
			},
			{
				RuleNumber: helpers.Int32Ptr(32767),
				Protocol:   helpers.StringPtr("-1"),
				RuleAction: RuleAction("deny"),
				Egress:     helpers.BoolPtr(true),
				CidrBlock:  helpers.StringPtr("0.0.0.0/0"),
			},
		},
		Associations: []NetworkAclAssociation{},
		Tags:         []Tag{},
	}
	s.state.Set(fmt.Sprintf("ec2:network-acls:%s", defaultNaclId), &defaultNacl)

	// Register default network ACL in graph and add relationship to VPC
	s.registerResource("network-acl", defaultNaclId, map[string]string{
		"vpcId":   defaultVpcId,
		"default": "true",
	})
	if err := s.addRelationship("network-acl", defaultNaclId, "ec2", "vpc", defaultVpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add default network-acl-vpc relationship in graph: %v", err)
	}

	// Create default route table for default VPC
	defaultRtbId := "rtb-default"
	defaultRtb := RouteTable{
		RouteTableId: helpers.StringPtr(defaultRtbId),
		VpcId:        &defaultVpcId,
		OwnerId:      helpers.StringPtr("123456789012"),
		Routes: []Route{
			{
				DestinationCidrBlock: helpers.StringPtr("172.31.0.0/16"),
				GatewayId:            helpers.StringPtr("local"),
				State:                RouteState("active"),
				Origin:               RouteOrigin("CreateRouteTable"),
			},
		},
		Associations: []RouteTableAssociation{
			{
				RouteTableAssociationId: helpers.StringPtr("rtbassoc-default"),
				RouteTableId:            helpers.StringPtr(defaultRtbId),
				Main:                    helpers.BoolPtr(true),
				AssociationState: &RouteTableAssociationState{
					State: RouteTableAssociationStateCode("associated"),
				},
			},
		},
		Tags: []Tag{},
	}
	s.state.Set(fmt.Sprintf("ec2:route-tables:%s", defaultRtbId), &defaultRtb)

	// Register default route table in graph and add relationship to VPC
	s.registerResource("route-table", defaultRtbId, map[string]string{
		"vpcId":   defaultVpcId,
		"main":    "true",
		"default": "true",
	})
	if err := s.addRelationship("route-table", defaultRtbId, "ec2", "vpc", defaultVpcId, graph.RelContains); err != nil {
		log.Printf("Warning: failed to add default route-table-vpc relationship in graph: %v", err)
	}

	// Pre-populate common AMIs
	amis := []struct {
		id          string
		name        string
		description string
		arch        ArchitectureValues
	}{
		{"ami-0c55b159cbfafe1f0", "amzn2-ami-hvm-2.0.20210721.2-x86_64-gp2", "Amazon Linux 2 AMI", ArchitectureValues("x86_64")},
		{"ami-0c94855ba95c71c99", "amzn2-ami-hvm-2.0.20210701.0-x86_64-gp2", "Amazon Linux 2 AMI", ArchitectureValues("x86_64")},
		{"ami-0885b1f6bd170450c", "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server", "Ubuntu 20.04 LTS", ArchitectureValues("x86_64")},
		{"ami-0dba2cb6798deb6d8", "ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server", "Ubuntu 18.04 LTS", ArchitectureValues("x86_64")},
	}

	for _, ami := range amis {
		image := Image{
			ImageId:            helpers.StringPtr(ami.id),
			Name:               helpers.StringPtr(ami.name),
			Description:        helpers.StringPtr(ami.description),
			Architecture:       ami.arch,
			ImageType:          ImageTypeValues("machine"),
			State:              ImageState("available"),
			RootDeviceType:     DeviceType("ebs"),
			RootDeviceName:     helpers.StringPtr("/dev/xvda"),
			OwnerId:            helpers.StringPtr("amazon"),
			Public:             helpers.BoolPtr(true),
			VirtualizationType: VirtualizationType("hvm"),
		}
		s.state.Set(fmt.Sprintf("ec2:images:%s", ami.id), &image)
	}
}

// HandleRequest handles EC2 API requests
func (s *EC2Service) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	params, err := s.parseParameters(req)
	if err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	if err := s.validator.ValidateAction(action, params); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	switch action {
	// Instance operations
	case "RunInstances":
		return s.runInstances(ctx, params)
	case "DescribeInstances":
		return s.describeInstances(ctx, params)
	case "DescribeInstanceTypes":
		return s.describeInstanceTypes(ctx, params)
	case "DescribeInstanceAttribute":
		return s.describeInstanceAttribute(ctx, params)
	case "DescribeInstanceCreditSpecifications":
		return s.describeInstanceCreditSpecifications(ctx, params)
	case "TerminateInstances":
		return s.terminateInstances(ctx, params)
	case "StartInstances":
		return s.startInstances(ctx, params)
	case "StopInstances":
		return s.stopInstances(ctx, params)

	// VPC operations
	case "CreateVpc":
		return s.createVpc(ctx, params)
	case "DescribeVpcs":
		return s.describeVpcs(ctx, params)
	case "DeleteVpc":
		return s.deleteVpc(ctx, params)
	case "ModifyVpcAttribute":
		return s.modifyVpcAttribute(ctx, params)
	case "DescribeVpcAttribute":
		return s.describeVpcAttribute(ctx, params)

	// Subnet operations
	case "CreateSubnet":
		return s.createSubnet(ctx, params)
	case "DescribeSubnets":
		return s.describeSubnets(ctx, params)
	case "DeleteSubnet":
		return s.deleteSubnet(ctx, params)

	// Security Group operations
	case "CreateSecurityGroup":
		return s.createSecurityGroup(ctx, params)
	case "DescribeSecurityGroups":
		return s.describeSecurityGroups(ctx, params)
	case "DeleteSecurityGroup":
		return s.deleteSecurityGroup(ctx, params)
	case "AuthorizeSecurityGroupIngress":
		return s.authorizeSecurityGroupIngress(ctx, params)
	case "AuthorizeSecurityGroupEgress":
		return s.authorizeSecurityGroupEgress(ctx, params)
	case "RevokeSecurityGroupIngress":
		return s.revokeSecurityGroupIngress(ctx, params)
	case "RevokeSecurityGroupEgress":
		return s.revokeSecurityGroupEgress(ctx, params)

	// Internet Gateway operations
	case "CreateInternetGateway":
		return s.createInternetGateway(ctx, params)
	case "DescribeInternetGateways":
		return s.describeInternetGateways(ctx, params)
	case "AttachInternetGateway":
		return s.attachInternetGateway(ctx, params)
	case "DetachInternetGateway":
		return s.detachInternetGateway(ctx, params)
	case "DeleteInternetGateway":
		return s.deleteInternetGateway(ctx, params)

	// AMI operations
	case "DescribeImages":
		return s.describeImages(ctx, params)

	// Volume operations
	case "CreateVolume":
		return s.createVolume(ctx, params)
	case "DescribeVolumes":
		return s.describeVolumes(ctx, params)
	case "AttachVolume":
		return s.attachVolume(ctx, params)
	case "DetachVolume":
		return s.detachVolume(ctx, params)
	case "DeleteVolume":
		return s.deleteVolume(ctx, params)

	// Key Pair operations
	case "CreateKeyPair":
		return s.createKeyPair(ctx, params)
	case "DescribeKeyPairs":
		return s.describeKeyPairs(ctx, params)
	case "DeleteKeyPair":
		return s.deleteKeyPair(ctx, params)
	case "ImportKeyPair":
		return s.importKeyPair(ctx, params)

	// Launch Template operations
	case "CreateLaunchTemplate":
		return s.createLaunchTemplate(ctx, params)
	case "DescribeLaunchTemplates":
		return s.describeLaunchTemplates(ctx, params)
	case "DeleteLaunchTemplate":
		return s.deleteLaunchTemplate(ctx, params)

	// Tag operations
	case "CreateTags":
		return s.createTags(ctx, params)
	case "DescribeTags":
		return s.describeTags(ctx, params)
	case "DeleteTags":
		return s.deleteTags(ctx, params)

	// Network Interface operations
	case "DescribeNetworkInterfaces":
		return s.describeNetworkInterfaces(ctx, params)

	// Network ACL operations
	case "DescribeNetworkAcls":
		return s.describeNetworkAcls(ctx, params)

	// Route Table operations
	case "DescribeRouteTables":
		return s.describeRouteTables(ctx, params)

	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *EC2Service) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *EC2Service) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
	if req.Parameters != nil {
		return req.Parameters, nil
	}

	contentType := req.Headers["Content-Type"]
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return s.parseFormData(string(req.Body))
	}

	if strings.Contains(contentType, "application/json") {
		var params map[string]interface{}
		if err := json.Unmarshal(req.Body, &params); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %w", err)
		}
		return params, nil
	}

	return make(map[string]interface{}), nil
}

func (s *EC2Service) parseFormData(body string) (map[string]interface{}, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			params[key] = vals[0]
		} else {
			params[key] = vals
		}
	}

	return params, nil
}
