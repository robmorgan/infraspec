package awshelpers

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/robmorgan/infraspec/internal/config"
)

const (
	AuthAssumeRoleEnvVar = "INFRASPEC_IAM_ROLE" // OS environment variable name through which Assume Role ARN may be passed for authentication
	// InfraspecCloudAccessKeyID is the access key ID used when authenticating with an InfraSpec Cloud token
	InfraspecCloudAccessKeyID        = "infraspec-test"
	InfraspecCloudDefaultEndpointURL = "http://api.infraspec.sh:8000"
)

// NewAuthenticatedSession creates an AWS Config following to standard AWS authentication workflow.
// If an InfraSpec Cloud token is configured, it uses that token as the secret access key with "infraspec-test" as the access key ID.
// If `INFRASPEC_IAM_ROLE` environment variable is set, it assumes IAM role specified in it.
// Otherwise, uses default credentials.
func NewAuthenticatedSession(region string) (*aws.Config, error) {
	if config.UseInfraspecVirtualCloud() {
		config.Logging.Logger.Info("Using InfraSpec Virtual Cloud")

		cloudToken, err := config.GetInfraspecCloudToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get InfraSpec Cloud token: %w", err)
		}

		if cloudToken != "" {
			return NewAuthenticatedSessionFromInfraspecCloudToken(region, cloudToken)
		}
	}

	// Fall back to existing behavior
	if assumeRoleArn, ok := os.LookupEnv(AuthAssumeRoleEnvVar); ok {
		return NewAuthenticatedSessionFromRole(region, assumeRoleArn)
	}

	return NewAuthenticatedSessionFromDefaultCredentials(region)
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

// NewAuthenticatedSessionFromInfraspecCloudToken creates an AWS Config using the InfraSpec Cloud token as the secret access key.

func NewAuthenticatedSessionFromInfraspecCloudToken(region, token string) (*aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     InfraspecCloudAccessKeyID,
				SecretAccessKey: token,
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
