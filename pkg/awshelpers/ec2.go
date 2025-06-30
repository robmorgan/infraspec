package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/ec2"

// NewEc2Client creates an EC2 client.
func NewEc2Client(region string) (*ec2.Client, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ec2.NewFromConfig(*sess), nil
}
