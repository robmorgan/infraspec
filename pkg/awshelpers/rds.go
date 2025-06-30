package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/rds"

// NewRdsClient creates an RDS client.
func NewRdsClient(region string) (*rds.Client, error) {
	s, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return rds.NewFromConfig(*s), nil
}

func NewRdsClientWithDefaultRegion() (*rds.Client, error) {
	return NewRdsClient(defaultRegion)
}
