package awshelpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupMocks configures the mock clients for testing
func setupMocks(t *testing.T) func() {
	t.Helper()

	// Store original factories
	originalEC2Factory := ec2ClientFactory
	originalSSMFactory := ssmClientFactory

	// Replace with mock factories
	ec2ClientFactory = func(region string) (EC2API, error) {
		return NewMockEC2Client(region), nil
	}

	ssmClientFactory = func(region string) (SSMAPI, error) {
		return NewMockSSMClient(region), nil
	}

	// Return cleanup function
	return func() {
		ec2ClientFactory = originalEC2Factory
		ssmClientFactory = originalSSMFactory
	}
}

func TestGetRandomRegion(t *testing.T) {
	cleanup := setupMocks(t)
	defer cleanup()

	randomRegion, err := GetRandomRegion(nil, nil)
	assert.NoError(t, err)
	assertLooksLikeRegionName(t, randomRegion)
}

func TestGetRandomRegionExcludesForbiddenRegions(t *testing.T) {
	t.Parallel()

	approvedRegions := []string{"ca-central-1", "us-east-1", "us-east-2", "us-west-1", "us-west-2", "eu-west-1", "eu-west-2", "eu-central-1", "ap-southeast-1", "ap-northeast-1", "ap-northeast-2", "ap-south-1"}
	forbiddenRegions := []string{"us-west-2", "ap-northeast-2"}

	for i := 0; i < 1000; i++ {
		randomRegion, err := GetRandomRegion(approvedRegions, forbiddenRegions)
		assert.NoError(t, err)
		assert.NotContains(t, forbiddenRegions, randomRegion)
	}
}

func TestGetAllAwsRegions(t *testing.T) {
	cleanup := setupMocks(t)
	defer cleanup()

	regions, err := GetAllAwsRegions()
	assert.NoError(t, err)

	// The typical account had access to 15 regions as of April, 2018: https://aws.amazon.com/about-aws/global-infrastructure/
	assert.True(t, len(regions) >= 15, "Number of regions: %d", len(regions))
	for _, region := range regions {
		assertLooksLikeRegionName(t, region)
	}
}

func assertLooksLikeRegionName(t *testing.T, regionName string) {
	t.Helper()
	assert.Regexp(t, "[a-z]{2}-[a-z]+?-[[:digit:]]+", regionName)
}

func TestGetAvailabilityZones(t *testing.T) {
	cleanup := setupMocks(t)
	defer cleanup()

	// Use us-east-1 since we have a golden file for it
	azs, err := GetAvailabilityZones("us-east-1")
	assert.NoError(t, err)

	// Every AWS account has access to different AZs, so he best we can do is make sure we get at least one back
	assert.True(t, len(azs) > 1)
	for _, az := range azs {
		assert.Regexp(t, fmt.Sprintf("^%s[a-z]$", "us-east-1"), az)
	}
}

func TestGetRandomRegionForService(t *testing.T) {
	cleanup := setupMocks(t)
	defer cleanup()

	serviceName := "apigatewayv2"

	regionsForService, err := GetRegionsForService(serviceName)
	assert.NoError(t, err)
	randomRegionForService, err := GetRandomRegionForService(serviceName)
	assert.NoError(t, err)

	assert.Contains(t, regionsForService, randomRegionForService)
}
