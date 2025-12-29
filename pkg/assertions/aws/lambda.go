package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `LambdaAsserter` interface.
var _ LambdaAsserter = (*AWSAsserter)(nil)

// LambdaAsserter defines Lambda-specific assertions
type LambdaAsserter interface {
	// Basic
	AssertFunctionExists(functionName string) error
	AssertFunctionNotExists(functionName string) error

	// Configuration
	AssertFunctionRuntime(functionName, runtime string) error
	AssertFunctionHandler(functionName, handler string) error
	AssertFunctionTimeout(functionName string, timeout int) error
	AssertFunctionMemory(functionName string, memory int) error
	AssertFunctionEnvironmentVariable(functionName, key, value string) error

	// Versions & Aliases
	AssertFunctionVersionExists(functionName, version string) error
	AssertFunctionAliasExists(functionName, aliasName string) error
	AssertFunctionAliasPointsToVersion(functionName, aliasName, version string) error

	// Function URLs
	AssertFunctionURLExists(functionName string) error
	AssertFunctionURLAuthType(functionName, authType string) error

	// Layers
	AssertFunctionHasLayer(functionName, layerArn string) error

	// Event Source Mappings
	AssertEventSourceMappingExists(uuid string) error
}

// AssertFunctionExists checks if a Lambda function exists
func (a *AWSAsserter) AssertFunctionExists(functionName string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s does not exist: %w", functionName, err)
	}

	return nil
}

// AssertFunctionNotExists checks if a Lambda function does not exist
func (a *AWSAsserter) AssertFunctionNotExists(functionName string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err == nil {
		return fmt.Errorf("Lambda function %s exists but should not", functionName)
	}

	// Check if the error is a ResourceNotFoundException
	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		// If it's a different error, return it
		return fmt.Errorf("unexpected error checking Lambda function %s: %w", functionName, err)
	}

	return nil
}

// AssertFunctionRuntime checks if a Lambda function has the expected runtime
func (a *AWSAsserter) AssertFunctionRuntime(functionName, runtime string) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	actualRuntime := string(config.Runtime)
	if actualRuntime == "" {
		return fmt.Errorf("Lambda function %s has no runtime configured", functionName)
	}

	if actualRuntime != runtime {
		return fmt.Errorf("expected runtime %s, but got %s", runtime, actualRuntime)
	}

	return nil
}

// AssertFunctionHandler checks if a Lambda function has the expected handler
func (a *AWSAsserter) AssertFunctionHandler(functionName, handler string) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	if aws.ToString(config.Handler) != handler {
		return fmt.Errorf("expected handler %s, but got %s", handler, aws.ToString(config.Handler))
	}

	return nil
}

// AssertFunctionTimeout checks if a Lambda function has the expected timeout
func (a *AWSAsserter) AssertFunctionTimeout(functionName string, timeout int) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	if aws.ToInt32(config.Timeout) != int32(timeout) {
		return fmt.Errorf("expected timeout %d, but got %d", timeout, aws.ToInt32(config.Timeout))
	}

	return nil
}

// AssertFunctionMemory checks if a Lambda function has the expected memory size
func (a *AWSAsserter) AssertFunctionMemory(functionName string, memory int) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	if aws.ToInt32(config.MemorySize) != int32(memory) {
		return fmt.Errorf("expected memory %d MB, but got %d MB", memory, aws.ToInt32(config.MemorySize))
	}

	return nil
}

// AssertFunctionEnvironmentVariable checks if a Lambda function has the expected environment variable
func (a *AWSAsserter) AssertFunctionEnvironmentVariable(functionName, key, value string) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	if config.Environment == nil || config.Environment.Variables == nil {
		return fmt.Errorf("Lambda function %s has no environment variables", functionName)
	}

	actualValue, exists := config.Environment.Variables[key]
	if !exists {
		return fmt.Errorf("environment variable %s not found", key)
	}

	if actualValue != value {
		return fmt.Errorf("expected environment variable %s to have value %s, but got %s", key, value, actualValue)
	}

	return nil
}

// AssertFunctionVersionExists checks if a published version exists for the function
func (a *AWSAsserter) AssertFunctionVersionExists(functionName, version string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(version),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s version %s does not exist: %w", functionName, version, err)
	}

	return nil
}

// AssertFunctionAliasExists checks if an alias exists for the function
func (a *AWSAsserter) AssertFunctionAliasExists(functionName, aliasName string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetAlias(context.TODO(), &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s alias %s does not exist: %w", functionName, aliasName, err)
	}

	return nil
}

// AssertFunctionAliasPointsToVersion checks if an alias points to the expected version
func (a *AWSAsserter) AssertFunctionAliasPointsToVersion(functionName, aliasName, version string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	alias, err := client.GetAlias(context.TODO(), &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s alias %s does not exist: %w", functionName, aliasName, err)
	}

	if aws.ToString(alias.FunctionVersion) != version {
		return fmt.Errorf("expected alias %s to point to version %s, but got %s", aliasName, version, aws.ToString(alias.FunctionVersion))
	}

	return nil
}

// AssertFunctionURLExists checks if a function URL exists
func (a *AWSAsserter) AssertFunctionURLExists(functionName string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetFunctionUrlConfig(context.TODO(), &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s does not have a function URL: %w", functionName, err)
	}

	return nil
}

// AssertFunctionURLAuthType checks if a function URL has the expected auth type
func (a *AWSAsserter) AssertFunctionURLAuthType(functionName, authType string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	urlConfig, err := client.GetFunctionUrlConfig(context.TODO(), &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("Lambda function %s does not have a function URL: %w", functionName, err)
	}

	if string(urlConfig.AuthType) != authType {
		return fmt.Errorf("expected auth type %s, but got %s", authType, urlConfig.AuthType)
	}

	return nil
}

// AssertFunctionHasLayer checks if a function has the specified layer attached
func (a *AWSAsserter) AssertFunctionHasLayer(functionName, layerArn string) error {
	config, err := a.getFunctionConfiguration(functionName)
	if err != nil {
		return err
	}

	for _, layer := range config.Layers {
		if aws.ToString(layer.Arn) == layerArn {
			return nil
		}
	}

	return fmt.Errorf("Lambda function %s does not have layer %s", functionName, layerArn)
}

// AssertEventSourceMappingExists checks if an event source mapping exists
func (a *AWSAsserter) AssertEventSourceMappingExists(uuid string) error {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return err
	}

	_, err = client.GetEventSourceMapping(context.TODO(), &lambda.GetEventSourceMappingInput{
		UUID: aws.String(uuid),
	})
	if err != nil {
		return fmt.Errorf("event source mapping %s does not exist: %w", uuid, err)
	}

	return nil
}

// Helper method to get function configuration
func (a *AWSAsserter) getFunctionConfiguration(functionName string) (*lambda.GetFunctionConfigurationOutput, error) {
	client, err := awshelpers.NewLambdaClientWithDefaultRegion()
	if err != nil {
		return nil, err
	}

	config, err := client.GetFunctionConfiguration(context.TODO(), &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting Lambda function configuration for %s: %w", functionName, err)
	}

	return config, nil
}
