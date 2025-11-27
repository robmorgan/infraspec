package terraform

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/awshelpers"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// RegisterSteps registers all Terraform-specific step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I run [Tt]erraform apply$`, newTerraformApplyStep)
	sc.Step(`^the Terraform module at "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I have a Terraform configuration in "([^"]*)"$`, newTerraformConfigStep)
	sc.Step(`^I set the variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep)
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)"$`, newTerraformSetVariableStep) // Alternative pattern without "the"
	sc.Step(`^I set the variable "([^"]*)" to "([^"]*)" with a random suffix$`, newTerraformSetVariableWithRandomSuffixStep)
	sc.Step(`^I set variable "([^"]*)" to "([^"]*)" with a random suffix$`, newTerraformSetVariableWithRandomSuffixStep) // Alternative pattern without "the"
	sc.Step(`^I set the variable "([^"]*)" to$`, newTerraformSetMapVariableStep)
	sc.Step(`^I set the variable "([^"]*)" to a random stable AWS region$`, newTerraformSetRandomStableAWSRegion)
	sc.Step(`^the "([^"]*)" output is "([^"]*)"$`, newTerraformOutputEqualsStep)
	sc.Step(`^the output "([^"]*)" should equal "([^"]*)"$`, newTerraformOutputEqualsStep)
	sc.Step(`^the output "([^"]*)" should contain "([^"]*)"$`, newTerraformOutputContainsStep)
}

func newTerraformConfigStep(ctx context.Context, path string) (context.Context, error) {
	// construct an absolute path to the terraform configuration. If the path is relative, we need to prepend the base
	// uri of the scenario otherwise, we can just use the path as is.
	var (
		absPath string
		err     error
	)
	if !filepath.IsAbs(path) {
		base := filepath.Dir(contexthelpers.GetUri(ctx))
		absPath, err = filepath.Abs(filepath.Join(base, path))
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
	} else {
		absPath = path
	}

	options, err := iacprovisioner.WithDefaultRetryableErrors(&iacprovisioner.Options{
		WorkingDir: absPath,
		Vars:       make(map[string]interface{}),
		EnvVars:    make(map[string]string),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform options: %w", err)
	}

	// Set AWS endpoint environment variables and generate provider file when virtual cloud is enabled
	if err := configureVirtualCloudEndpoints(options, absPath); err != nil {
		return nil, fmt.Errorf("failed to configure virtual cloud endpoints: %w", err)
	}

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformApplyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	out, err := iacprovisioner.InitAndApply(options)
	if err != nil {
		// Include both the error message and output for better debugging
		if out != "" {
			return ctx, fmt.Errorf("there was an error running terraform apply: %s\n%s", err.Error(), out)
		}
		return ctx, fmt.Errorf("there was an error running terraform apply: %s", err.Error())
	}
	return contexthelpers.SetTerraformHasApplied(ctx, true), nil
}

func NewTerraformDestroyStep(ctx context.Context) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	out, err := iacprovisioner.Destroy(options)
	if err != nil {
		return ctx, fmt.Errorf("there was an error running terraform destroy: %s", out)
	}
	return contexthelpers.SetTerraformHasApplied(ctx, false), nil
}

func newTerraformSetVariableStep(ctx context.Context, name, value string) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	options.Vars[name] = value
	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformSetMapVariableStep(ctx context.Context, name string, table *godog.Table) (context.Context, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)

	// convert the table to a map[string]string
	varMap := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		varMap[row.Cells[0].Value] = row.Cells[1].Value
	}

	options.Vars[name] = varMap

	return context.WithValue(ctx, contexthelpers.TFOptionsCtxKey{}, options), nil
}

func newTerraformSetRandomStableAWSRegion(ctx context.Context, name string) (context.Context, error) {
	awsRegion, err := awshelpers.GetRandomStableRegion(nil, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to get random stable AWS region: %w", err)
	}
	ctx = contexthelpers.SetAwsRegion(ctx, awsRegion)
	return newTerraformSetVariableStep(ctx, name, awsRegion)
}

func newTerraformSetVariableWithRandomSuffixStep(ctx context.Context, name, value string) (context.Context, error) {
	randomSuffix := uniqueId()
	valueWithSuffix := fmt.Sprintf("%s-%s", value, randomSuffix)
	return newTerraformSetVariableStep(ctx, name, valueWithSuffix)
}

const base62chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base36chars = "0123456789abcdefghijklmnopqrstuvwxyz" // Lowercase only for RDS/S3 compatibility
const uniqueIDLength = 6                                    // Should be good for 62^6 = 56+ billion combinations

// uniqueId returns a unique (ish) id we can attach to resources so they don't conflict with each other.
// Uses base 36 (lowercase alphanumeric) to generate a 6 character string that's unlikely to collide with the handful
// of tests we run in parallel. Uses lowercase only to ensure compatibility with AWS resources that have strict
// naming requirements (e.g., RDS identifiers, S3 bucket names).
// Based on code here: http://stackoverflow.com/a/9543797/483528
func uniqueId() string {
	var out bytes.Buffer

	generator := newRand()
	for i := 0; i < uniqueIDLength; i++ {
		out.WriteByte(base36chars[generator.Intn(len(base36chars))])
	}

	return out.String()
}

// newRand creates a new random number generator, seeding it with the current system time.
func newRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func newTerraformOutputEqualsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	actualValue, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s, got %s: %w", outputName, actualValue, err)
	}

	if actualValue != expectedValue {
		return fmt.Errorf("expected output %s to be %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}

func newTerraformOutputContainsStep(ctx context.Context, outputName, expectedValue string) error {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	actualValue, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return fmt.Errorf("failed to get output %s, got %s: %w", outputName, actualValue, err)
	}

	// check if the expected value is a substring of the actual value
	if !strings.Contains(actualValue, expectedValue) {
		return fmt.Errorf("expected output %s to contain %s, got %s", outputName, expectedValue, actualValue)
	}
	return nil
}

// configureVirtualCloudEndpoints sets AWS endpoint environment variables when InfraSpec Virtual Cloud is enabled. This
// configures Terraform/OpenTofu to use the InfraSpec Cloud API endpoints instead of real AWS.
//
// The function sets service-specific AWS_ENDPOINT_URL_* environment variables that are
// automatically recognized by the AWS provider. Each service gets its own subdomain endpoint
// (e.g., dynamodb.infraspec.sh, sts.infraspec.sh) for proper AWS SigV4 authentication.
// See: https://search.opentofu.org/provider/opentofu/aws/v6.1.0/docs/guides/custom-service-endpoints
func configureVirtualCloudEndpoints(options *iacprovisioner.Options, workingDir string) error {
	if !config.UseInfraspecVirtualCloud() {
		return nil
	}

	config.Logging.Logger.Infof("Configuring virtual cloud endpoints for Terraform/OpenTofu")

	// Map of AWS SDK service identifiers to infraspec subdomain names
	// The AWS SDK uses specific service identifiers (e.g., APPLICATION_AUTO_SCALING)
	// while infraspec uses simplified subdomains (e.g., autoscaling.infraspec.sh)
	serviceMap := map[string]string{
		"DYNAMODB":                 "dynamodb",
		"STS":                      "sts",
		"RDS":                      "rds",
		"S3":                       "s3",
		"S3CONTROL":                "s3",
		"EC2":                      "ec2",
		"SSM":                      "ssm",
		"APPLICATION_AUTO_SCALING": "autoscaling",
	}

	// Set service-specific endpoint environment variables
	for envVarSuffix, subdomain := range serviceMap {
		envVar := fmt.Sprintf("AWS_ENDPOINT_URL_%s", envVarSuffix)

		// Check if a service-specific endpoint is already set in the environment
		if existingServiceEndpoint := os.Getenv(envVar); existingServiceEndpoint != "" {
			config.Logging.Logger.Debugf("%s already set to: %s", envVar, existingServiceEndpoint)
			options.EnvVars[envVar] = existingServiceEndpoint
			continue
		}

		// Get service-specific endpoint URL (automatically builds subdomain endpoint)
		serviceEndpoint, ok := awshelpers.GetVirtualCloudEndpoint(subdomain)
		if !ok {
			continue
		}
		options.EnvVars[envVar] = serviceEndpoint
		config.Logging.Logger.Debugf("Setting %s=%s", envVar, serviceEndpoint)
	}

	// Set credentials for InfraSpec Cloud authentication
	// The access key and secret key are used by the AWS SDK to sign requests
	options.EnvVars["AWS_ACCESS_KEY_ID"] = awshelpers.InfraspecCloudAccessKeyID

	// Get the InfraSpec Cloud token and set it as the secret access key
	if config.UseInfraspecVirtualCloud() {
		token, err := config.GetInfraspecCloudToken()
		if err != nil {
			return fmt.Errorf("failed to get InfraSpec Cloud token: %w", err)
		}
		if token == "" {
			return fmt.Errorf("virtual cloud is enabled but no token provided")
		}
		options.EnvVars["AWS_SECRET_ACCESS_KEY"] = token
	}

	return nil
}
