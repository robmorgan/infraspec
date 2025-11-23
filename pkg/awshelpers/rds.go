package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/rds"

// NewRdsClient creates an RDS client.
func NewRdsClient(region string) (*rds.Client, error) {
	s, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	opts := make([]func(*rds.Options), 0, 1)
	if endpoint, ok := GetVirtualCloudEndpoint("rds"); ok {
		opts = append(opts, func(o *rds.Options) {
			o.EndpointResolver = rds.EndpointResolverFromURL(endpoint)
		})
	}

	return rds.NewFromConfig(*s, opts...), nil
}

// NewRdsClientWithDefaultRegion creates an RDS client with the default region.
func NewRdsClientWithDefaultRegion() (*rds.Client, error) {
	return NewRdsClient(defaultRegion)
}
