package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/ec2"

// NewEc2Client creates an EC2 client.
func NewEc2Client(region string) (*ec2.Client, error) {
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
