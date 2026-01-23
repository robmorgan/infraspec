package iam

import (
	"context"
	"fmt"
	"log"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// ============================================================================
// User Policy Attachment Operations
// ============================================================================

func (s *IAMService) attachUserPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "ValidationError", "PolicyArn is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Verify policy exists
	policyName := extractPolicyNameFromArn(policyArn)
	policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
	var policy XMLPolicy
	if err := s.state.Get(policyKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s does not exist.", policyArn)), nil
	}

	// Add to user attachments
	attachKey := fmt.Sprintf("iam:user-policies:%s", userName)
	var attachments UserAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = UserAttachments{PolicyArns: []string{}}
	}

	// Check if already attached
	for _, arn := range attachments.PolicyArns {
		if arn == policyArn {
			// Already attached, idempotent success
			return s.successResponse("AttachUserPolicy", EmptyResult{})
		}
	}

	attachments.PolicyArns = append(attachments.PolicyArns, policyArn)
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to attach policy"), nil
	}

	// Add relationship in graph: policy -> user
	if err := s.addRelationship("policy", policyName, "user", userName, graph.RelAssociatedWith); err != nil {
		if s.isStrictMode() {
			attachments.PolicyArns = attachments.PolicyArns[:len(attachments.PolicyArns)-1]
			s.state.Set(attachKey, &attachments)
			return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create user-policy relationship: %v", err)), nil
		}
		log.Printf("Warning: failed to add user-policy relationship in graph: %v", err)
	}

	// Increment attachment count on policy
	var policyToUpdate XMLPolicy
	if err := s.state.Update(policyKey, &policyToUpdate, func() error {
		policyToUpdate.AttachmentCount++
		return nil
	}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update policy attachment count"), nil
	}

	return s.successResponse("AttachUserPolicy", EmptyResult{})
}

func (s *IAMService) detachUserPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "ValidationError", "PolicyArn is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get user attachments
	attachKey := fmt.Sprintf("iam:user-policies:%s", userName)
	var attachments UserAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to user %s.", policyArn, userName)), nil
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
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to user %s.", policyArn, userName)), nil
	}

	attachments.PolicyArns = newArns
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to detach policy"), nil
	}

	// Remove relationship in graph
	policyName := extractPolicyNameFromArn(policyArn)
	if err := s.removeRelationship("policy", policyName, "user", userName, graph.RelAssociatedWith); err != nil {
		log.Printf("Warning: failed to remove user-policy relationship in graph: %v", err)
	}

	// Decrement attachment count on policy
	policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
	var policyToUpdate XMLPolicy
	_ = s.state.Update(policyKey, &policyToUpdate, func() error {
		if policyToUpdate.AttachmentCount > 0 {
			policyToUpdate.AttachmentCount--
		}
		return nil
	})

	return s.successResponse("DetachUserPolicy", EmptyResult{})
}

func (s *IAMService) listAttachedUserPolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	attachKey := fmt.Sprintf("iam:user-policies:%s", userName)
	var attachments UserAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = UserAttachments{PolicyArns: []string{}}
	}

	var policies []XMLAttachedPolicy
	for _, arn := range attachments.PolicyArns {
		policyName := extractPolicyNameFromArn(arn)
		policies = append(policies, XMLAttachedPolicy{
			PolicyName: policyName,
			PolicyArn:  arn,
		})
	}

	result := ListAttachedUserPoliciesResult{
		AttachedPolicies: policies,
		IsTruncated:      false,
	}
	return s.successResponse("ListAttachedUserPolicies", result)
}

// ============================================================================
// User Inline Policy Operations
// ============================================================================

func (s *IAMService) putUserPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	policyDocument := getStringValue(params, "PolicyDocument")
	if policyDocument == "" {
		return s.errorResponse(400, "ValidationError", "PolicyDocument is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get or create inline policies storage
	inlineKey := fmt.Sprintf("iam:user-inline-policies:%s", userName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		inlinePolicies = UserInlinePolicies{Policies: make(map[string]string)}
	}
	if inlinePolicies.Policies == nil {
		inlinePolicies.Policies = make(map[string]string)
	}

	// Store the inline policy
	inlinePolicies.Policies[policyName] = policyDocument
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store inline policy"), nil
	}

	return s.successResponse("PutUserPolicy", EmptyResult{})
}

func (s *IAMService) getUserPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:user-inline-policies:%s", userName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user policy with name %s cannot be found.", policyName)), nil
	}

	policyDocument, exists := inlinePolicies.Policies[policyName]
	if !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user policy with name %s cannot be found.", policyName)), nil
	}

	result := GetUserPolicyResult{
		UserName:       userName,
		PolicyName:     policyName,
		PolicyDocument: policyDocument,
	}
	return s.successResponse("GetUserPolicy", result)
}

func (s *IAMService) deleteUserPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:user-inline-policies:%s", userName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user policy with name %s cannot be found.", policyName)), nil
	}

	if _, exists := inlinePolicies.Policies[policyName]; !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user policy with name %s cannot be found.", policyName)), nil
	}

	// Delete the inline policy
	delete(inlinePolicies.Policies, policyName)
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete inline policy"), nil
	}

	return s.successResponse("DeleteUserPolicy", EmptyResult{})
}

func (s *IAMService) listUserPolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:user-inline-policies:%s", userName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		// No inline policies is not an error, just return empty list
		inlinePolicies = UserInlinePolicies{Policies: make(map[string]string)}
	}

	var policyNames []string
	for name := range inlinePolicies.Policies {
		policyNames = append(policyNames, name)
	}

	result := ListUserPoliciesResult{
		PolicyNames: policyNames,
		IsTruncated: false,
	}
	return s.successResponse("ListUserPolicies", result)
}
