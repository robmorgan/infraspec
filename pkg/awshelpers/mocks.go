package awshelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// EC2API is an interface that matches the methods we use from ec2.Client
type EC2API interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
	DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error)
}

// SSMAPI is an interface that matches the methods we use from ssm.Client
type SSMAPI interface {
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

// MockEC2Client implements EC2API using golden files
type MockEC2Client struct {
	region string
}

// MockSSMClient implements SSMAPI using golden files
type MockSSMClient struct {
	region string
}

// NewMockEC2Client creates a new mock EC2 client
func NewMockEC2Client(region string) *MockEC2Client {
	return &MockEC2Client{region: region}
}

// NewMockSSMClient creates a new mock SSM client
func NewMockSSMClient(region string) *MockSSMClient {
	return &MockSSMClient{region: region}
}

// DescribeRegions loads regions from golden file
func (m *MockEC2Client) DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
	data, err := loadGoldenFile("ec2_describe_regions.json")
	if err != nil {
		return nil, err
	}

	var golden struct {
		Regions []struct {
			RegionName string `json:"regionName"`
			Endpoint   string `json:"endpoint"`
		} `json:"regions"`
	}

	if err := json.Unmarshal(data, &golden); err != nil {
		return nil, err
	}

	output := &ec2.DescribeRegionsOutput{
		Regions: make([]types.Region, len(golden.Regions)),
	}

	for i, r := range golden.Regions {
		output.Regions[i] = types.Region{
			RegionName: aws.String(r.RegionName),
			Endpoint:   aws.String(r.Endpoint),
		}
	}

	return output, nil
}

// DescribeAvailabilityZones loads availability zones from golden file
func (m *MockEC2Client) DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error) {
	filename := fmt.Sprintf("ec2_describe_availability_zones_%s.json", m.region)
	data, err := loadGoldenFile(filename)
	if err != nil {
		return nil, err
	}

	var golden struct {
		AvailabilityZones []struct {
			ZoneName string `json:"zoneName"`
			State    string `json:"state"`
		} `json:"availabilityZones"`
	}

	if err := json.Unmarshal(data, &golden); err != nil {
		return nil, err
	}

	output := &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: make([]types.AvailabilityZone, len(golden.AvailabilityZones)),
	}

	for i, az := range golden.AvailabilityZones {
		state := types.AvailabilityZoneStateAvailable
		output.AvailabilityZones[i] = types.AvailabilityZone{
			ZoneName: aws.String(az.ZoneName),
			State:    state,
		}
	}

	return output, nil
}

// GetParametersByPath loads SSM parameters from golden file
func (m *MockSSMClient) GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	// Extract service name from path
	// Path format: /aws/service/global-infrastructure/services/{serviceName}/regions
	var filename string
	if params.Path != nil && *params.Path != "" {
		// For simplicity, we'll use a predefined file based on the service name
		// In this case, we know it's apigatewayv2
		filename = "ssm_get_parameters_apigatewayv2_regions.json"
	} else {
		return nil, fmt.Errorf("path parameter is required")
	}

	data, err := loadGoldenFile(filename)
	if err != nil {
		return nil, err
	}

	var golden struct {
		Parameters []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"parameters"`
	}

	if err := json.Unmarshal(data, &golden); err != nil {
		return nil, err
	}

	output := &ssm.GetParametersByPathOutput{
		Parameters: make([]ssmtypes.Parameter, len(golden.Parameters)),
	}

	for i, p := range golden.Parameters {
		output.Parameters[i] = ssmtypes.Parameter{
			Name:  aws.String(p.Name),
			Value: aws.String(p.Value),
		}
	}

	return output, nil
}

// loadGoldenFile loads a golden file from testdata directory
func loadGoldenFile(filename string) ([]byte, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	testdataDir := filepath.Join(filepath.Dir(currentFile), "testdata")
	filePath := filepath.Join(testdataDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read golden file %s: %w", filename, err)
	}

	return data, nil
}
