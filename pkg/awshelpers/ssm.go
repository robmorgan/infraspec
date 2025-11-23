package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/ssm"

// NewSsmClient creates an SSM client.
func NewSsmClient(region string) (SSMAPI, error) {
	s, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	opts := make([]func(*ssm.Options), 0, 1)
	if endpoint, ok := GetVirtualCloudEndpoint("ssm"); ok {
		opts = append(opts, func(o *ssm.Options) {
			o.EndpointResolver = ssm.EndpointResolverFromURL(endpoint)
		})
	}

	return ssm.NewFromConfig(*s, opts...), nil
}
