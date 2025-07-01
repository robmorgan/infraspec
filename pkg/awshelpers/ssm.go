package awshelpers

import (
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// NewSsmClient creates an SSM client.
func NewSsmClient(region string) (*ssm.Client, error) {
	s, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ssm.NewFromConfig(*s), nil
}
