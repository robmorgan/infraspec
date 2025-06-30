package awshelpers

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	AuthAssumeRoleEnvVar = "INFRASPEC_IAM_ROLE" // OS environment variable name through which Assume Role ARN may be passed for authentication
)

// NewAuthenticatedSession creates an AWS Config following to standard AWS authentication workflow.
// If AuthAssumeIamRoleEnvVar environment variable is set, assumes IAM role specified in it.
func NewAuthenticatedSession(region string) (*aws.Config, error) {
	if assumeRoleArn, ok := os.LookupEnv(AuthAssumeRoleEnvVar); ok {
		return NewAuthenticatedSessionFromRole(region, assumeRoleArn)
	} else {
		return NewAuthenticatedSessionFromDefaultCredentials(region)
	}
}

// NewAuthenticatedSessionFromDefaultCredentials gets an AWS Config, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSessionFromDefaultCredentials(region string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, CredentialsError{UnderlyingErr: err}
	}

	return &cfg, nil
}

// NewAuthenticatedSessionFromRole returns a new AWS Config after assuming the
// role whose ARN is provided in roleARN. If the credentials are not properly
// configured in the underlying environment, an error is returned.
func NewAuthenticatedSessionFromRole(region string, roleARN string) (*aws.Config, error) {
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
