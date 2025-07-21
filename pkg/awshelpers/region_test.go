package awshelpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRandomRegion(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	randomRegion, err := GetRandomStableRegion(nil, nil)
	assert.NoError(t, err)
	azs, err := GetAvailabilityZones(randomRegion)
	assert.NoError(t, err)

	// Every AWS account has access to different AZs, so he best we can do is make sure we get at least one back
	assert.True(t, len(azs) > 1)
	for _, az := range azs {
		assert.Regexp(t, fmt.Sprintf("^%s[a-z]$", randomRegion), az)
	}
}

func TestGetRandomRegionForService(t *testing.T) {
	t.Parallel()

	serviceName := "apigatewayv2"

	regionsForService, err := GetRegionsForService(serviceName)
	assert.NoError(t, err)
	randomRegionForService, err := GetRandomRegionForService(serviceName)
	assert.NoError(t, err)

	assert.Contains(t, regionsForService, randomRegionForService)
}
