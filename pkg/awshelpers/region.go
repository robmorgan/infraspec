package awshelpers

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/robmorgan/infraspec/internal/collections"
	"github.com/robmorgan/infraspec/internal/config"
)

// You can set this environment variable to force InfraSpec to use a specific region rather than a random one. This is
// convenient when iterating locally.
const regionOverrideEnvVarName = "INFRASPEC_REGION"

// AWS API calls typically require an AWS region. We typically require the user to set one explicitly, but in some
// cases, this doesn't make sense (e.g., for fetching the list of regions in an account), so for those cases, we use
// this region as a default.
const defaultRegion = "us-east-1"

// Reference for launch dates: https://aws.amazon.com/about-aws/global-infrastructure/
var stableRegions = []string{
	"us-east-1",      // Launched 2006
	"us-east-2",      // Launched 2016
	"us-west-1",      // Launched 2009
	"us-west-2",      // Launched 2011
	"ca-central-1",   // Launched 2016
	"sa-east-1",      // Launched 2011
	"eu-west-1",      // Launched 2007
	"eu-west-2",      // Launched 2016
	"eu-west-3",      // Launched 2017
	"eu-central-1",   // Launched 2014
	"ap-southeast-1", // Launched 2010
	"ap-southeast-2", // Launched 2012
	"ap-northeast-1", // Launched 2011
	"ap-northeast-2", // Launched 2016
	"ap-south-1",     // Launched 2016
	"eu-north-1",     // Launched 2018
}

// GetRandomStableRegion gets a randomly chosen AWS region that is considered stable. Like GetRandomRegion, you can
// further restrict the stable region list using approvedRegions and forbiddenRegions. We consider stable regions to be
// those that have been around for at least 1 year.
// Note that regions in the approvedRegions list that are not considered stable are ignored.
func GetRandomStableRegion(approvedRegions, forbiddenRegions []string) (string, error) {
	regionsToPickFrom := stableRegions
	if len(approvedRegions) > 0 {
		regionsToPickFrom = collections.Intersection[string](regionsToPickFrom, approvedRegions)
	}
	if len(forbiddenRegions) > 0 {
		regionsToPickFrom = collections.Subtract[string](regionsToPickFrom, forbiddenRegions)
	}
	return GetRandomRegion(regionsToPickFrom, nil)
}

// GetRandomRegion gets a randomly chosen AWS region. If approvedRegions is not empty, this will be a region from the
// approvedRegions list; otherwise, this method will fetch the latest list of regions from the AWS APIs and pick one of
// those. If forbiddenRegions is not empty, this method will make sure the returned region is not in the forbiddenRegions list.
func GetRandomRegion(approvedRegions, forbiddenRegions []string) (string, error) {
	regionFromEnvVar := os.Getenv(regionOverrideEnvVarName)
	if regionFromEnvVar != "" {
		config.Logging.Logger.Infof("Using AWS region %s from environment variable %s", regionFromEnvVar, regionOverrideEnvVarName)
		return regionFromEnvVar, nil
	}

	regionsToPickFrom := approvedRegions

	if len(regionsToPickFrom) == 0 {
		allRegions, err := GetAllAwsRegions()
		if err != nil {
			return "", err
		}
		regionsToPickFrom = allRegions
	}

	regionsToPickFrom = collections.Subtract[string](regionsToPickFrom, forbiddenRegions)
	region, found := collections.RandomElement[string](regionsToPickFrom)
	if !found {
		return "", fmt.Errorf("no regions available")
	}

	return region, nil
}

// GetAllAwsRegions gets the list of AWS regions available in this account.
func GetAllAwsRegions() ([]string, error) {
	config.Logging.Logger.Infof("Looking up all AWS regions available in this account")

	ec2Client, err := NewEc2Client(defaultRegion)
	if err != nil {
		return nil, err
	}

	out, err := ec2Client.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	regions := make([]string, len(out.Regions))
	for i := range out.Regions {
		regions[i] = aws.ToString(out.Regions[i].RegionName)
	}

	return regions, nil
}

// GetAvailabilityZones gets the Availability Zones for a given AWS region. Note that for certain regions (e.g. us-east-1), different AWS
// accounts have access to different availability zones.
func GetAvailabilityZones(region string) ([]string, error) {
	config.Logging.Logger.Infof("Looking up all availability zones available in this account for region %s", region)

	ec2Client, err := NewEc2Client(region)
	if err != nil {
		return nil, err
	}

	resp, err := ec2Client.DescribeAvailabilityZones(context.Background(), &ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, err
	}

	out := make([]string, len(resp.AvailabilityZones))
	for i := range resp.AvailabilityZones {
		out[i] = aws.ToString(resp.AvailabilityZones[i].ZoneName)
	}

	return out, nil
}

// GetRegionsForService gets all AWS regions in which a service is available and returns errors.
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-public-parameters-global-infrastructure.html
func GetRegionsForService(serviceName string) ([]string, error) {
	// These values are available in any region, defaulting to us-east-1 since it's the oldest
	ssmClient, err := NewSsmClient("us-east-1")
	if err != nil {
		return nil, err
	}

	paramPath := "/aws/service/global-infrastructure/services/%s/regions"
	resp, err := ssmClient.GetParametersByPath(context.Background(), &ssm.GetParametersByPathInput{
		Path: aws.String(fmt.Sprintf(paramPath, serviceName)),
	})
	if err != nil {
		return nil, err
	}

	availableRegions := make([]string, len(resp.Parameters))
	for i := range resp.Parameters {
		availableRegions[i] = *resp.Parameters[i].Value
	}

	return availableRegions, nil
}

// GetRandomRegionForService retrieves a list of AWS regions in which a service is available
// Then returns one region randomly from the list
func GetRandomRegionForService(serviceName string) (string, error) {
	availableRegions, err := GetRegionsForService(serviceName)
	if err != nil {
		return "", err
	}

	return GetRandomRegion(availableRegions, nil)
}
