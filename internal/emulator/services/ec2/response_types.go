package ec2

import (
	"encoding/xml"
)

// ============================================================================
// Response Wrapper Types
// ============================================================================
//
// These types wrap Smithy-generated types for XML serialization in EC2 responses.
// The Smithy types (from smithy_types.go) have correct camelCase XML tags.
// SDK types (github.com/aws/aws-sdk-go-v2/service/ec2/types) lack XML tags.
//
// IMPORTANT: XMLName Requirement for EC2 Response Types
// -----------------------------------------------------
// Response types used with BuildEC2Response() MUST have an XMLName field that
// matches the AWS operation name + "Response". The struct name alone does NOT
// control the XML root element - only XMLName does.
//
// Without XMLName, Go's xml.Marshal uses the struct name as the root element,
// causing AWS SDK parsing failures (e.g., "<TagSetResponse>" instead of
// "<DescribeTagsResponse>").
//
// Pattern for Describe* operations:
//
//     type DescribeTagsResponse struct {
//         XMLName xml.Name         `xml:"DescribeTagsResponse"`  // <-- REQUIRED
//         TagSet  []TagDescription `xml:"tagSet>item"`
//     }
//
// Pattern for void operations (Create/Delete/Modify that return only success):
//
//     type CreateTagsResponse struct {
//         XMLName xml.Name `xml:"CreateTagsResponse"`  // <-- REQUIRED
//         Return  bool     `xml:"return"`
//     }
//
// If you add a new response type and get XML parsing errors, check that:
// 1. XMLName field exists with correct xml tag
// 2. The xml tag value matches "{OperationName}Response" exactly

// VpcSetResponse wraps VPCs for DescribeVpcs response using Smithy types
type VpcSetResponse struct {
	VpcSet []Vpc `xml:"vpcSet>item"`
}

// VpcResponse wraps a single VPC for CreateVpc response
type VpcResponse struct {
	XMLName xml.Name `xml:"vpc"`
	Vpc     Vpc      `xml:"vpc"`
}

// SubnetSetResponse wraps subnets for DescribeSubnets response
type SubnetSetResponse struct {
	SubnetSet []Subnet `xml:"subnetSet>item"`
}

// SubnetResponse wraps a single subnet for CreateSubnet response
type SubnetResponse struct {
	Subnet Subnet `xml:"subnet"`
}

// SecurityGroupInfoResponse wraps security groups for DescribeSecurityGroups response
type SecurityGroupInfoResponse struct {
	SecurityGroupInfo []SecurityGroup `xml:"securityGroupInfo>item"`
}

// InternetGatewaySetResponse wraps internet gateways for DescribeInternetGateways response
type InternetGatewaySetResponse struct {
	InternetGatewaySet []InternetGateway `xml:"internetGatewaySet>item"`
}

// InternetGatewayResponse wraps a single internet gateway
type InternetGatewayResponse struct {
	InternetGateway InternetGateway `xml:"internetGateway"`
}

// ImagesSetResponse wraps images for DescribeImages response
type ImagesSetResponse struct {
	ImagesSet []Image `xml:"imagesSet>item"`
}

// VolumeSetResponse wraps volumes for DescribeVolumes response
type VolumeSetResponse struct {
	VolumeSet []Volume `xml:"volumeSet>item"`
}

// KeySetResponse wraps key pairs for DescribeKeyPairs response
type KeySetResponse struct {
	KeySet []KeyPairInfo `xml:"keySet>item"`
}

// LaunchTemplatesResponse wraps launch templates for DescribeLaunchTemplates response
type LaunchTemplatesResponse struct {
	LaunchTemplates []LaunchTemplate `xml:"launchTemplates>item"`
}

// LaunchTemplateResponse wraps a single launch template
type LaunchTemplateResponse struct {
	LaunchTemplate LaunchTemplate `xml:"launchTemplate"`
}

// TagSetResponse wraps tags for DescribeTags response
type TagSetResponse struct {
	XMLName xml.Name         `xml:"DescribeTagsResponse"`
	TagSet  []TagDescription `xml:"tagSet>item"`
}

// ReturnResponse for operations that return a boolean result (generic, used when XMLName not needed)
type ReturnResponse struct {
	Return bool `xml:"return"`
}

// EC2 void operation response types with correct XMLName
// These are needed because EC2 protocol requires {Operation}Response as root element

type AttachInternetGatewayResponse struct {
	XMLName xml.Name `xml:"AttachInternetGatewayResponse"`
	Return  bool     `xml:"return"`
}

type DetachInternetGatewayResponse struct {
	XMLName xml.Name `xml:"DetachInternetGatewayResponse"`
	Return  bool     `xml:"return"`
}

type DeleteInternetGatewayResponse struct {
	XMLName xml.Name `xml:"DeleteInternetGatewayResponse"`
	Return  bool     `xml:"return"`
}

type DeleteVpcResponse struct {
	XMLName xml.Name `xml:"DeleteVpcResponse"`
	Return  bool     `xml:"return"`
}

type ModifyVpcAttributeResponse struct {
	XMLName xml.Name `xml:"ModifyVpcAttributeResponse"`
	Return  bool     `xml:"return"`
}

type DeleteSubnetResponse struct {
	XMLName xml.Name `xml:"DeleteSubnetResponse"`
	Return  bool     `xml:"return"`
}

type DeleteSecurityGroupResponse struct {
	XMLName xml.Name `xml:"DeleteSecurityGroupResponse"`
	Return  bool     `xml:"return"`
}

type AuthorizeSecurityGroupIngressResponse struct {
	XMLName xml.Name `xml:"AuthorizeSecurityGroupIngressResponse"`
	Return  bool     `xml:"return"`
}

type AuthorizeSecurityGroupEgressResponse struct {
	XMLName xml.Name `xml:"AuthorizeSecurityGroupEgressResponse"`
	Return  bool     `xml:"return"`
}

type RevokeSecurityGroupIngressResponse struct {
	XMLName xml.Name `xml:"RevokeSecurityGroupIngressResponse"`
	Return  bool     `xml:"return"`
}

type RevokeSecurityGroupEgressResponse struct {
	XMLName xml.Name `xml:"RevokeSecurityGroupEgressResponse"`
	Return  bool     `xml:"return"`
}

type CreateTagsResponse struct {
	XMLName xml.Name `xml:"CreateTagsResponse"`
	Return  bool     `xml:"return"`
}

type DeleteTagsResponse struct {
	XMLName xml.Name `xml:"DeleteTagsResponse"`
	Return  bool     `xml:"return"`
}

type DeleteVolumeResponse struct {
	XMLName xml.Name `xml:"DeleteVolumeResponse"`
	Return  bool     `xml:"return"`
}

type DeleteKeyPairResponse struct {
	XMLName xml.Name `xml:"DeleteKeyPairResponse"`
	Return  bool     `xml:"return"`
}

// CreateSecurityGroupResponse for CreateSecurityGroup response (distinct from smithy CreateSecurityGroupResult)
type CreateSecurityGroupResponse struct {
	GroupId string `xml:"groupId"`
	Return  bool   `xml:"return"`
}

// KeyPairResponse for CreateKeyPair response
type KeyPairResponse struct {
	KeyPairId      string `xml:"keyPairId"`
	KeyName        string `xml:"keyName"`
	KeyFingerprint string `xml:"keyFingerprint"`
	KeyMaterial    string `xml:"keyMaterial,omitempty"`
}

// KeyPairImportResponse for ImportKeyPair response
type KeyPairImportResponse struct {
	KeyPairId      string `xml:"keyPairId"`
	KeyName        string `xml:"keyName"`
	KeyFingerprint string `xml:"keyFingerprint"`
}

// VpcAttributes stores the modifiable VPC attributes that aren't part of the Vpc type
type VpcAttributes struct {
	EnableDnsHostnames               bool `json:"enableDnsHostnames"`
	EnableDnsSupport                 bool `json:"enableDnsSupport"`
	EnableNetworkAddressUsageMetrics bool `json:"enableNetworkAddressUsageMetrics"`
}

// Note: CreateVolumeResponse, AttachVolumeResponse, DetachVolumeResponse are not needed
// because BuildEC2Response adds the {Operation}Response wrapper and the Smithy types
// (Volume, VolumeAttachment) are passed directly.

// VpcAttributeBooleanValue represents a boolean attribute value for VPC attributes
type VpcAttributeBooleanValue struct {
	Value bool `xml:"value"`
}

// DescribeVpcAttributeResponse is the response for DescribeVpcAttribute
type DescribeVpcAttributeResponse struct {
	VpcId                            string                    `xml:"vpcId"`
	EnableDnsHostnames               *VpcAttributeBooleanValue `xml:"enableDnsHostnames,omitempty"`
	EnableDnsSupport                 *VpcAttributeBooleanValue `xml:"enableDnsSupport,omitempty"`
	EnableNetworkAddressUsageMetrics *VpcAttributeBooleanValue `xml:"enableNetworkAddressUsageMetrics,omitempty"`
}

// NetworkInterfaceSetResponse wraps network interfaces for DescribeNetworkInterfaces response
type NetworkInterfaceSetResponse struct {
	NetworkInterfaces []NetworkInterface `xml:"networkInterfaceSet>item"`
}

// NetworkAclSetResponse wraps network ACLs for DescribeNetworkAcls response
type NetworkAclSetResponse struct {
	NetworkAcls []NetworkAcl `xml:"networkAclSet>item"`
}

// RouteTableSetResponse wraps route tables for DescribeRouteTables response
type RouteTableSetResponse struct {
	RouteTables []RouteTable `xml:"routeTableSet>item"`
}

// InstanceTypeSetResponse wraps instance types for DescribeInstanceTypes response
type InstanceTypeSetResponse struct {
	XMLName       xml.Name           `xml:"DescribeInstanceTypesResponse"`
	InstanceTypes []InstanceTypeInfo `xml:"instanceTypeSet>item"`
}
