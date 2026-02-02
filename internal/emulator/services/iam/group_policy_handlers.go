package iam

import (
	"context"
	"fmt"
	"log"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// ============================================================================
// Group Policy Attachment Operations
// ============================================================================

func (s *IAMService) attachGroupPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "ValidationError", "PolicyArn is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Verify policy exists
	policyName := extractPolicyNameFromArn(policyArn)
	policyKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
	var policy XMLPolicy
	if err := s.state.Get(policyKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s does not exist.", policyArn)), nil
	}

	// Add to group attachments
	attachKey := fmt.Sprintf("iam:group-policies:%s", groupName)
	var attachments GroupAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = GroupAttachments{PolicyArns: []string{}}
	}

	// Check if already attached
	for _, arn := range attachments.PolicyArns {
		if arn == policyArn {
			// Already attached, idempotent success
			return s.successResponse("AttachGroupPolicy", EmptyResult{})
		}
	}

	attachments.PolicyArns = append(attachments.PolicyArns, policyArn)
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to attach policy"), nil
	}

	// Add relationship in graph: policy -> group
	if err := s.addRelationship("policy", policyName, "group", groupName, graph.RelAssociatedWith); err != nil {
		if s.isStrictMode() {
			attachments.PolicyArns = attachments.PolicyArns[:len(attachments.PolicyArns)-1]
			s.state.Set(attachKey, &attachments)
			return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create group-policy relationship: %v", err)), nil
		}
		log.Printf("Warning: failed to add group-policy relationship in graph: %v", err)
	}

	// Increment attachment count on policy
	var policyToUpdate XMLPolicy
	if err := s.state.Update(policyKey, &policyToUpdate, func() error {
		policyToUpdate.AttachmentCount++
		return nil
	}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update policy attachment count"), nil
	}

	return s.successResponse("AttachGroupPolicy", EmptyResult{})
}

func (s *IAMService) detachGroupPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "ValidationError", "PolicyArn is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get group attachments
	attachKey := fmt.Sprintf("iam:group-policies:%s", groupName)
	var attachments GroupAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to group %s.", policyArn, groupName)), nil
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
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s is not attached to group %s.", policyArn, groupName)), nil
	}

	attachments.PolicyArns = newArns
	if err := s.state.Set(attachKey, &attachments); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to detach policy"), nil
	}

	// Remove relationship in graph
	policyName := extractPolicyNameFromArn(policyArn)
	if err := s.removeRelationship("policy", policyName, "group", groupName, graph.RelAssociatedWith); err != nil {
		log.Printf("Warning: failed to remove group-policy relationship in graph: %v", err)
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

	return s.successResponse("DetachGroupPolicy", EmptyResult{})
}

func (s *IAMService) listAttachedGroupPolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	attachKey := fmt.Sprintf("iam:group-policies:%s", groupName)
	var attachments GroupAttachments
	if err := s.state.Get(attachKey, &attachments); err != nil {
		attachments = GroupAttachments{PolicyArns: []string{}}
	}

	var policies []XMLAttachedPolicy
	for _, arn := range attachments.PolicyArns {
		policyName := extractPolicyNameFromArn(arn)
		policies = append(policies, XMLAttachedPolicy{
			PolicyName: policyName,
			PolicyArn:  arn,
		})
	}

	result := ListAttachedGroupPoliciesResult{
		AttachedPolicies: policies,
		IsTruncated:      false,
	}
	return s.successResponse("ListAttachedGroupPolicies", result)
}

// ============================================================================
// Group Inline Policy Operations
// ============================================================================

func (s *IAMService) putGroupPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	policyDocument := getStringValue(params, "PolicyDocument")
	if policyDocument == "" {
		return s.errorResponse(400, "ValidationError", "PolicyDocument is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get or create inline policies storage
	inlineKey := fmt.Sprintf("iam:group-inline-policies:%s", groupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		inlinePolicies = GroupInlinePolicies{Policies: make(map[string]string)}
	}
	if inlinePolicies.Policies == nil {
		inlinePolicies.Policies = make(map[string]string)
	}

	// Store the inline policy
	inlinePolicies.Policies[policyName] = policyDocument
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store inline policy"), nil
	}

	return s.successResponse("PutGroupPolicy", EmptyResult{})
}

func (s *IAMService) getGroupPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:group-inline-policies:%s", groupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group policy with name %s cannot be found.", policyName)), nil
	}

	policyDocument, exists := inlinePolicies.Policies[policyName]
	if !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group policy with name %s cannot be found.", policyName)), nil
	}

	result := GetGroupPolicyResult{
		GroupName:      groupName,
		PolicyName:     policyName,
		PolicyDocument: policyDocument,
	}
	return s.successResponse("GetGroupPolicy", result)
}

func (s *IAMService) deleteGroupPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "ValidationError", "PolicyName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:group-inline-policies:%s", groupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group policy with name %s cannot be found.", policyName)), nil
	}

	if _, exists := inlinePolicies.Policies[policyName]; !exists {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group policy with name %s cannot be found.", policyName)), nil
	}

	// Delete the inline policy
	delete(inlinePolicies.Policies, policyName)
	if err := s.state.Set(inlineKey, &inlinePolicies); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete inline policy"), nil
	}

	return s.successResponse("DeleteGroupPolicy", EmptyResult{})
}

func (s *IAMService) listGroupPolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get inline policies
	inlineKey := fmt.Sprintf("iam:group-inline-policies:%s", groupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err != nil {
		// No inline policies is not an error, just return empty list
		inlinePolicies = GroupInlinePolicies{Policies: make(map[string]string)}
	}

	var policyNames []string
	for name := range inlinePolicies.Policies {
		policyNames = append(policyNames, name)
	}

	result := ListGroupPoliciesResult{
		PolicyNames: policyNames,
		IsTruncated: false,
	}
	return s.successResponse("ListGroupPolicies", result)
}
