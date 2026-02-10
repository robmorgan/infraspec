package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *IAMService) createPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	policyName := getStringValue(params, "PolicyName")
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyName is required"), nil
	}

	policyDocument := getStringValue(params, "PolicyDocument")
	if policyDocument == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyDocument is required"), nil
	}

	if !json.Valid([]byte(policyDocument)) {
		return s.errorResponse(400, "MalformedPolicyDocument", "PolicyDocument is not valid JSON"), nil
	}

	path := getStringValue(params, "Path")
	if path == "" {
		path = "/"
	}

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy%s%s", defaultAccountID, path, policyName)

	// Check if policy already exists
	stateKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("A policy called %s already exists.", policyName)), nil
	}

	now := time.Now().UTC()
	description := getStringValue(params, "Description")

	policy := XMLPolicy{
		PolicyName:       policyName,
		PolicyId:         generateIAMId("ANPA"),
		Arn:              policyArn,
		Path:             path,
		Description:      description,
		DefaultVersionId: "v1",
		CreateDate:       now,
		UpdateDate:       now,
		AttachmentCount:  0,
		IsAttachable:     true,
		Tags:             s.parseTags(params),
	}

	if err := s.state.Set(stateKey, &policy); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store policy"), nil
	}

	// Store the policy version
	version := XMLPolicyVersion{
		VersionId:        "v1",
		Document:         policyDocument,
		IsDefaultVersion: true,
		CreateDate:       now,
	}

	versionKey := fmt.Sprintf("iam:policy-version:%s:%s:v1", defaultAccountID, policyName)
	if err := s.state.Set(versionKey, &version); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store policy version"), nil
	}

	// Register policy in the relationship graph
	s.registerResource("policy", policyName, map[string]string{
		"arn":  policy.Arn,
		"path": policy.Path,
	})

	result := CreatePolicyResult{Policy: policy}
	return s.successResponse("CreatePolicy", result)
}

func (s *IAMService) getPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "Invalid PolicyArn format"), nil
	}

	// Check if this is an AWS managed policy
	if managedPolicy := getAWSManagedPolicy(policyArn); managedPolicy != nil {
		result := GetPolicyResult{Policy: managedPolicy.toXMLPolicy()}
		return s.successResponse("GetPolicy", result)
	}

	var policy XMLPolicy
	stateKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)
	if err := s.state.Get(stateKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s does not exist or is not attachable.", policyArn)), nil
	}

	result := GetPolicyResult{Policy: policy}
	return s.successResponse("GetPolicy", result)
}

func (s *IAMService) getPolicyVersion(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	versionId := getStringValue(params, "VersionId")
	if versionId == "" {
		return s.errorResponse(400, "InvalidInput", "VersionId is required"), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "Invalid PolicyArn format"), nil
	}

	// Check if this is an AWS managed policy
	if managedPolicy := getAWSManagedPolicy(policyArn); managedPolicy != nil {
		// AWS managed policies only have v1
		if versionId != "v1" {
			return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy version %s does not exist.", versionId)), nil
		}
		result := GetPolicyVersionResult{PolicyVersion: managedPolicy.toXMLPolicyVersion()}
		return s.successResponse("GetPolicyVersion", result)
	}

	var version XMLPolicyVersion
	versionKey := fmt.Sprintf("iam:policy-version:%s:%s:%s", defaultAccountID, policyName, versionId)
	if err := s.state.Get(versionKey, &version); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy version %s does not exist.", versionId)), nil
	}

	result := GetPolicyVersionResult{PolicyVersion: version}
	return s.successResponse("GetPolicyVersion", result)
}

func (s *IAMService) deletePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "Invalid PolicyArn format"), nil
	}

	stateKey := fmt.Sprintf("iam:policy:%s:%s", defaultAccountID, policyName)

	var policy XMLPolicy
	if err := s.state.Get(stateKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Policy %s does not exist.", policyArn)), nil
	}

	// Unregister from graph (validates no dependents via graph relationships)
	// The graph tracks role-policy attachments as edges
	if err := s.unregisterResource("policy", policyName); err != nil {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete policy: %v", err)), nil
	}

	// Delete all versions
	versionKeys, _ := s.state.List(fmt.Sprintf("iam:policy-version:%s:%s:", defaultAccountID, policyName))
	for _, key := range versionKeys {
		s.state.Delete(key)
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete policy"), nil
	}

	return s.successResponse("DeletePolicy", EmptyResult{})
}

func (s *IAMService) listPolicyVersions(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	policyArn := getStringValue(params, "PolicyArn")
	if policyArn == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyArn is required"), nil
	}

	policyName := extractPolicyNameFromArn(policyArn)
	if policyName == "" {
		return s.errorResponse(400, "InvalidInput", "Invalid PolicyArn format"), nil
	}

	prefix := fmt.Sprintf("iam:policy-version:%s:%s:", defaultAccountID, policyName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list policy versions"), nil
	}

	var versions []XMLPolicyVersion
	for _, key := range keys {
		var version XMLPolicyVersion
		if err := s.state.Get(key, &version); err == nil {
			versions = append(versions, version)
		}
	}

	result := ListPolicyVersionsResult{
		Versions:    versions,
		IsTruncated: false,
	}
	return s.successResponse("ListPolicyVersions", result)
}

func (s *IAMService) listPolicies(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := getStringValue(params, "PathPrefix")

	keys, err := s.state.List("iam:policy:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list policies"), nil
	}

	var policies []XMLPolicy
	for _, key := range keys {
		// Skip version keys
		if strings.Contains(key, "policy-version:") {
			continue
		}
		var policy XMLPolicy
		if err := s.state.Get(key, &policy); err == nil {
			if pathPrefix == "" || strings.HasPrefix(policy.Path, pathPrefix) {
				policies = append(policies, policy)
			}
		}
	}

	result := ListPoliciesResult{
		Policies:    policies,
		IsTruncated: false,
	}
	return s.successResponse("ListPolicies", result)
}
