package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `EC2Asserter` interface.
var _ EC2Asserter = (*AWSAsserter)(nil)

// EC2Asserter defines EC2-specific assertions
type EC2Asserter interface {
	// Instance assertions
	AssertEC2InstanceExists(instanceID, region string) error
	AssertEC2InstanceState(instanceID, state, region string) error
	AssertEC2InstanceType(instanceID, instanceType, region string) error
	AssertEC2InstanceAMI(instanceID, amiID, region string) error
	AssertEC2InstanceSubnet(instanceID, subnetID, region string) error
	AssertEC2InstanceVPC(instanceID, vpcID, region string) error
	AssertEC2InstanceSecurityGroups(instanceID string, securityGroupIDs []string, region string) error
	AssertEC2InstanceTags(instanceID string, expectedTags map[string]string, region string) error

	// VPC assertions
	AssertVPCExists(vpcID, region string) error
	AssertVPCState(vpcID, state, region string) error
	AssertVPCCIDR(vpcID, cidrBlock, region string) error
	AssertVPCIsDefault(vpcID string, isDefault bool, region string) error
	AssertVPCTags(vpcID string, expectedTags map[string]string, region string) error

	// Subnet assertions
	AssertSubnetExists(subnetID, region string) error
	AssertSubnetState(subnetID, state, region string) error
	AssertSubnetCIDR(subnetID, cidrBlock, region string) error
	AssertSubnetVPC(subnetID, vpcID, region string) error
	AssertSubnetAvailabilityZone(subnetID, az, region string) error
	AssertSubnetTags(subnetID string, expectedTags map[string]string, region string) error

	// Security Group assertions
	AssertSecurityGroupExists(groupID, region string) error
	AssertSecurityGroupName(groupID, groupName, region string) error
	AssertSecurityGroupVPC(groupID, vpcID, region string) error
	AssertSecurityGroupDescription(groupID, description, region string) error
	AssertSecurityGroupTags(groupID string, expectedTags map[string]string, region string) error

	// Internet Gateway assertions
	AssertInternetGatewayExists(igwID, region string) error
	AssertInternetGatewayAttachedToVPC(igwID, vpcID, region string) error
	AssertInternetGatewayTags(igwID string, expectedTags map[string]string, region string) error

	// EBS Volume assertions
	AssertEBSVolumeExists(volumeID, region string) error
	AssertEBSVolumeState(volumeID, state, region string) error
	AssertEBSVolumeSize(volumeID string, sizeGB int32, region string) error
	AssertEBSVolumeType(volumeID, volumeType, region string) error
	AssertEBSVolumeTags(volumeID string, expectedTags map[string]string, region string) error

	// Key Pair assertions
	AssertKeyPairExists(keyName, region string) error
}

// ==================== Instance Assertions ====================

// AssertEC2InstanceExists checks if an EC2 instance exists
func (a *AWSAsserter) AssertEC2InstanceExists(instanceID, region string) error {
	_, err := a.getEC2Instance(instanceID, region)
	return err
}

// AssertEC2InstanceState checks if an EC2 instance has the expected state
func (a *AWSAsserter) AssertEC2InstanceState(instanceID, state, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	if instance.State == nil {
		return fmt.Errorf("instance %s has no state", instanceID)
	}

	if string(instance.State.Name) != state {
		return fmt.Errorf("expected instance state %s, but got %s", state, instance.State.Name)
	}

	return nil
}

// AssertEC2InstanceType checks if an EC2 instance has the expected instance type
func (a *AWSAsserter) AssertEC2InstanceType(instanceID, instanceType, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	if string(instance.InstanceType) != instanceType {
		return fmt.Errorf("expected instance type %s, but got %s", instanceType, instance.InstanceType)
	}

	return nil
}

// AssertEC2InstanceAMI checks if an EC2 instance was launched from the expected AMI
func (a *AWSAsserter) AssertEC2InstanceAMI(instanceID, amiID, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.ImageId) != amiID {
		return fmt.Errorf("expected AMI ID %s, but got %s", amiID, aws.ToString(instance.ImageId))
	}

	return nil
}

// AssertEC2InstanceSubnet checks if an EC2 instance is in the expected subnet
func (a *AWSAsserter) AssertEC2InstanceSubnet(instanceID, subnetID, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.SubnetId) != subnetID {
		return fmt.Errorf("expected subnet ID %s, but got %s", subnetID, aws.ToString(instance.SubnetId))
	}

	return nil
}

// AssertEC2InstanceVPC checks if an EC2 instance is in the expected VPC
func (a *AWSAsserter) AssertEC2InstanceVPC(instanceID, vpcID, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	if aws.ToString(instance.VpcId) != vpcID {
		return fmt.Errorf("expected VPC ID %s, but got %s", vpcID, aws.ToString(instance.VpcId))
	}

	return nil
}

// AssertEC2InstanceSecurityGroups checks if an EC2 instance has the expected security groups
func (a *AWSAsserter) AssertEC2InstanceSecurityGroups(instanceID string, securityGroupIDs []string, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	actualSGs := make(map[string]bool)
	for _, sg := range instance.SecurityGroups {
		actualSGs[aws.ToString(sg.GroupId)] = true
	}

	for _, expectedSG := range securityGroupIDs {
		if !actualSGs[expectedSG] {
			return fmt.Errorf("instance %s is not associated with security group %s", instanceID, expectedSG)
		}
	}

	return nil
}

// AssertEC2InstanceTags checks if an EC2 instance has the expected tags
func (a *AWSAsserter) AssertEC2InstanceTags(instanceID string, expectedTags map[string]string, region string) error {
	instance, err := a.getEC2Instance(instanceID, region)
	if err != nil {
		return err
	}

	return a.checkTags(instance.Tags, expectedTags)
}

// ==================== VPC Assertions ====================

// AssertVPCExists checks if a VPC exists
func (a *AWSAsserter) AssertVPCExists(vpcID, region string) error {
	_, err := a.getVPC(vpcID, region)
	return err
}

// AssertVPCState checks if a VPC has the expected state
func (a *AWSAsserter) AssertVPCState(vpcID, state, region string) error {
	vpc, err := a.getVPC(vpcID, region)
	if err != nil {
		return err
	}

	if string(vpc.State) != state {
		return fmt.Errorf("expected VPC state %s, but got %s", state, vpc.State)
	}

	return nil
}

// AssertVPCCIDR checks if a VPC has the expected CIDR block
func (a *AWSAsserter) AssertVPCCIDR(vpcID, cidrBlock, region string) error {
	vpc, err := a.getVPC(vpcID, region)
	if err != nil {
		return err
	}

	if aws.ToString(vpc.CidrBlock) != cidrBlock {
		return fmt.Errorf("expected CIDR block %s, but got %s", cidrBlock, aws.ToString(vpc.CidrBlock))
	}

	return nil
}

// AssertVPCIsDefault checks if a VPC is or is not the default VPC
func (a *AWSAsserter) AssertVPCIsDefault(vpcID string, isDefault bool, region string) error {
	vpc, err := a.getVPC(vpcID, region)
	if err != nil {
		return err
	}

	if aws.ToBool(vpc.IsDefault) != isDefault {
		return fmt.Errorf("expected VPC IsDefault to be %t, but got %t", isDefault, aws.ToBool(vpc.IsDefault))
	}

	return nil
}

// AssertVPCTags checks if a VPC has the expected tags
func (a *AWSAsserter) AssertVPCTags(vpcID string, expectedTags map[string]string, region string) error {
	vpc, err := a.getVPC(vpcID, region)
	if err != nil {
		return err
	}

	return a.checkTags(vpc.Tags, expectedTags)
}

// ==================== Subnet Assertions ====================

// AssertSubnetExists checks if a subnet exists
func (a *AWSAsserter) AssertSubnetExists(subnetID, region string) error {
	_, err := a.getSubnet(subnetID, region)
	return err
}

// AssertSubnetState checks if a subnet has the expected state
func (a *AWSAsserter) AssertSubnetState(subnetID, state, region string) error {
	subnet, err := a.getSubnet(subnetID, region)
	if err != nil {
		return err
	}

	if string(subnet.State) != state {
		return fmt.Errorf("expected subnet state %s, but got %s", state, subnet.State)
	}

	return nil
}

// AssertSubnetCIDR checks if a subnet has the expected CIDR block
func (a *AWSAsserter) AssertSubnetCIDR(subnetID, cidrBlock, region string) error {
	subnet, err := a.getSubnet(subnetID, region)
	if err != nil {
		return err
	}

	if aws.ToString(subnet.CidrBlock) != cidrBlock {
		return fmt.Errorf("expected CIDR block %s, but got %s", cidrBlock, aws.ToString(subnet.CidrBlock))
	}

	return nil
}

// AssertSubnetVPC checks if a subnet belongs to the expected VPC
func (a *AWSAsserter) AssertSubnetVPC(subnetID, vpcID, region string) error {
	subnet, err := a.getSubnet(subnetID, region)
	if err != nil {
		return err
	}

	if aws.ToString(subnet.VpcId) != vpcID {
		return fmt.Errorf("expected VPC ID %s, but got %s", vpcID, aws.ToString(subnet.VpcId))
	}

	return nil
}

// AssertSubnetAvailabilityZone checks if a subnet is in the expected availability zone
func (a *AWSAsserter) AssertSubnetAvailabilityZone(subnetID, az, region string) error {
	subnet, err := a.getSubnet(subnetID, region)
	if err != nil {
		return err
	}

	if aws.ToString(subnet.AvailabilityZone) != az {
		return fmt.Errorf("expected availability zone %s, but got %s", az, aws.ToString(subnet.AvailabilityZone))
	}

	return nil
}

// AssertSubnetTags checks if a subnet has the expected tags
func (a *AWSAsserter) AssertSubnetTags(subnetID string, expectedTags map[string]string, region string) error {
	subnet, err := a.getSubnet(subnetID, region)
	if err != nil {
		return err
	}

	return a.checkTags(subnet.Tags, expectedTags)
}

// ==================== Security Group Assertions ====================

// AssertSecurityGroupExists checks if a security group exists
func (a *AWSAsserter) AssertSecurityGroupExists(groupID, region string) error {
	_, err := a.getSecurityGroup(groupID, region)
	return err
}

// AssertSecurityGroupName checks if a security group has the expected name
func (a *AWSAsserter) AssertSecurityGroupName(groupID, groupName, region string) error {
	sg, err := a.getSecurityGroup(groupID, region)
	if err != nil {
		return err
	}

	if aws.ToString(sg.GroupName) != groupName {
		return fmt.Errorf("expected security group name %s, but got %s", groupName, aws.ToString(sg.GroupName))
	}

	return nil
}

// AssertSecurityGroupVPC checks if a security group belongs to the expected VPC
func (a *AWSAsserter) AssertSecurityGroupVPC(groupID, vpcID, region string) error {
	sg, err := a.getSecurityGroup(groupID, region)
	if err != nil {
		return err
	}

	if aws.ToString(sg.VpcId) != vpcID {
		return fmt.Errorf("expected VPC ID %s, but got %s", vpcID, aws.ToString(sg.VpcId))
	}

	return nil
}

// AssertSecurityGroupDescription checks if a security group has the expected description
func (a *AWSAsserter) AssertSecurityGroupDescription(groupID, description, region string) error {
	sg, err := a.getSecurityGroup(groupID, region)
	if err != nil {
		return err
	}

	if aws.ToString(sg.Description) != description {
		return fmt.Errorf("expected description %s, but got %s", description, aws.ToString(sg.Description))
	}

	return nil
}

// AssertSecurityGroupTags checks if a security group has the expected tags
func (a *AWSAsserter) AssertSecurityGroupTags(groupID string, expectedTags map[string]string, region string) error {
	sg, err := a.getSecurityGroup(groupID, region)
	if err != nil {
		return err
	}

	return a.checkTags(sg.Tags, expectedTags)
}

// ==================== Internet Gateway Assertions ====================

// AssertInternetGatewayExists checks if an internet gateway exists
func (a *AWSAsserter) AssertInternetGatewayExists(igwID, region string) error {
	_, err := a.getInternetGateway(igwID, region)
	return err
}

// AssertInternetGatewayAttachedToVPC checks if an internet gateway is attached to the expected VPC
func (a *AWSAsserter) AssertInternetGatewayAttachedToVPC(igwID, vpcID, region string) error {
	igw, err := a.getInternetGateway(igwID, region)
	if err != nil {
		return err
	}

	for _, attachment := range igw.Attachments {
		if aws.ToString(attachment.VpcId) == vpcID {
			return nil
		}
	}

	return fmt.Errorf("internet gateway %s is not attached to VPC %s", igwID, vpcID)
}

// AssertInternetGatewayTags checks if an internet gateway has the expected tags
func (a *AWSAsserter) AssertInternetGatewayTags(igwID string, expectedTags map[string]string, region string) error {
	igw, err := a.getInternetGateway(igwID, region)
	if err != nil {
		return err
	}

	return a.checkTags(igw.Tags, expectedTags)
}

// ==================== EBS Volume Assertions ====================

// AssertEBSVolumeExists checks if an EBS volume exists
func (a *AWSAsserter) AssertEBSVolumeExists(volumeID, region string) error {
	_, err := a.getEBSVolume(volumeID, region)
	return err
}

// AssertEBSVolumeState checks if an EBS volume has the expected state
func (a *AWSAsserter) AssertEBSVolumeState(volumeID, state, region string) error {
	volume, err := a.getEBSVolume(volumeID, region)
	if err != nil {
		return err
	}

	if string(volume.State) != state {
		return fmt.Errorf("expected volume state %s, but got %s", state, volume.State)
	}

	return nil
}

// AssertEBSVolumeSize checks if an EBS volume has the expected size
func (a *AWSAsserter) AssertEBSVolumeSize(volumeID string, sizeGB int32, region string) error {
	volume, err := a.getEBSVolume(volumeID, region)
	if err != nil {
		return err
	}

	if aws.ToInt32(volume.Size) != sizeGB {
		return fmt.Errorf("expected volume size %d GB, but got %d GB", sizeGB, aws.ToInt32(volume.Size))
	}

	return nil
}

// AssertEBSVolumeType checks if an EBS volume has the expected type
func (a *AWSAsserter) AssertEBSVolumeType(volumeID, volumeType, region string) error {
	volume, err := a.getEBSVolume(volumeID, region)
	if err != nil {
		return err
	}

	if string(volume.VolumeType) != volumeType {
		return fmt.Errorf("expected volume type %s, but got %s", volumeType, volume.VolumeType)
	}

	return nil
}

// AssertEBSVolumeTags checks if an EBS volume has the expected tags
func (a *AWSAsserter) AssertEBSVolumeTags(volumeID string, expectedTags map[string]string, region string) error {
	volume, err := a.getEBSVolume(volumeID, region)
	if err != nil {
		return err
	}

	return a.checkTags(volume.Tags, expectedTags)
}

// ==================== Key Pair Assertions ====================

// AssertKeyPairExists checks if a key pair exists
func (a *AWSAsserter) AssertKeyPairExists(keyName, region string) error {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return err
	}

	input := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	}

	result, err := client.DescribeKeyPairs(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error describing key pair %s: %w", keyName, err)
	}

	if len(result.KeyPairs) == 0 {
		return fmt.Errorf("key pair %s does not exist", keyName)
	}

	return nil
}

// ==================== Helper Methods ====================

// getEC2Instance retrieves an EC2 instance by ID
func (a *AWSAsserter) getEC2Instance(instanceID, region string) (*types.Instance, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing instance %s: %w", instanceID, err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance %s does not exist", instanceID)
	}

	return &result.Reservations[0].Instances[0], nil
}

// getVPC retrieves a VPC by ID
func (a *AWSAsserter) getVPC(vpcID, region string) (*types.Vpc, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	}

	result, err := client.DescribeVpcs(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing VPC %s: %w", vpcID, err)
	}

	if len(result.Vpcs) == 0 {
		return nil, fmt.Errorf("VPC %s does not exist", vpcID)
	}

	return &result.Vpcs[0], nil
}

// getSubnet retrieves a subnet by ID
func (a *AWSAsserter) getSubnet(subnetID, region string) (*types.Subnet, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}

	result, err := client.DescribeSubnets(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing subnet %s: %w", subnetID, err)
	}

	if len(result.Subnets) == 0 {
		return nil, fmt.Errorf("subnet %s does not exist", subnetID)
	}

	return &result.Subnets[0], nil
}

// getSecurityGroup retrieves a security group by ID
func (a *AWSAsserter) getSecurityGroup(groupID, region string) (*types.SecurityGroup, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupID},
	}

	result, err := client.DescribeSecurityGroups(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing security group %s: %w", groupID, err)
	}

	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("security group %s does not exist", groupID)
	}

	return &result.SecurityGroups[0], nil
}

// getInternetGateway retrieves an internet gateway by ID
func (a *AWSAsserter) getInternetGateway(igwID, region string) (*types.InternetGateway, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{igwID},
	}

	result, err := client.DescribeInternetGateways(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing internet gateway %s: %w", igwID, err)
	}

	if len(result.InternetGateways) == 0 {
		return nil, fmt.Errorf("internet gateway %s does not exist", igwID)
	}

	return &result.InternetGateways[0], nil
}

// getEBSVolume retrieves an EBS volume by ID
func (a *AWSAsserter) getEBSVolume(volumeID, region string) (*types.Volume, error) {
	client, err := awshelpers.NewEc2FullClient(region)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeVolumesInput{
		VolumeIds: []string{volumeID},
	}

	result, err := client.DescribeVolumes(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("error describing volume %s: %w", volumeID, err)
	}

	if len(result.Volumes) == 0 {
		return nil, fmt.Errorf("volume %s does not exist", volumeID)
	}

	return &result.Volumes[0], nil
}

// checkTags compares expected tags against actual tags
func (a *AWSAsserter) checkTags(actualTags []types.Tag, expectedTags map[string]string) error {
	tagMap := make(map[string]string)
	for _, tag := range actualTags {
		tagMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	for key, value := range expectedTags {
		actualValue, exists := tagMap[key]
		if !exists {
			return fmt.Errorf("expected tag %s not found", key)
		}
		if actualValue != value {
			return fmt.Errorf("expected tag %s to have value %s, but got %s", key, value, actualValue)
		}
	}

	return nil
}
