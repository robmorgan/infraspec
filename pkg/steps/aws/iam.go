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

// registerIAMSteps registers all IAM-related Gherkin step definitions
func registerIAMSteps(sc *godog.ScenarioContext) {
	// Permission check
	sc.Step(`^I have the necessary IAM permissions to describe IAM roles$`, newVerifyIAMDescribeRolesStep)

	// Role assertions - direct
	sc.Step(`^the IAM role "([^"]*)" should exist$`, newIAMRoleExistsStep)
	sc.Step(`^the IAM role "([^"]*)" path should be "([^"]*)"$`, newIAMRolePathStep)
	sc.Step(`^the IAM role "([^"]*)" max session duration should be (\d+)$`, newIAMRoleMaxSessionDurationStep)
	sc.Step(`^the IAM role "([^"]*)" should have the tags$`, newIAMRoleTagsStep)

	// Role assertions - from Terraform output
	sc.Step(`^the IAM role from output "([^"]*)" should exist$`, newIAMRoleFromOutputExistsStep)
	sc.Step(`^the IAM role from output "([^"]*)" path should be "([^"]*)"$`, newIAMRoleFromOutputPathStep)
	sc.Step(`^the IAM role from output "([^"]*)" max session duration should be (\d+)$`, newIAMRoleFromOutputMaxSessionDurationStep)
	sc.Step(`^the IAM role from output "([^"]*)" should have the tags$`, newIAMRoleFromOutputTagsStep)

	// Policy assertions - direct
	sc.Step(`^the IAM policy "([^"]*)" should exist$`, newIAMPolicyExistsStep)
	sc.Step(`^the IAM policy "([^"]*)" should be attached to role "([^"]*)"$`, newIAMPolicyAttachedToRoleStep)

	// Policy assertions - from Terraform output
	sc.Step(`^the IAM policy from output "([^"]*)" should exist$`, newIAMPolicyFromOutputExistsStep)
	sc.Step(`^the IAM policy from output "([^"]*)" should be attached to role from output "([^"]*)"$`, newIAMPolicyAttachedToRoleFromOutputStep)

	// Instance profile assertions - direct
	sc.Step(`^the IAM instance profile "([^"]*)" should exist$`, newIAMInstanceProfileExistsStep)
	sc.Step(`^the IAM instance profile "([^"]*)" should have role "([^"]*)"$`, newIAMInstanceProfileHasRoleStep)

	// Instance profile assertions - from Terraform output
	sc.Step(`^the IAM instance profile from output "([^"]*)" should exist$`, newIAMInstanceProfileFromOutputExistsStep)
	sc.Step(`^the IAM instance profile from output "([^"]*)" should have role from output "([^"]*)"$`, newIAMInstanceProfileHasRoleFromOutputStep)
}

// Permission check step
func newVerifyIAMDescribeRolesStep(ctx context.Context) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertIAMDescribeRoles()
}

// Role steps - direct
func newIAMRoleExistsStep(ctx context.Context, roleName string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertRoleExists(roleName)
}

func newIAMRolePathStep(ctx context.Context, roleName, expectedPath string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertRolePath(roleName, expectedPath)
}

func newIAMRoleMaxSessionDurationStep(ctx context.Context, roleName string, durationStr string) error {
	duration, err := strconv.ParseInt(durationStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid max session duration %s: %w", durationStr, err)
	}

	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertRoleMaxSessionDuration(roleName, int32(duration))
}

func newIAMRoleTagsStep(ctx context.Context, roleName string, table *godog.Table) error {
	expectedTags, err := parseTagsTable(table)
	if err != nil {
		return err
	}

	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertRoleTags(roleName, expectedTags)
}

// Role steps - from Terraform output
func newIAMRoleFromOutputExistsStep(ctx context.Context, outputName string) error {
	roleName, err := getRoleNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMRoleExistsStep(ctx, roleName)
}

func newIAMRoleFromOutputPathStep(ctx context.Context, outputName, expectedPath string) error {
	roleName, err := getRoleNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMRolePathStep(ctx, roleName, expectedPath)
}

func newIAMRoleFromOutputMaxSessionDurationStep(ctx context.Context, outputName string, durationStr string) error {
	roleName, err := getRoleNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMRoleMaxSessionDurationStep(ctx, roleName, durationStr)
}

func newIAMRoleFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	roleName, err := getRoleNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMRoleTagsStep(ctx, roleName, table)
}

// Policy steps - direct
func newIAMPolicyExistsStep(ctx context.Context, policyArn string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertPolicyExists(policyArn)
}

func newIAMPolicyAttachedToRoleStep(ctx context.Context, policyArn, roleName string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertPolicyAttachedToRole(roleName, policyArn)
}

// Policy steps - from Terraform output
func newIAMPolicyFromOutputExistsStep(ctx context.Context, outputName string) error {
	policyArn, err := getPolicyArnFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMPolicyExistsStep(ctx, policyArn)
}

func newIAMPolicyAttachedToRoleFromOutputStep(ctx context.Context, policyOutputName, roleOutputName string) error {
	policyArn, err := getPolicyArnFromOutput(ctx, policyOutputName)
	if err != nil {
		return err
	}
	roleName, err := getRoleNameFromOutput(ctx, roleOutputName)
	if err != nil {
		return err
	}
	return newIAMPolicyAttachedToRoleStep(ctx, policyArn, roleName)
}

// Instance profile steps - direct
func newIAMInstanceProfileExistsStep(ctx context.Context, instanceProfileName string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertInstanceProfileExists(instanceProfileName)
}

func newIAMInstanceProfileHasRoleStep(ctx context.Context, instanceProfileName, roleName string) error {
	iamAssert, err := getIAMAsserter(ctx)
	if err != nil {
		return err
	}
	return iamAssert.AssertInstanceProfileHasRole(instanceProfileName, roleName)
}

// Instance profile steps - from Terraform output
func newIAMInstanceProfileFromOutputExistsStep(ctx context.Context, outputName string) error {
	instanceProfileName, err := getInstanceProfileNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newIAMInstanceProfileExistsStep(ctx, instanceProfileName)
}

func newIAMInstanceProfileHasRoleFromOutputStep(ctx context.Context, profileOutputName, roleOutputName string) error {
	instanceProfileName, err := getInstanceProfileNameFromOutput(ctx, profileOutputName)
	if err != nil {
		return err
	}
	roleName, err := getRoleNameFromOutput(ctx, roleOutputName)
	if err != nil {
		return err
	}
	return newIAMInstanceProfileHasRoleStep(ctx, instanceProfileName, roleName)
}

// Helper functions

func getIAMAsserter(ctx context.Context) (aws.IAMAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	iamAssert, ok := asserter.(aws.IAMAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement IAMAsserter")
	}
	return iamAssert, nil
}

func getRoleNameFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	roleName, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get role name from output %s: %w", outputName, err)
	}
	return roleName, nil
}

func getPolicyArnFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	policyArn, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get policy ARN from output %s: %w", outputName, err)
	}
	return policyArn, nil
}

func getInstanceProfileNameFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	instanceProfileName, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get instance profile name from output %s: %w", outputName, err)
	}
	return instanceProfileName, nil
}

func parseTagsTable(table *godog.Table) (map[string]string, error) {
	if len(table.Rows) < 2 {
		return nil, fmt.Errorf("tags table must have at least a header and one data row")
	}

	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		if len(row.Cells) < 2 {
			return nil, fmt.Errorf("each tag row must have Key and Value columns")
		}
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}
	return tags, nil
}
