package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

const (
	defaultAccountID     = "123456789012"
	defaultMaxSessionDur = 3600
)

// IAMService implements the AWS IAM service emulator
type IAMService struct {
	state           emulator.StateManager
	validator       emulator.Validator
	resourceManager *graph.ResourceManager
}

// NewIAMService creates a new IAM service instance
func NewIAMService(state emulator.StateManager, validator emulator.Validator) *IAMService {
	return &IAMService{
		state:     state,
		validator: validator,
	}
}

// NewIAMServiceWithGraph creates a new IAM service instance with ResourceManager for relationship tracking
func NewIAMServiceWithGraph(state emulator.StateManager, validator emulator.Validator, rm *graph.ResourceManager) *IAMService {
	return &IAMService{
		state:           state,
		validator:       validator,
		resourceManager: rm,
	}
}

// ServiceName returns the service identifier
func (s *IAMService) ServiceName() string {
	return "iam"
}

// SupportedActions returns the list of AWS API actions this service handles.
// Used by the router to determine which service handles a given Query Protocol request.
func (s *IAMService) SupportedActions() []string {
	return []string{
		// Role operations
		"CreateRole",
		"GetRole",
		"DeleteRole",
		"UpdateAssumeRolePolicy",
		"ListRoles",
		// Policy operations
		"CreatePolicy",
		"GetPolicy",
		"GetPolicyVersion",
		"DeletePolicy",
		"ListPolicyVersions",
		"ListPolicies",
		// Policy attachment operations (roles)
		"AttachRolePolicy",
		"DetachRolePolicy",
		"ListAttachedRolePolicies",
		// Inline policy operations (roles)
		"PutRolePolicy",
		"GetRolePolicy",
		"DeleteRolePolicy",
		"ListRolePolicies",
		// Instance profile operations
		"CreateInstanceProfile",
		"GetInstanceProfile",
		"DeleteInstanceProfile",
		"AddRoleToInstanceProfile",
		"RemoveRoleFromInstanceProfile",
		"ListInstanceProfiles",
		"ListInstanceProfilesForRole",
		// User operations
		"CreateUser",
		"GetUser",
		"DeleteUser",
		"UpdateUser",
		"ListUsers",
		// User tag operations
		"TagUser",
		"UntagUser",
		"ListUserTags",
		// User login profile operations
		"CreateLoginProfile",
		"GetLoginProfile",
		"UpdateLoginProfile",
		"DeleteLoginProfile",
		// User policy attachment operations
		"AttachUserPolicy",
		"DetachUserPolicy",
		"ListAttachedUserPolicies",
		// User inline policy operations
		"PutUserPolicy",
		"GetUserPolicy",
		"DeleteUserPolicy",
		"ListUserPolicies",
		// Group operations
		"CreateGroup",
		"GetGroup",
		"DeleteGroup",
		"UpdateGroup",
		"ListGroups",
		// Group membership operations
		"AddUserToGroup",
		"RemoveUserFromGroup",
		"ListGroupsForUser",
		// Group policy attachment operations
		"AttachGroupPolicy",
		"DetachGroupPolicy",
		"ListAttachedGroupPolicies",
		// Group inline policy operations
		"PutGroupPolicy",
		"GetGroupPolicy",
		"DeleteGroupPolicy",
		"ListGroupPolicies",
		// Access key operations
		"CreateAccessKey",
		"DeleteAccessKey",
		"UpdateAccessKey",
		"ListAccessKeys",
		"GetAccessKeyLastUsed",
		// Role enhancement operations
		"UpdateRole",
		"UpdateRoleDescription",
		"TagRole",
		"UntagRole",
		"ListRoleTags",
		// Service-linked role operations
		"CreateServiceLinkedRole",
		"DeleteServiceLinkedRole",
		"GetServiceLinkedRoleDeletionStatus",
		// SAML provider operations
		"CreateSAMLProvider",
		"GetSAMLProvider",
		"UpdateSAMLProvider",
		"DeleteSAMLProvider",
		"ListSAMLProviders",
		// OIDC provider operations
		"CreateOpenIDConnectProvider",
		"GetOpenIDConnectProvider",
		"DeleteOpenIDConnectProvider",
		"ListOpenIDConnectProviders",
		"UpdateOpenIDConnectProviderThumbprint",
		// MFA device operations
		"CreateVirtualMFADevice",
		"DeleteVirtualMFADevice",
		"ListVirtualMFADevices",
		"EnableMFADevice",
		"DeactivateMFADevice",
		"ListMFADevices",
		"ResyncMFADevice",
		// Server certificate operations
		"UploadServerCertificate",
		"GetServerCertificate",
		"DeleteServerCertificate",
		"ListServerCertificates",
		"UpdateServerCertificate",
		// SSH public key operations
		"UploadSSHPublicKey",
		"GetSSHPublicKey",
		"UpdateSSHPublicKey",
		"DeleteSSHPublicKey",
		"ListSSHPublicKeys",
		// Account alias operations
		"CreateAccountAlias",
		"DeleteAccountAlias",
		"ListAccountAliases",
		// Account password policy operations
		"UpdateAccountPasswordPolicy",
		"GetAccountPasswordPolicy",
		"DeleteAccountPasswordPolicy",
	}
}

// HandleRequest routes incoming requests to the appropriate handler
func (s *IAMService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	params, err := s.parseParameters(req)
	if err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	if err := s.validator.ValidateAction(action, params); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	switch action {
	// Role operations
	case "CreateRole":
		return s.createRole(ctx, params)
	case "GetRole":
		return s.getRole(ctx, params)
	case "DeleteRole":
		return s.deleteRole(ctx, params)
	case "UpdateAssumeRolePolicy":
		return s.updateAssumeRolePolicy(ctx, params)
	case "ListRoles":
		return s.listRoles(ctx, params)

	// Policy operations
	case "CreatePolicy":
		return s.createPolicy(ctx, params)
	case "GetPolicy":
		return s.getPolicy(ctx, params)
	case "GetPolicyVersion":
		return s.getPolicyVersion(ctx, params)
	case "DeletePolicy":
		return s.deletePolicy(ctx, params)
	case "ListPolicyVersions":
		return s.listPolicyVersions(ctx, params)
	case "ListPolicies":
		return s.listPolicies(ctx, params)

	// Policy attachment operations
	case "AttachRolePolicy":
		return s.attachRolePolicy(ctx, params)
	case "DetachRolePolicy":
		return s.detachRolePolicy(ctx, params)
	case "ListAttachedRolePolicies":
		return s.listAttachedRolePolicies(ctx, params)

	// Inline policy operations
	case "PutRolePolicy":
		return s.putRolePolicy(ctx, params)
	case "GetRolePolicy":
		return s.getRolePolicy(ctx, params)
	case "DeleteRolePolicy":
		return s.deleteRolePolicy(ctx, params)
	case "ListRolePolicies":
		return s.listRolePolicies(ctx, params)

	// Instance profile operations
	case "CreateInstanceProfile":
		return s.createInstanceProfile(ctx, params)
	case "GetInstanceProfile":
		return s.getInstanceProfile(ctx, params)
	case "DeleteInstanceProfile":
		return s.deleteInstanceProfile(ctx, params)
	case "AddRoleToInstanceProfile":
		return s.addRoleToInstanceProfile(ctx, params)
	case "RemoveRoleFromInstanceProfile":
		return s.removeRoleFromInstanceProfile(ctx, params)
	case "ListInstanceProfiles":
		return s.listInstanceProfiles(ctx, params)
	case "ListInstanceProfilesForRole":
		return s.listInstanceProfilesForRole(ctx, params)

	// User operations
	case "CreateUser":
		return s.createUser(ctx, params)
	case "GetUser":
		return s.getUser(ctx, params)
	case "DeleteUser":
		return s.deleteUser(ctx, params)
	case "UpdateUser":
		return s.updateUser(ctx, params)
	case "ListUsers":
		return s.listUsers(ctx, params)

	// User tag operations
	case "TagUser":
		return s.tagUser(ctx, params)
	case "UntagUser":
		return s.untagUser(ctx, params)
	case "ListUserTags":
		return s.listUserTags(ctx, params)

	// User login profile operations
	case "CreateLoginProfile":
		return s.createLoginProfile(ctx, params)
	case "GetLoginProfile":
		return s.getLoginProfile(ctx, params)
	case "UpdateLoginProfile":
		return s.updateLoginProfile(ctx, params)
	case "DeleteLoginProfile":
		return s.deleteLoginProfile(ctx, params)

	// User policy attachment operations
	case "AttachUserPolicy":
		return s.attachUserPolicy(ctx, params)
	case "DetachUserPolicy":
		return s.detachUserPolicy(ctx, params)
	case "ListAttachedUserPolicies":
		return s.listAttachedUserPolicies(ctx, params)

	// User inline policy operations
	case "PutUserPolicy":
		return s.putUserPolicy(ctx, params)
	case "GetUserPolicy":
		return s.getUserPolicy(ctx, params)
	case "DeleteUserPolicy":
		return s.deleteUserPolicy(ctx, params)
	case "ListUserPolicies":
		return s.listUserPolicies(ctx, params)

	// Group operations
	case "CreateGroup":
		return s.createGroup(ctx, params)
	case "GetGroup":
		return s.getGroup(ctx, params)
	case "DeleteGroup":
		return s.deleteGroup(ctx, params)
	case "UpdateGroup":
		return s.updateGroup(ctx, params)
	case "ListGroups":
		return s.listGroups(ctx, params)

	// Group membership operations
	case "AddUserToGroup":
		return s.addUserToGroup(ctx, params)
	case "RemoveUserFromGroup":
		return s.removeUserFromGroup(ctx, params)
	case "ListGroupsForUser":
		return s.listGroupsForUser(ctx, params)

	// Group policy attachment operations
	case "AttachGroupPolicy":
		return s.attachGroupPolicy(ctx, params)
	case "DetachGroupPolicy":
		return s.detachGroupPolicy(ctx, params)
	case "ListAttachedGroupPolicies":
		return s.listAttachedGroupPolicies(ctx, params)

	// Group inline policy operations
	case "PutGroupPolicy":
		return s.putGroupPolicy(ctx, params)
	case "GetGroupPolicy":
		return s.getGroupPolicy(ctx, params)
	case "DeleteGroupPolicy":
		return s.deleteGroupPolicy(ctx, params)
	case "ListGroupPolicies":
		return s.listGroupPolicies(ctx, params)

	// Access key operations
	case "CreateAccessKey":
		return s.createAccessKey(ctx, params)
	case "DeleteAccessKey":
		return s.deleteAccessKey(ctx, params)
	case "UpdateAccessKey":
		return s.updateAccessKey(ctx, params)
	case "ListAccessKeys":
		return s.listAccessKeys(ctx, params)
	case "GetAccessKeyLastUsed":
		return s.getAccessKeyLastUsed(ctx, params)

	// Role enhancement operations
	case "UpdateRole":
		return s.updateRole(ctx, params)
	case "UpdateRoleDescription":
		return s.updateRoleDescription(ctx, params)
	case "TagRole":
		return s.tagRole(ctx, params)
	case "UntagRole":
		return s.untagRole(ctx, params)
	case "ListRoleTags":
		return s.listRoleTags(ctx, params)

	// Service-linked role operations
	case "CreateServiceLinkedRole":
		return s.createServiceLinkedRole(ctx, params)
	case "DeleteServiceLinkedRole":
		return s.deleteServiceLinkedRole(ctx, params)
	case "GetServiceLinkedRoleDeletionStatus":
		return s.getServiceLinkedRoleDeletionStatus(ctx, params)

	// SAML provider operations
	case "CreateSAMLProvider":
		return s.createSAMLProvider(ctx, params)
	case "GetSAMLProvider":
		return s.getSAMLProvider(ctx, params)
	case "UpdateSAMLProvider":
		return s.updateSAMLProvider(ctx, params)
	case "DeleteSAMLProvider":
		return s.deleteSAMLProvider(ctx, params)
	case "ListSAMLProviders":
		return s.listSAMLProviders(ctx, params)

	// OIDC provider operations
	case "CreateOpenIDConnectProvider":
		return s.createOpenIDConnectProvider(ctx, params)
	case "GetOpenIDConnectProvider":
		return s.getOpenIDConnectProvider(ctx, params)
	case "DeleteOpenIDConnectProvider":
		return s.deleteOpenIDConnectProvider(ctx, params)
	case "ListOpenIDConnectProviders":
		return s.listOpenIDConnectProviders(ctx, params)
	case "UpdateOpenIDConnectProviderThumbprint":
		return s.updateOpenIDConnectProviderThumbprint(ctx, params)

	// MFA device operations
	case "CreateVirtualMFADevice":
		return s.createVirtualMFADevice(ctx, params)
	case "DeleteVirtualMFADevice":
		return s.deleteVirtualMFADevice(ctx, params)
	case "ListVirtualMFADevices":
		return s.listVirtualMFADevices(ctx, params)
	case "EnableMFADevice":
		return s.enableMFADevice(ctx, params)
	case "DeactivateMFADevice":
		return s.deactivateMFADevice(ctx, params)
	case "ListMFADevices":
		return s.listMFADevices(ctx, params)
	case "ResyncMFADevice":
		return s.resyncMFADevice(ctx, params)

	// Server certificate operations
	case "UploadServerCertificate":
		return s.uploadServerCertificate(ctx, params)
	case "GetServerCertificate":
		return s.getServerCertificate(ctx, params)
	case "DeleteServerCertificate":
		return s.deleteServerCertificate(ctx, params)
	case "ListServerCertificates":
		return s.listServerCertificates(ctx, params)
	case "UpdateServerCertificate":
		return s.updateServerCertificate(ctx, params)

	// SSH public key operations
	case "UploadSSHPublicKey":
		return s.uploadSSHPublicKey(ctx, params)
	case "GetSSHPublicKey":
		return s.getSSHPublicKey(ctx, params)
	case "UpdateSSHPublicKey":
		return s.updateSSHPublicKey(ctx, params)
	case "DeleteSSHPublicKey":
		return s.deleteSSHPublicKey(ctx, params)
	case "ListSSHPublicKeys":
		return s.listSSHPublicKeys(ctx, params)

	// Account alias operations
	case "CreateAccountAlias":
		return s.createAccountAlias(ctx, params)
	case "DeleteAccountAlias":
		return s.deleteAccountAlias(ctx, params)
	case "ListAccountAliases":
		return s.listAccountAliases(ctx, params)

	// Account password policy operations
	case "UpdateAccountPasswordPolicy":
		return s.updateAccountPasswordPolicy(ctx, params)
	case "GetAccountPasswordPolicy":
		return s.getAccountPasswordPolicy(ctx, params)
	case "DeleteAccountPasswordPolicy":
		return s.deleteAccountPasswordPolicy(ctx, params)

	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *IAMService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *IAMService) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
	if req.Parameters != nil {
		return req.Parameters, nil
	}

	contentType := req.Headers["Content-Type"]
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return s.parseFormData(string(req.Body))
	}

	if strings.Contains(contentType, "application/json") {
		var params map[string]interface{}
		if err := json.Unmarshal(req.Body, &params); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %w", err)
		}
		return params, nil
	}

	return make(map[string]interface{}), nil
}

func (s *IAMService) parseFormData(body string) (map[string]interface{}, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			params[key] = vals[0]
		} else {
			params[key] = vals
		}
	}

	return params, nil
}
