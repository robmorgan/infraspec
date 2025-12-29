package awshelpers

import "github.com/aws/aws-sdk-go-v2/service/lambda"

// NewLambdaClient creates a Lambda client.
func NewLambdaClient(region string) (*lambda.Client, error) {
	s, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	opts := make([]func(*lambda.Options), 0, 1)
	if endpoint, ok := GetVirtualCloudEndpoint("lambda"); ok {
		opts = append(opts, func(o *lambda.Options) {
			o.EndpointResolver = lambda.EndpointResolverFromURL(endpoint)
		})
	}

	return lambda.NewFromConfig(*s, opts...), nil
}

// NewLambdaClientWithDefaultRegion creates a Lambda client with the default region.
func NewLambdaClientWithDefaultRegion() (*lambda.Client, error) {
	return NewLambdaClient(defaultRegion)
}
