package iam

import (
	"context"
	"fmt"
	"log"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

func (s *IAMService) attachRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	isAWSManaged := isAWSManagedPolicyArn(policyArn)

	// Verify policy exists (skip for AWS managed policies - they always exist)
	if !isAWSManaged {
		policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
		var policy XMLPolicy
		if err := s.state.Get(policyKey, &policy); err != nil {
			return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s does not exist.", policyArn)), nil
		}
	}

	// Add to role attachments
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	var attachments RoleAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = RoleAttachments{PolicyArns: []string{}}
	}

	// Check if already attached
	for _, arn := range attachments.PolicyArns {
		if arn == policyArn {
			// Already attached, idempotent success
			return s.successResponse("AttachRolePolicy", EmptyResult{})
		}
	}

	attachments.PolicyArns = append(attachments.PolicyArns, policyArn)
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to attach policy"), nil
	}

	// Add relationship in graph: policy -> role (policy attachment depends on role)
	// This edge direction means: deleting the role will fail if policies are attached
	// Skip graph relationship for AWS managed policies (they're not in the graph)
	if !isAWSManaged {
		if err := s.addRelationship("policy", policyName, "role", roleName, graph.RelAssociatedWith); err != nil {
			if s.isStrictMode() {
				// Rollback: remove the policy from attachments
				attachments.PolicyArns = attachments.PolicyArns[:len(attachments.PolicyArns)-1]
				s.state.Set(attachKey, &attachments)
				return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create role-policy relationship: %v", err)), nil
			}
			log.Printf("Warning: failed to add role-policy relationship in graph: %v", err)
		}

		// Increment attachment count on policy atomically (only for customer-managed policies)
		policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
		var policyToUpdate XMLPolicy
		if err := s.state.Update(policyKey, &policyToUpdate, func() error {
			policyToUpdate.AttachmentCount++
			return nil
		}); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to update policy attachment count"), nil
		}
	}

	return s.successResponse("AttachRolePolicy", EmptyResult{})
}

func (s *IAMService) detachRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get role attachments
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	var attachments RoleAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to role %s.", policyArn, roleName)), nil
	}

	// Find and remove the policy
	found := false
	newArns := make([]string, 0, len(attachments.PolicyArns))
	for _, arn := range attachments.PolicyArns {
		if arn == policyArn {
			found = true
		} else {
			newArns = append(newArns, arn)
		}
	}

	if !found {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to role %s.", policyArn, roleName)), nil
	}

	attachments.PolicyArns = newArns
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to detach policy"), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	isAWSManaged := isAWSManagedPolicyArn(policyArn)

	// Skip graph and attachment count updates for AWS managed policies
	if !isAWSManaged {
		// Remove relationship in graph: policy -> role
		if err := s.removeRelationship("policy", policyName, "role", roleName, graph.RelAssociatedWith); err != nil {
			log.Printf("Warning: failed to remove role-policy relationship in graph: %v", err)
		}

		// Decrement attachment count on policy atomically
		policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
		var policyToUpdate XMLPolicy
		// Use Update for atomic decrement; ignore error if policy was deleted
		_ = s.state.Update(policyKey, &policyToUpdate, func() error {
			if policyToUpdate.AttachmentCount > 0 {
				policyToUpdate.AttachmentCount--
			}
			return nil
		})
	}

	return s.successResponse("DetachRolePolicy", EmptyResult{})
}

func (s *IAMService) listAttachedRolePolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	var attachments RoleAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = RoleAttachments{PolicyArns: []string{}}
	}

	var policies []XMLAttachedPolicy
	for _, arn := range attachments.PolicyArns {
		policyName := extractPolicyNameFromArn(arn)
		policies = append(policies, XMLAttachedPolicy{
			PolicyName: policyName,
			PolicyArn:  arn,
		})
	}

	result := ListAttachedRolePoliciesResult{
		AttachedPolicies: policies,
		IsTruncated:      false,
	}
	return s.successResponse("ListAttachedRolePolicies", result)
}
