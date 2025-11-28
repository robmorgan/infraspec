package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `IAMAsserter` interface.
var _ IAMAsserter = (*AWSAsserter)(nil)

// IAMAsserter defines IAM-specific assertions
type IAMAsserter interface {
	AssertIAMDescribeRoles() error
	AssertRoleExists(roleName string) error
	AssertRolePath(roleName, expectedPath string) error
	AssertRoleMaxSessionDuration(roleName string, expectedDuration int32) error
	AssertRoleTags(roleName string, expectedTags map[string]string) error
	AssertPolicyExists(policyArn string) error
	AssertPolicyAttachedToRole(roleName, policyArn string) error
	AssertInstanceProfileExists(instanceProfileName string) error
	AssertInstanceProfileHasRole(instanceProfileName, roleName string) error
}

// AssertIAMDescribeRoles checks if the AWS account has permission to describe IAM roles
func (a *AWSAsserter) AssertIAMDescribeRoles() error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	// List roles to verify access
	_, err = client.ListRoles(context.TODO(), &iam.ListRolesInput{
		MaxItems: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("error listing IAM roles: %w", err)
	}

	return nil
}

// AssertRoleExists checks if an IAM role exists
func (a *AWSAsserter) AssertRoleExists(roleName string) error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	_, err = client.GetRole(context.TODO(), &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("IAM role %s does not exist or is not accessible: %w", roleName, err)
	}

	return nil
}

// AssertRolePath checks if an IAM role has the expected path
func (a *AWSAsserter) AssertRolePath(roleName, expectedPath string) error {
	role, err := a.getRole(roleName)
	if err != nil {
		return err
	}

	actualPath := aws.ToString(role.Path)
	if actualPath != expectedPath {
		return fmt.Errorf("IAM role %s has path %s, expected %s", roleName, actualPath, expectedPath)
	}

	return nil
}

// AssertRoleMaxSessionDuration checks if an IAM role has the expected max session duration
func (a *AWSAsserter) AssertRoleMaxSessionDuration(roleName string, expectedDuration int32) error {
	role, err := a.getRole(roleName)
	if err != nil {
		return err
	}

	actualDuration := aws.ToInt32(role.MaxSessionDuration)
	if actualDuration != expectedDuration {
		return fmt.Errorf("IAM role %s has max session duration %d, expected %d", roleName, actualDuration, expectedDuration)
	}

	return nil
}

// AssertRoleTags checks if an IAM role has the expected tags
func (a *AWSAsserter) AssertRoleTags(roleName string, expectedTags map[string]string) error {
	role, err := a.getRole(roleName)
	if err != nil {
		return err
	}

	// Convert role tags to map
	actualTags := make(map[string]string)
	for _, tag := range role.Tags {
		actualTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Check that all expected tags are present with correct values
	for key, expectedValue := range expectedTags {
		actualValue, exists := actualTags[key]
		if !exists {
			return fmt.Errorf("IAM role %s is missing tag %s", roleName, key)
		}
		if actualValue != expectedValue {
			return fmt.Errorf("IAM role %s has tag %s=%s, expected %s", roleName, key, actualValue, expectedValue)
		}
	}

	return nil
}

// AssertPolicyExists checks if an IAM managed policy exists
func (a *AWSAsserter) AssertPolicyExists(policyArn string) error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	_, err = client.GetPolicy(context.TODO(), &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return fmt.Errorf("IAM policy %s does not exist or is not accessible: %w", policyArn, err)
	}

	return nil
}

// AssertPolicyAttachedToRole checks if a policy is attached to a role
func (a *AWSAsserter) AssertPolicyAttachedToRole(roleName, policyArn string) error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	// List attached policies for the role
	result, err := client.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("error listing attached policies for role %s: %w", roleName, err)
	}

	// Check if the policy is in the list
	for _, policy := range result.AttachedPolicies {
		if aws.ToString(policy.PolicyArn) == policyArn {
			return nil
		}
	}

	return fmt.Errorf("policy %s is not attached to role %s", policyArn, roleName)
}

// AssertInstanceProfileExists checks if an IAM instance profile exists
func (a *AWSAsserter) AssertInstanceProfileExists(instanceProfileName string) error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	_, err = client.GetInstanceProfile(context.TODO(), &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	})
	if err != nil {
		return fmt.Errorf("IAM instance profile %s does not exist or is not accessible: %w", instanceProfileName, err)
	}

	return nil
}

// AssertInstanceProfileHasRole checks if an instance profile contains a specific role
func (a *AWSAsserter) AssertInstanceProfileHasRole(instanceProfileName, roleName string) error {
	client, err := a.createIAMClient()
	if err != nil {
		return err
	}

	result, err := client.GetInstanceProfile(context.TODO(), &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	})
	if err != nil {
		return fmt.Errorf("error getting instance profile %s: %w", instanceProfileName, err)
	}

	// Check if the role is in the instance profile
	for _, role := range result.InstanceProfile.Roles {
		if aws.ToString(role.RoleName) == roleName {
			return nil
		}
	}

	return fmt.Errorf("role %s is not in instance profile %s", roleName, instanceProfileName)
}

// getRole is a helper method to get an IAM role
func (a *AWSAsserter) getRole(roleName string) (*types.Role, error) {
	client, err := a.createIAMClient()
	if err != nil {
		return nil, err
	}

	result, err := client.GetRole(context.TODO(), &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting IAM role %s: %w", roleName, err)
	}

	return result.Role, nil
}

// createIAMClient creates an IAM client with optional virtual cloud endpoint
func (a *AWSAsserter) createIAMClient() (*iam.Client, error) {
	cfg, err := awshelpers.NewAuthenticatedSessionWithDefaultRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	opts := make([]func(*iam.Options), 0, 1)
	if endpoint, ok := awshelpers.GetVirtualCloudEndpoint("iam"); ok {
		opts = append(opts, func(o *iam.Options) {
			o.EndpointResolver = iam.EndpointResolverFromURL(endpoint)
		})
	}

	return iam.NewFromConfig(*cfg, opts...), nil
}
