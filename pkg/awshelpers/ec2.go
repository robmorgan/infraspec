package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/ec2"

// NewEc2Client creates an EC2 client that implements EC2API interface.
func NewEc2Client(region string) (EC2API, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	opts := make([]func(*ec2.Options), 0, 1)
	if endpoint, ok := GetVirtualCloudEndpoint("ec2"); ok {
		opts = append(opts, func(o *ec2.Options) {
			o.EndpointResolver = ec2.EndpointResolverFromURL(endpoint)
		})
	}

	return ec2.NewFromConfig(*sess, opts...), nil
}

// NewEc2FullClient creates a full EC2 client (not limited to EC2API interface).
func NewEc2FullClient(region string) (*ec2.Client, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	opts := make([]func(*ec2.Options), 0, 1)
	if endpoint, ok := GetVirtualCloudEndpoint("ec2"); ok {
		opts = append(opts, func(o *ec2.Options) {
			o.EndpointResolver = ec2.EndpointResolverFromURL(endpoint)
		})
	}

	return ec2.NewFromConfig(*sess, opts...), nil
}

// NewEc2FullClientWithDefaultRegion creates an EC2 client with the default region.
func NewEc2FullClientWithDefaultRegion() (*ec2.Client, error) {
	return NewEc2FullClient(defaultRegion)
}
