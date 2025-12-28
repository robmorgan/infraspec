package awshelpers

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	AuthAssumeRoleEnvVar = "INFRASPEC_IAM_ROLE" // OS environment variable name through which Assume Role ARN may be passed for authentication
)

// NewAuthenticatedSession creates an AWS Config following to standard AWS authentication workflow.
// If AWS_ENDPOINT_URL points to localhost (embedded emulator mode), uses dummy credentials.
// If `INFRASPEC_IAM_ROLE` environment variable is set, it assumes IAM role specified in it.
// Otherwise, uses default credentials.
func NewAuthenticatedSession(region string) (*aws.Config, error) {
	// If endpoint is localhost (embedded emulator), use dummy credentials
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); isLocalhost(endpoint) {
		return NewAuthenticatedSessionWithCredentials(region, "test", "test")
	}

	// Fall back to existing behavior
	if assumeRoleArn, ok := os.LookupEnv(AuthAssumeRoleEnvVar); ok {
		return NewAuthenticatedSessionFromRole(region, assumeRoleArn)
	}

	return NewAuthenticatedSessionFromDefaultCredentials(region)
}

// isLocalhost checks if the given endpoint URL points to localhost.
func isLocalhost(endpoint string) bool {
	if endpoint == "" {
		return false
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// NewAuthenticatedSessionWithDefaultRegion creates an AWS Config with the default region.
func NewAuthenticatedSessionWithDefaultRegion() (*aws.Config, error) {
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		region = defaultRegion
	}
	return NewAuthenticatedSession(region)
}

// NewAuthenticatedSessionWithCredentials creates an AWS Config using the provided credentials.
func NewAuthenticatedSessionWithCredentials(region, accessKeyID, secretAccessKey string) (*aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
		}),
	)
	if err != nil {
		return nil, CredentialsError{UnderlyingErr: err}
	}

	return &cfg, nil
}

// NewAuthenticatedSessionFromDefaultCredentials gets an AWS Config, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSessionFromDefaultCredentials(region string) (*aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, CredentialsError{UnderlyingErr: err}
	}

	return &cfg, nil
}

// NewAuthenticatedSessionFromRole returns a new AWS Config after assuming the
// role whose ARN is provided in roleARN. If the credentials are not properly
// configured in the underlying environment, an error is returned.
func NewAuthenticatedSessionFromRole(region, roleARN string) (*aws.Config, error) {
	cfg, err := NewAuthenticatedSessionFromDefaultCredentials(region)
	if err != nil {
		return nil, err
	}

	client := sts.NewFromConfig(*cfg)

	roleProvider := stscreds.NewAssumeRoleProvider(client, roleARN)
	retrieve, err := roleProvider.Retrieve(context.Background())
	if err != nil {
		return nil, CredentialsError{UnderlyingErr: err}
	}

	return &aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(credentials.StaticCredentialsProvider{
			Value: retrieve,
		}),
	}, nil
}

// CredentialsError is an error that occurs because AWS credentials can't be found.
type CredentialsError struct {
	UnderlyingErr error
}

func (err CredentialsError) Error() string {
	return fmt.Sprintf("Error finding AWS credentials. Did you set the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables or configure an AWS profile? Underlying error: %v", err.UnderlyingErr)
}
