package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// EC2 Step Definitions
func registerEC2Steps(sc *godog.ScenarioContext) {
	// Instance steps with direct IDs
	sc.Step(`^the EC2 instance "([^"]*)" should exist$`, newEC2InstanceExistsStep)
	sc.Step(`^the EC2 instance "([^"]*)" state should be "([^"]*)"$`, newEC2InstanceStateStep)
	sc.Step(`^the EC2 instance "([^"]*)" instance type should be "([^"]*)"$`, newEC2InstanceTypeStep)
	sc.Step(`^the EC2 instance "([^"]*)" AMI should be "([^"]*)"$`, newEC2InstanceAMIStep)
	sc.Step(`^the EC2 instance "([^"]*)" should be in subnet "([^"]*)"$`, newEC2InstanceSubnetStep)
	sc.Step(`^the EC2 instance "([^"]*)" should be in VPC "([^"]*)"$`, newEC2InstanceVPCStep)
	sc.Step(`^the EC2 instance "([^"]*)" should have the tags$`, newEC2InstanceTagsStep)

	// Instance steps reading from Terraform output
	sc.Step(`^the EC2 instance from output "([^"]*)" should exist$`, newEC2InstanceFromOutputExistsStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" state should be "([^"]*)"$`, newEC2InstanceFromOutputStateStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" instance type should be "([^"]*)"$`, newEC2InstanceFromOutputTypeStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" AMI should be "([^"]*)"$`, newEC2InstanceFromOutputAMIStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" should be in subnet "([^"]*)"$`, newEC2InstanceFromOutputSubnetStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" should be in VPC "([^"]*)"$`, newEC2InstanceFromOutputVPCStep)
	sc.Step(`^the EC2 instance from output "([^"]*)" should have the tags$`, newEC2InstanceFromOutputTagsStep)

	// VPC steps with direct IDs
	sc.Step(`^the VPC "([^"]*)" should exist$`, newVPCExistsStep)
	sc.Step(`^the VPC "([^"]*)" state should be "([^"]*)"$`, newVPCStateStep)
	sc.Step(`^the VPC "([^"]*)" CIDR block should be "([^"]*)"$`, newVPCCIDRStep)
	sc.Step(`^the VPC "([^"]*)" should be the default VPC$`, newVPCIsDefaultStep)
	sc.Step(`^the VPC "([^"]*)" should not be the default VPC$`, newVPCIsNotDefaultStep)
	sc.Step(`^the VPC "([^"]*)" should have the tags$`, newVPCTagsStep)

	// VPC steps reading from Terraform output
	sc.Step(`^the VPC from output "([^"]*)" should exist$`, newVPCFromOutputExistsStep)
	sc.Step(`^the VPC from output "([^"]*)" state should be "([^"]*)"$`, newVPCFromOutputStateStep)
	sc.Step(`^the VPC from output "([^"]*)" CIDR block should be "([^"]*)"$`, newVPCFromOutputCIDRStep)
	sc.Step(`^the VPC from output "([^"]*)" should have the tags$`, newVPCFromOutputTagsStep)

	// Subnet steps with direct IDs
	sc.Step(`^the subnet "([^"]*)" should exist$`, newSubnetExistsStep)
	sc.Step(`^the subnet "([^"]*)" state should be "([^"]*)"$`, newSubnetStateStep)
	sc.Step(`^the subnet "([^"]*)" CIDR block should be "([^"]*)"$`, newSubnetCIDRStep)
	sc.Step(`^the subnet "([^"]*)" should be in VPC "([^"]*)"$`, newSubnetVPCStep)
	sc.Step(`^the subnet "([^"]*)" availability zone should be "([^"]*)"$`, newSubnetAZStep)
	sc.Step(`^the subnet "([^"]*)" should have the tags$`, newSubnetTagsStep)

	// Subnet steps reading from Terraform output
	sc.Step(`^the subnet from output "([^"]*)" should exist$`, newSubnetFromOutputExistsStep)
	sc.Step(`^the subnet from output "([^"]*)" state should be "([^"]*)"$`, newSubnetFromOutputStateStep)
	sc.Step(`^the subnet from output "([^"]*)" CIDR block should be "([^"]*)"$`, newSubnetFromOutputCIDRStep)
	sc.Step(`^the subnet from output "([^"]*)" should be in VPC "([^"]*)"$`, newSubnetFromOutputVPCStep)
	sc.Step(`^the subnet from output "([^"]*)" availability zone should be "([^"]*)"$`, newSubnetFromOutputAZStep)
	sc.Step(`^the subnet from output "([^"]*)" should have the tags$`, newSubnetFromOutputTagsStep)

	// Security Group steps with direct IDs
	sc.Step(`^the security group "([^"]*)" should exist$`, newSecurityGroupExistsStep)
	sc.Step(`^the security group "([^"]*)" name should be "([^"]*)"$`, newSecurityGroupNameStep)
	sc.Step(`^the security group "([^"]*)" should be in VPC "([^"]*)"$`, newSecurityGroupVPCStep)
	sc.Step(`^the security group "([^"]*)" description should be "([^"]*)"$`, newSecurityGroupDescriptionStep)
	sc.Step(`^the security group "([^"]*)" should have the tags$`, newSecurityGroupTagsStep)

	// Security Group steps reading from Terraform output
	sc.Step(`^the security group from output "([^"]*)" should exist$`, newSecurityGroupFromOutputExistsStep)
	sc.Step(`^the security group from output "([^"]*)" name should be "([^"]*)"$`, newSecurityGroupFromOutputNameStep)
	sc.Step(`^the security group from output "([^"]*)" should be in VPC "([^"]*)"$`, newSecurityGroupFromOutputVPCStep)
	sc.Step(`^the security group from output "([^"]*)" should have the tags$`, newSecurityGroupFromOutputTagsStep)

	// Internet Gateway steps with direct IDs
	sc.Step(`^the internet gateway "([^"]*)" should exist$`, newInternetGatewayExistsStep)
	sc.Step(`^the internet gateway "([^"]*)" should be attached to VPC "([^"]*)"$`, newInternetGatewayAttachedStep)
	sc.Step(`^the internet gateway "([^"]*)" should have the tags$`, newInternetGatewayTagsStep)

	// Internet Gateway steps reading from Terraform output
	sc.Step(`^the internet gateway from output "([^"]*)" should exist$`, newInternetGatewayFromOutputExistsStep)
	sc.Step(`^the internet gateway from output "([^"]*)" should be attached to VPC "([^"]*)"$`, newInternetGatewayFromOutputAttachedStep)
	sc.Step(`^the internet gateway from output "([^"]*)" should have the tags$`, newInternetGatewayFromOutputTagsStep)

	// EBS Volume steps with direct IDs
	sc.Step(`^the EBS volume "([^"]*)" should exist$`, newEBSVolumeExistsStep)
	sc.Step(`^the EBS volume "([^"]*)" state should be "([^"]*)"$`, newEBSVolumeStateStep)
	sc.Step(`^the EBS volume "([^"]*)" size should be (\d+) GB$`, newEBSVolumeSizeStep)
	sc.Step(`^the EBS volume "([^"]*)" type should be "([^"]*)"$`, newEBSVolumeTypeStep)
	sc.Step(`^the EBS volume "([^"]*)" should have the tags$`, newEBSVolumeTagsStep)

	// EBS Volume steps reading from Terraform output
	sc.Step(`^the EBS volume from output "([^"]*)" should exist$`, newEBSVolumeFromOutputExistsStep)
	sc.Step(`^the EBS volume from output "([^"]*)" state should be "([^"]*)"$`, newEBSVolumeFromOutputStateStep)
	sc.Step(`^the EBS volume from output "([^"]*)" size should be (\d+) GB$`, newEBSVolumeFromOutputSizeStep)
	sc.Step(`^the EBS volume from output "([^"]*)" type should be "([^"]*)"$`, newEBSVolumeFromOutputTypeStep)
	sc.Step(`^the EBS volume from output "([^"]*)" should have the tags$`, newEBSVolumeFromOutputTagsStep)

	// Key Pair steps
	sc.Step(`^the key pair "([^"]*)" should exist$`, newKeyPairExistsStep)
	sc.Step(`^the key pair from output "([^"]*)" should exist$`, newKeyPairFromOutputExistsStep)
}

// ==================== Instance Steps ====================

func newEC2InstanceExistsStep(ctx context.Context, instanceID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceExists(instanceID, region)
}

func newEC2InstanceStateStep(ctx context.Context, instanceID, state string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceState(instanceID, state, region)
}

func newEC2InstanceTypeStep(ctx context.Context, instanceID, instanceType string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceType(instanceID, instanceType, region)
}

func newEC2InstanceAMIStep(ctx context.Context, instanceID, amiID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceAMI(instanceID, amiID, region)
}

func newEC2InstanceSubnetStep(ctx context.Context, instanceID, subnetID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceSubnet(instanceID, subnetID, region)
}

func newEC2InstanceVPCStep(ctx context.Context, instanceID, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceVPC(instanceID, vpcID, region)
}

func newEC2InstanceTagsStep(ctx context.Context, instanceID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEC2InstanceTags(instanceID, tags, region)
}

// Instance steps from Terraform output

func newEC2InstanceFromOutputExistsStep(ctx context.Context, outputName string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceExistsStep(ctx, instanceID)
}

func newEC2InstanceFromOutputStateStep(ctx context.Context, outputName, state string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceStateStep(ctx, instanceID, state)
}

func newEC2InstanceFromOutputTypeStep(ctx context.Context, outputName, instanceType string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceTypeStep(ctx, instanceID, instanceType)
}

func newEC2InstanceFromOutputAMIStep(ctx context.Context, outputName, amiID string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceAMIStep(ctx, instanceID, amiID)
}

func newEC2InstanceFromOutputSubnetStep(ctx context.Context, outputName, subnetID string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceSubnetStep(ctx, instanceID, subnetID)
}

func newEC2InstanceFromOutputVPCStep(ctx context.Context, outputName, vpcID string) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceVPCStep(ctx, instanceID, vpcID)
}

func newEC2InstanceFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	instanceID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEC2InstanceTagsStep(ctx, instanceID, table)
}

// ==================== VPC Steps ====================

func newVPCExistsStep(ctx context.Context, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCExists(vpcID, region)
}

func newVPCStateStep(ctx context.Context, vpcID, state string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCState(vpcID, state, region)
}

func newVPCCIDRStep(ctx context.Context, vpcID, cidrBlock string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCCIDR(vpcID, cidrBlock, region)
}

func newVPCIsDefaultStep(ctx context.Context, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCIsDefault(vpcID, true, region)
}

func newVPCIsNotDefaultStep(ctx context.Context, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCIsDefault(vpcID, false, region)
}

func newVPCTagsStep(ctx context.Context, vpcID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertVPCTags(vpcID, tags, region)
}

// VPC steps from Terraform output

func newVPCFromOutputExistsStep(ctx context.Context, outputName string) error {
	vpcID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newVPCExistsStep(ctx, vpcID)
}

func newVPCFromOutputStateStep(ctx context.Context, outputName, state string) error {
	vpcID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newVPCStateStep(ctx, vpcID, state)
}

func newVPCFromOutputCIDRStep(ctx context.Context, outputName, cidrBlock string) error {
	vpcID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newVPCCIDRStep(ctx, vpcID, cidrBlock)
}

func newVPCFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	vpcID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newVPCTagsStep(ctx, vpcID, table)
}

// ==================== Subnet Steps ====================

func newSubnetExistsStep(ctx context.Context, subnetID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetExists(subnetID, region)
}

func newSubnetStateStep(ctx context.Context, subnetID, state string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetState(subnetID, state, region)
}

func newSubnetCIDRStep(ctx context.Context, subnetID, cidrBlock string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetCIDR(subnetID, cidrBlock, region)
}

func newSubnetVPCStep(ctx context.Context, subnetID, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetVPC(subnetID, vpcID, region)
}

func newSubnetAZStep(ctx context.Context, subnetID, az string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetAvailabilityZone(subnetID, az, region)
}

func newSubnetTagsStep(ctx context.Context, subnetID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSubnetTags(subnetID, tags, region)
}

// Subnet steps from Terraform output

func newSubnetFromOutputExistsStep(ctx context.Context, outputName string) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetExistsStep(ctx, subnetID)
}

func newSubnetFromOutputStateStep(ctx context.Context, outputName, state string) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetStateStep(ctx, subnetID, state)
}

func newSubnetFromOutputCIDRStep(ctx context.Context, outputName, cidrBlock string) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetCIDRStep(ctx, subnetID, cidrBlock)
}

func newSubnetFromOutputVPCStep(ctx context.Context, outputName, vpcID string) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetVPCStep(ctx, subnetID, vpcID)
}

func newSubnetFromOutputAZStep(ctx context.Context, outputName, az string) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetAZStep(ctx, subnetID, az)
}

func newSubnetFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	subnetID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSubnetTagsStep(ctx, subnetID, table)
}

// ==================== Security Group Steps ====================

func newSecurityGroupExistsStep(ctx context.Context, groupID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSecurityGroupExists(groupID, region)
}

func newSecurityGroupNameStep(ctx context.Context, groupID, groupName string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSecurityGroupName(groupID, groupName, region)
}

func newSecurityGroupVPCStep(ctx context.Context, groupID, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSecurityGroupVPC(groupID, vpcID, region)
}

func newSecurityGroupDescriptionStep(ctx context.Context, groupID, description string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSecurityGroupDescription(groupID, description, region)
}

func newSecurityGroupTagsStep(ctx context.Context, groupID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertSecurityGroupTags(groupID, tags, region)
}

// Security Group steps from Terraform output

func newSecurityGroupFromOutputExistsStep(ctx context.Context, outputName string) error {
	groupID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSecurityGroupExistsStep(ctx, groupID)
}

func newSecurityGroupFromOutputNameStep(ctx context.Context, outputName, groupName string) error {
	groupID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSecurityGroupNameStep(ctx, groupID, groupName)
}

func newSecurityGroupFromOutputVPCStep(ctx context.Context, outputName, vpcID string) error {
	groupID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSecurityGroupVPCStep(ctx, groupID, vpcID)
}

func newSecurityGroupFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	groupID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSecurityGroupTagsStep(ctx, groupID, table)
}

// ==================== Internet Gateway Steps ====================

func newInternetGatewayExistsStep(ctx context.Context, igwID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertInternetGatewayExists(igwID, region)
}

func newInternetGatewayAttachedStep(ctx context.Context, igwID, vpcID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertInternetGatewayAttachedToVPC(igwID, vpcID, region)
}

func newInternetGatewayTagsStep(ctx context.Context, igwID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertInternetGatewayTags(igwID, tags, region)
}

// Internet Gateway steps from Terraform output

func newInternetGatewayFromOutputExistsStep(ctx context.Context, outputName string) error {
	igwID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newInternetGatewayExistsStep(ctx, igwID)
}

func newInternetGatewayFromOutputAttachedStep(ctx context.Context, outputName, vpcID string) error {
	igwID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newInternetGatewayAttachedStep(ctx, igwID, vpcID)
}

func newInternetGatewayFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	igwID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newInternetGatewayTagsStep(ctx, igwID, table)
}

// ==================== EBS Volume Steps ====================

func newEBSVolumeExistsStep(ctx context.Context, volumeID string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEBSVolumeExists(volumeID, region)
}

func newEBSVolumeStateStep(ctx context.Context, volumeID, state string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEBSVolumeState(volumeID, state, region)
}

func newEBSVolumeSizeStep(ctx context.Context, volumeID string, sizeGB int) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEBSVolumeSize(volumeID, int32(sizeGB), region)
}

func newEBSVolumeTypeStep(ctx context.Context, volumeID, volumeType string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEBSVolumeType(volumeID, volumeType, region)
}

func newEBSVolumeTagsStep(ctx context.Context, volumeID string, table *godog.Table) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	tags := tableToTags(table)

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertEBSVolumeTags(volumeID, tags, region)
}

// EBS Volume steps from Terraform output

func newEBSVolumeFromOutputExistsStep(ctx context.Context, outputName string) error {
	volumeID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEBSVolumeExistsStep(ctx, volumeID)
}

func newEBSVolumeFromOutputStateStep(ctx context.Context, outputName, state string) error {
	volumeID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEBSVolumeStateStep(ctx, volumeID, state)
}

func newEBSVolumeFromOutputSizeStep(ctx context.Context, outputName string, sizeGB int) error {
	volumeID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEBSVolumeSizeStep(ctx, volumeID, sizeGB)
}

func newEBSVolumeFromOutputTypeStep(ctx context.Context, outputName, volumeType string) error {
	volumeID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEBSVolumeTypeStep(ctx, volumeID, volumeType)
}

func newEBSVolumeFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	volumeID, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newEBSVolumeTagsStep(ctx, volumeID, table)
}

// ==================== Key Pair Steps ====================

func newKeyPairExistsStep(ctx context.Context, keyName string) error {
	asserter, err := getEC2Asserter(ctx)
	if err != nil {
		return err
	}

	region := contexthelpers.GetAwsRegion(ctx)
	if region == "" {
		return fmt.Errorf("no AWS region available")
	}

	return asserter.AssertKeyPairExists(keyName, region)
}

func newKeyPairFromOutputExistsStep(ctx context.Context, outputName string) error {
	keyName, err := getResourceIDFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newKeyPairExistsStep(ctx, keyName)
}

// ==================== Helper Functions ====================

func getEC2Asserter(ctx context.Context) (aws.EC2Asserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	ec2Assert, ok := asserter.(aws.EC2Asserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement EC2Asserter")
	}
	return ec2Assert, nil
}

func getResourceIDFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	resourceID, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get resource ID from output %s: %w", outputName, err)
	}
	return resourceID, nil
}

func tableToTags(table *godog.Table) map[string]string {
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}
	return tags
}

// strconv is used for potential boolean parsing in future steps
var _ = strconv.ParseBool
