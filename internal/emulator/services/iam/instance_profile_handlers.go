package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

func (s *IAMService) createInstanceProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	profileName := getStringValue(params, "InstanceProfileName")
	if profileName == "" {
		return s.errorResponse(400, "InvalidInput", "InstanceProfileName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:instance-profile:%s", profileName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Instance Profile %s already exists.", profileName)), nil
	}

	path := getStringValue(params, "Path")
	if path == "" {
		path = "/"
	}

	profile := XMLInstanceProfile{
		InstanceProfileName: profileName,
		InstanceProfileId:   generateIAMId("AIPA"),
		Arn:                 fmt.Sprintf("arn:aws:iam::%s:instance-profile%s%s", defaultAccountID, path, profileName),
		Path:                path,
		CreateDate:          time.Now().UTC(),
		Roles:               []XMLRoleListItem{},
		Tags:                s.parseTags(params),
	}

	if err := s.state.Set(stateKey, &profile); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store instance profile"), nil
	}

	// Register instance profile in the relationship graph
	s.registerResource("instance-profile", profileName, map[string]string{
		"arn":  profile.Arn,
		"path": profile.Path,
	})

	result := CreateInstanceProfileResult{InstanceProfile: profile}
	return s.successResponse("CreateInstanceProfile", result)
}

func (s *IAMService) getInstanceProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	profileName := getStringValue(params, "InstanceProfileName")
	if profileName == "" {
		return s.errorResponse(400, "InvalidInput", "InstanceProfileName is required"), nil
	}

	var profile XMLInstanceProfile
	stateKey := fmt.Sprintf("iam:instance-profile:%s", profileName)
	if err := s.state.Get(stateKey, &profile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Instance profile %s cannot be found.", profileName)), nil
	}

	result := GetInstanceProfileResult{InstanceProfile: profile}
	return s.successResponse("GetInstanceProfile", result)
}

func (s *IAMService) deleteInstanceProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	profileName := getStringValue(params, "InstanceProfileName")
	if profileName == "" {
		return s.errorResponse(400, "InvalidInput", "InstanceProfileName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:instance-profile:%s", profileName)

	var profile XMLInstanceProfile
	if err := s.state.Get(stateKey, &profile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Instance profile %s cannot be found.", profileName)), nil
	}

	// Unregister from graph (validates no dependents via graph relationships)
	// The graph tracks instance-profile to role relationships as edges
	if err := s.unregisterResource("instance-profile", profileName); err != nil {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete instance profile: %v", err)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete instance profile"), nil
	}

	return s.successResponse("DeleteInstanceProfile", EmptyResult{})
}

func (s *IAMService) addRoleToInstanceProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	profileName := getStringValue(params, "InstanceProfileName")
	if profileName == "" {
		return s.errorResponse(400, "InvalidInput", "InstanceProfileName is required"), nil
	}

	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	// Verify role exists
	var role XMLRole
	roleKey := fmt.Sprintf("iam:role:%s", roleName)
	if err := s.state.Get(roleKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Get instance profile
	profileKey := fmt.Sprintf("iam:instance-profile:%s", profileName)
	var profile XMLInstanceProfile
	if err := s.state.Get(profileKey, &profile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Instance profile %s cannot be found.", profileName)), nil
	}

	// Check if role already in profile
	for _, r := range profile.Roles {
		if r.RoleName == roleName {
			// Already added, idempotent success
			return s.successResponse("AddRoleToInstanceProfile", EmptyResult{})
		}
	}

	// AWS limits instance profiles to 1 role
	if len(profile.Roles) >= 1 {
		return s.errorResponse(400, "LimitExceeded", "Cannot exceed quota for RolesPerInstanceProfile: 1"), nil
	}

	profile.Roles = append(profile.Roles, roleToListItem(role))
	if err := s.state.Set(profileKey, &profile); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update instance profile"), nil
	}

	// Add relationship in graph: instance-profile -> role (profile contains role)
	if err := s.addRelationship("instance-profile", profileName, "role", roleName, graph.RelContains); err != nil {
		if s.isStrictMode() {
			// Rollback: remove the role from profile
			profile.Roles = profile.Roles[:len(profile.Roles)-1]
			s.state.Set(profileKey, &profile)
			return s.errorResponse(500, "InternalFailure", fmt.Sprintf("Failed to create instance-profile-role relationship: %v", err)), nil
		}
		log.Printf("Warning: failed to add instance-profile-role relationship in graph: %v", err)
	}

	return s.successResponse("AddRoleToInstanceProfile", EmptyResult{})
}

func (s *IAMService) removeRoleFromInstanceProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	profileName := getStringValue(params, "InstanceProfileName")
	if profileName == "" {
		return s.errorResponse(400, "InvalidInput", "InstanceProfileName is required"), nil
	}

	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	// Get instance profile
	profileKey := fmt.Sprintf("iam:instance-profile:%s", profileName)
	var profile XMLInstanceProfile
	if err := s.state.Get(profileKey, &profile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Instance profile %s cannot be found.", profileName)), nil
	}

	// Find and remove role
	found := false
	newRoles := make([]XMLRoleListItem, 0, len(profile.Roles))
	for _, r := range profile.Roles {
		if r.RoleName == roleName {
			found = true
		} else {
			newRoles = append(newRoles, r)
		}
	}

	if !found {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Role %s is not in instance profile %s.", roleName, profileName)), nil
	}

	profile.Roles = newRoles
	if err := s.state.Set(profileKey, &profile); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update instance profile"), nil
	}

	// Remove relationship in graph: instance-profile -> role
	if err := s.removeRelationship("instance-profile", profileName, "role", roleName, graph.RelContains); err != nil {
		log.Printf("Warning: failed to remove instance-profile-role relationship in graph: %v", err)
	}

	return s.successResponse("RemoveRoleFromInstanceProfile", EmptyResult{})
}

func (s *IAMService) listInstanceProfiles(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := getStringValue(params, "PathPrefix")

	keys, err := s.state.List("iam:instance-profile:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list instance profiles"), nil
	}

	var profiles []XMLInstanceProfileListItem
	for _, key := range keys {
		var profile XMLInstanceProfile
		if err := s.state.Get(key, &profile); err == nil {
			if pathPrefix == "" || strings.HasPrefix(profile.Path, pathPrefix) {
				profiles = append(profiles, instanceProfileToListItem(profile))
			}
		}
	}

	result := ListInstanceProfilesResult{
		InstanceProfiles: profiles,
		IsTruncated:      false,
	}
	return s.successResponse("ListInstanceProfiles", result)
}

func (s *IAMService) listInstanceProfilesForRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "ValidationError", "RoleName is required"), nil
	}

	// Verify role exists
	roleKey := fmt.Sprintf("iam:roles:%s", roleName)
	var role XMLRole
	if err := s.state.Get(roleKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Find instance profiles containing this role
	keys, err := s.state.List("iam:instance-profile:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list instance profiles"), nil
	}

	var profiles []XMLInstanceProfileListItem
	for _, key := range keys {
		var profile XMLInstanceProfile
		if err := s.state.Get(key, &profile); err == nil {
			for _, r := range profile.Roles {
				if r.RoleName == roleName {
					profiles = append(profiles, instanceProfileToListItem(profile))
					break
				}
			}
		}
	}

	result := ListInstanceProfilesForRoleResult{
		InstanceProfiles: profiles,
		IsTruncated:      false,
	}
	return s.successResponse("ListInstanceProfilesForRole", result)
}
