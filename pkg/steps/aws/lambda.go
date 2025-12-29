package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// Lambda Step Definitions
func registerLambdaSteps(sc *godog.ScenarioContext) {
	// Basic existence - direct name
	sc.Step(`^the Lambda function "([^"]*)" should exist$`, newLambdaFunctionExistsStep)
	sc.Step(`^the Lambda function "([^"]*)" should not exist$`, newLambdaFunctionNotExistsStep)

	// Basic existence - from output
	sc.Step(`^the Lambda function from output "([^"]*)" should exist$`, newLambdaFunctionFromOutputExistsStep)
	sc.Step(`^the Lambda function from output "([^"]*)" should not exist$`, newLambdaFunctionFromOutputNotExistsStep)

	// Configuration - direct name
	sc.Step(`^the Lambda function "([^"]*)" runtime should be "([^"]*)"$`, newLambdaFunctionRuntimeStep)
	sc.Step(`^the Lambda function "([^"]*)" handler should be "([^"]*)"$`, newLambdaFunctionHandlerStep)
	sc.Step(`^the Lambda function "([^"]*)" timeout should be (\d+) seconds$`, newLambdaFunctionTimeoutStep)
	sc.Step(`^the Lambda function "([^"]*)" memory should be (\d+) MB$`, newLambdaFunctionMemoryStep)
	sc.Step(`^the Lambda function "([^"]*)" should have environment variable "([^"]*)" with value "([^"]*)"$`, newLambdaFunctionEnvVarStep)

	// Configuration - from output
	sc.Step(`^the Lambda function from output "([^"]*)" runtime should be "([^"]*)"$`, newLambdaFunctionFromOutputRuntimeStep)
	sc.Step(`^the Lambda function from output "([^"]*)" handler should be "([^"]*)"$`, newLambdaFunctionFromOutputHandlerStep)
	sc.Step(`^the Lambda function from output "([^"]*)" timeout should be (\d+) seconds$`, newLambdaFunctionFromOutputTimeoutStep)
	sc.Step(`^the Lambda function from output "([^"]*)" memory should be (\d+) MB$`, newLambdaFunctionFromOutputMemoryStep)
	sc.Step(`^the Lambda function from output "([^"]*)" should have environment variable "([^"]*)" with value "([^"]*)"$`, newLambdaFunctionFromOutputEnvVarStep)

	// Versions & Aliases - direct name
	sc.Step(`^the Lambda function "([^"]*)" version "([^"]*)" should exist$`, newLambdaFunctionVersionExistsStep)
	sc.Step(`^the Lambda function "([^"]*)" alias "([^"]*)" should exist$`, newLambdaFunctionAliasExistsStep)
	sc.Step(`^the Lambda function "([^"]*)" alias "([^"]*)" should point to version "([^"]*)"$`, newLambdaFunctionAliasVersionStep)

	// Versions & Aliases - from output
	sc.Step(`^the Lambda function from output "([^"]*)" alias "([^"]*)" should exist$`, newLambdaFunctionFromOutputAliasExistsStep)
	sc.Step(`^the Lambda function from output "([^"]*)" alias "([^"]*)" should point to version "([^"]*)"$`, newLambdaFunctionFromOutputAliasVersionStep)

	// Function URLs - direct name
	sc.Step(`^the Lambda function "([^"]*)" should have a function URL$`, newLambdaFunctionURLExistsStep)
	sc.Step(`^the Lambda function "([^"]*)" function URL auth type should be "([^"]*)"$`, newLambdaFunctionURLAuthTypeStep)

	// Function URLs - from output
	sc.Step(`^the Lambda function from output "([^"]*)" should have a function URL$`, newLambdaFunctionFromOutputURLExistsStep)
	sc.Step(`^the Lambda function from output "([^"]*)" function URL auth type should be "([^"]*)"$`, newLambdaFunctionFromOutputURLAuthTypeStep)

	// Layers - direct name
	sc.Step(`^the Lambda function "([^"]*)" should have layer "([^"]*)"$`, newLambdaFunctionLayerStep)

	// Layers - from output
	sc.Step(`^the Lambda function from output "([^"]*)" should have layer "([^"]*)"$`, newLambdaFunctionFromOutputLayerStep)

	// Event Source Mappings
	sc.Step(`^the event source mapping "([^"]*)" should exist$`, newEventSourceMappingExistsStep)
	sc.Step(`^the event source mapping from output "([^"]*)" should exist$`, newEventSourceMappingFromOutputExistsStep)
}

// Helper function to get Lambda asserter
func getLambdaAsserter(ctx context.Context) (aws.LambdaAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	lambdaAssert, ok := asserter.(aws.LambdaAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement LambdaAsserter")
	}
	return lambdaAssert, nil
}

// Helper function to get function name from Terraform output
func getFunctionNameFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	functionName, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get function name from output %s: %w", outputName, err)
	}
	return functionName, nil
}

// Basic existence steps

func newLambdaFunctionExistsStep(ctx context.Context, functionName string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionExists(functionName)
}

func newLambdaFunctionNotExistsStep(ctx context.Context, functionName string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionNotExists(functionName)
}

func newLambdaFunctionFromOutputExistsStep(ctx context.Context, outputName string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionExistsStep(ctx, functionName)
}

func newLambdaFunctionFromOutputNotExistsStep(ctx context.Context, outputName string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionNotExistsStep(ctx, functionName)
}

// Configuration steps

func newLambdaFunctionRuntimeStep(ctx context.Context, functionName, runtime string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionRuntime(functionName, runtime)
}

func newLambdaFunctionHandlerStep(ctx context.Context, functionName, handler string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionHandler(functionName, handler)
}

func newLambdaFunctionTimeoutStep(ctx context.Context, functionName string, timeout int) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionTimeout(functionName, timeout)
}

func newLambdaFunctionMemoryStep(ctx context.Context, functionName string, memory int) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionMemory(functionName, memory)
}

func newLambdaFunctionEnvVarStep(ctx context.Context, functionName, key, value string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionEnvironmentVariable(functionName, key, value)
}

func newLambdaFunctionFromOutputRuntimeStep(ctx context.Context, outputName, runtime string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionRuntimeStep(ctx, functionName, runtime)
}

func newLambdaFunctionFromOutputHandlerStep(ctx context.Context, outputName, handler string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionHandlerStep(ctx, functionName, handler)
}

func newLambdaFunctionFromOutputTimeoutStep(ctx context.Context, outputName string, timeout int) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionTimeoutStep(ctx, functionName, timeout)
}

func newLambdaFunctionFromOutputMemoryStep(ctx context.Context, outputName string, memory int) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionMemoryStep(ctx, functionName, memory)
}

func newLambdaFunctionFromOutputEnvVarStep(ctx context.Context, outputName, key, value string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionEnvVarStep(ctx, functionName, key, value)
}

// Versions & Aliases steps

func newLambdaFunctionVersionExistsStep(ctx context.Context, functionName, version string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionVersionExists(functionName, version)
}

func newLambdaFunctionAliasExistsStep(ctx context.Context, functionName, aliasName string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionAliasExists(functionName, aliasName)
}

func newLambdaFunctionAliasVersionStep(ctx context.Context, functionName, aliasName, version string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionAliasPointsToVersion(functionName, aliasName, version)
}

func newLambdaFunctionFromOutputAliasExistsStep(ctx context.Context, outputName, aliasName string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionAliasExistsStep(ctx, functionName, aliasName)
}

func newLambdaFunctionFromOutputAliasVersionStep(ctx context.Context, outputName, aliasName, version string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionAliasVersionStep(ctx, functionName, aliasName, version)
}

// Function URL steps

func newLambdaFunctionURLExistsStep(ctx context.Context, functionName string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionURLExists(functionName)
}

func newLambdaFunctionURLAuthTypeStep(ctx context.Context, functionName, authType string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionURLAuthType(functionName, authType)
}

func newLambdaFunctionFromOutputURLExistsStep(ctx context.Context, outputName string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionURLExistsStep(ctx, functionName)
}

func newLambdaFunctionFromOutputURLAuthTypeStep(ctx context.Context, outputName, authType string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionURLAuthTypeStep(ctx, functionName, authType)
}

// Layer steps

func newLambdaFunctionLayerStep(ctx context.Context, functionName, layerArn string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertFunctionHasLayer(functionName, layerArn)
}

func newLambdaFunctionFromOutputLayerStep(ctx context.Context, outputName, layerArn string) error {
	functionName, err := getFunctionNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newLambdaFunctionLayerStep(ctx, functionName, layerArn)
}

// Event Source Mapping steps

func newEventSourceMappingExistsStep(ctx context.Context, uuid string) error {
	lambdaAssert, err := getLambdaAsserter(ctx)
	if err != nil {
		return err
	}
	return lambdaAssert.AssertEventSourceMappingExists(uuid)
}

func newEventSourceMappingFromOutputExistsStep(ctx context.Context, outputName string) error {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	uuid, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get event source mapping UUID from output %s: %w", outputName, err)
	}
	return newEventSourceMappingExistsStep(ctx, uuid)
}

// Silence unused import warning for strconv
var _ = strconv.Itoa
