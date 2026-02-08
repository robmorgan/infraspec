package iam

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *IAMService) putRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyName is required"), nil
	}

	policyDocument := getStringValue(params, "PolicyDocument")
	if policyDocument == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyDocument is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get or create inline policies storage
	inlineKey := fmt.Sprintf("iam:role-inline-policies:%s", roleName)
	var inlinePolicies RoleInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		inlinePolicies = RoleInlinePolicies{Policies: make(map[string]string)}
	}
	if inlinePolicies.Policies == nil {
		inlinePolicies.Policies = make(map[string]string)
	}

	// Store the inline policy
	inlinePolicies.Policies[policyName] = policyDocument
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store inline policy"), nil
	}

	return s.successResponse("PutRolePolicy", EmptyResult{})
}

func (s *IAMService) getRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyName is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:role-inline-policies:%s", roleName)
	var inlinePolicies RoleInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role policy with name %s cannot be found.", policyName)), nil
	}

	policyDocument, exists := inlinePolicies.Policies[policyName]
	if !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role policy with name %s cannot be found.", policyName)), nil
	}

	result := GetRolePolicyResult{
		RoleName:       roleName,
		PolicyName:     policyName,
		PolicyDocument: policyDocument,
	}
	return s.successResponse("GetRolePolicy", result)
}

func (s *IAMService) deleteRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyName is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:role-inline-policies:%s", roleName)
	var inlinePolicies RoleInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role policy with name %s cannot be found.", policyName)), nil
	}

	if _, exists := inlinePolicies.Policies[policyName]; !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role policy with name %s cannot be found.", policyName)), nil
	}

	// Delete the inline policy
	delete(inlinePolicies.Policies, policyName)
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete inline policy"), nil
	}

	return s.successResponse("DeleteRolePolicy", EmptyResult{})
}

func (s *IAMService) listRolePolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(roleKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:role-inline-policies:%s", roleName)
	var inlinePolicies RoleInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		// No inline policies is not an error, just return empty list
		inlinePolicies = RoleInlinePolicies{Policies: make(map[string]string)}
	}

	var policyNames []string
	for name := range inlinePolicies.Policies {
		policyNames = append(policyNames, name)
	}

	result := ListRolePoliciesResult{
		PolicyNames: policyNames,
		IsTruncated: false,
	}
	return s.successResponse("ListRolePolicies", result)
}
