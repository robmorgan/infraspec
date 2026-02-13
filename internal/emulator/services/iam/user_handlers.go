package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// User CRUD Operations
// ============================================================================

func (s *IAMService) createUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Check if user already exists
	stateKey := fmt.Sprintf("iam:user:%s", userName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("User with name %s already exists.", userName)), nil
	}

	path := getStringValue(params, "Path")
	if path == "" {
		path = "/"
	}

	user := XMLUser{
		UserName:   userName,
		UserId:     generateIAMId("AIDA"),
		Arn:        fmt.Sprintf("arn:aws:iam::%s:user%s%s", defaultAccountID, path, userName),
		Path:       path,
		CreateDate: time.Now().UTC(),
		Tags:       s.parseTags(params),
	}

	if err := s.state.Set(stateKey, &user); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store user"), nil
	}

	// Initialize empty policy attachments
	attachKey := fmt.Sprintf("iam:user-policies:%s", userName)
	if err := s.state.Set(attachKey, &UserAttachments{PolicyArns: []string{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to initialize user attachments"), nil
	}

	// Register user in the relationship graph
	s.registerResource("user", userName, map[string]string{
		"arn":  user.Arn,
		"path": user.Path,
	})

	result := CreateUserResult{User: user}
	return s.successResponse("CreateUser", result)
}

func (s *IAMService) getUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	var user XMLUser
	stateKey := fmt.Sprintf("iam:user:%s", userName)
	if err := s.state.Get(stateKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	result := GetUserResult{User: user}
	return s.successResponse("GetUser", result)
}

func (s *IAMService) deleteUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Check for login profile - must be deleted first
	loginKey := fmt.Sprintf("iam:user-login-profile:%s", userName)
	if s.state.Exists(loginKey) {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must delete login profile first."), nil
	}

	// Check for attached policies - must be detached first
	attachKey := fmt.Sprintf("iam:user-policies:%s", userName)
	var attachments UserAttachments
	if err := s.state.Get(attachKey, &attachments); err == nil && len(attachments.PolicyArns) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must detach all policies first."), nil
	}

	// Check for inline policies - must be deleted first
	inlineKey := fmt.Sprintf("iam:user-inline-policies:%s", userName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err == nil && len(inlinePolicies.Policies) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must delete inline policies first."), nil
	}

	// Check for access keys - must be deleted first
	accessKeyPrefix := fmt.Sprintf("iam:access-key:%s:", userName)
	if accessKeys, err := s.state.List(accessKeyPrefix); err == nil && len(accessKeys) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must delete access keys first."), nil
	}

	// Check for SSH public keys - must be deleted first
	sshKeyPrefix := fmt.Sprintf("iam:ssh-public-key:%s:", userName)
	if sshKeys, err := s.state.List(sshKeyPrefix); err == nil && len(sshKeys) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must delete SSH public keys first."), nil
	}

	// Check for MFA devices - must be deactivated first
	if s.userHasMFADevices(userName) {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must deactivate MFA devices first."), nil
	}

	// Check for group memberships - must be removed first
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err == nil && len(userGroups.GroupNames) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must remove user from all groups first."), nil
	}

	// Unregister from graph
	if err := s.unregisterResource("user", userName); err != nil {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete user: %v", err)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete user"), nil
	}

	// Clean up related state
	s.state.Delete(attachKey)
	s.state.Delete(inlineKey)
	s.state.Delete(userGroupsKey)

	return s.successResponse("DeleteUser", EmptyResult{})
}

// userHasMFADevices checks if a user has any MFA devices assigned
func (s *IAMService) userHasMFADevices(userName string) bool {
	keys, err := s.state.List("iam:mfa-device:")
	if err != nil {
		return false
	}
	for _, key := range keys {
		var device VirtualMFADeviceData
		if err := s.state.Get(key, &device); err == nil {
			if device.UserName == userName {
				return true
			}
		}
	}
	return false
}

func (s *IAMService) updateUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(stateKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	newUserName := getStringValue(params, "NewUserName")
	newPath := getStringValue(params, "NewPath")

	// If renaming user
	if newUserName != "" && newUserName != userName {
		newStateKey := fmt.Sprintf("iam:user:%s", newUserName)
		if s.state.Exists(newStateKey) {
			return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("User with name %s already exists.", newUserName)), nil
		}

		// Update user fields
		user.UserName = newUserName
		if newPath != "" {
			user.Path = newPath
		}
		user.Arn = fmt.Sprintf("arn:aws:iam::%s:user%s%s", defaultAccountID, user.Path, newUserName)

		// Store with new key first (safer order - new key exists before old is deleted)
		if err := s.state.Set(newStateKey, &user); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to update user"), nil
		}

		// Only delete old key after new key is successfully created
		s.state.Delete(stateKey)

		// Migrate related state
		s.migrateUserState(userName, newUserName)

		// Update graph registration
		s.unregisterResource("user", userName)
		s.registerResource("user", newUserName, map[string]string{
			"arn":  user.Arn,
			"path": user.Path,
		})
	} else if newPath != "" {
		// Just updating path
		user.Path = newPath
		user.Arn = fmt.Sprintf("arn:aws:iam::%s:user%s%s", defaultAccountID, user.Path, userName)
		if err := s.state.Set(stateKey, &user); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to update user"), nil
		}
	}

	return s.successResponse("UpdateUser", EmptyResult{})
}

func (s *IAMService) listUsers(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := getStringValue(params, "PathPrefix")

	keys, err := s.state.List("iam:user:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list users"), nil
	}

	var users []XMLUserListItem
	for _, key := range keys {
		// Skip non-user keys (like iam:user-policies:, iam:user-inline-policies:)
		if strings.Contains(key, "user-policies:") || strings.Contains(key, "user-inline-policies:") || strings.Contains(key, "user-login-profile:") {
			continue
		}
		var user XMLUser
		if err := s.state.Get(key, &user); err == nil {
			if pathPrefix == "" || strings.HasPrefix(user.Path, pathPrefix) {
				users = append(users, userToListItem(user))
			}
		}
	}

	result := ListUsersResult{
		Users:       users,
		IsTruncated: false,
	}
	return s.successResponse("ListUsers", result)
}

// ============================================================================
// User Tag Operations
// ============================================================================

func (s *IAMService) tagUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(stateKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	newTags := s.parseTags(params)
	if len(newTags) == 0 {
		return s.errorResponse(400, "ValidationError", "Tags is required"), nil
	}

	// Merge tags - new tags override existing ones with same key
	tagMap := make(map[string]string)
	for _, tag := range user.Tags {
		tagMap[tag.Key] = tag.Value
	}
	for _, tag := range newTags {
		tagMap[tag.Key] = tag.Value
	}

	// Convert back to slice
	user.Tags = make([]XMLTag, 0, len(tagMap))
	for k, v := range tagMap {
		user.Tags = append(user.Tags, XMLTag{Key: k, Value: v})
	}

	if err := s.state.Set(stateKey, &user); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update user tags"), nil
	}

	return s.successResponse("TagUser", EmptyResult{})
}

func (s *IAMService) untagUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(stateKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Parse tag keys to remove
	tagKeysToRemove := s.parseTagKeys(params)
	if len(tagKeysToRemove) == 0 {
		return s.errorResponse(400, "ValidationError", "TagKeys is required"), nil
	}

	// Remove specified tags
	removeSet := make(map[string]bool)
	for _, key := range tagKeysToRemove {
		removeSet[key] = true
	}

	newTags := make([]XMLTag, 0)
	for _, tag := range user.Tags {
		if !removeSet[tag.Key] {
			newTags = append(newTags, tag)
		}
	}
	user.Tags = newTags

	if err := s.state.Set(stateKey, &user); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update user tags"), nil
	}

	return s.successResponse("UntagUser", EmptyResult{})
}

func (s *IAMService) listUserTags(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(stateKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	result := ListUserTagsResult{
		Tags:        user.Tags,
		IsTruncated: false,
	}
	return s.successResponse("ListUserTags", result)
}

// ============================================================================
// Helper Functions
// ============================================================================

// migrateUserState migrates user-related state when a user is renamed.
// Note: This is best-effort migration for the in-memory emulator. In a production
// system, this would require transactional semantics. For testing purposes,
// partial migration failures are acceptable as state resets on restart.
func (s *IAMService) migrateUserState(oldUserName, newUserName string) {
	// Migrate policy attachments
	oldAttachKey := fmt.Sprintf("iam:user-policies:%s", oldUserName)
	newAttachKey := fmt.Sprintf("iam:user-policies:%s", newUserName)
	var attachments UserAttachments
	if err := s.state.Get(oldAttachKey, &attachments); err == nil {
		if err := s.state.Set(newAttachKey, &attachments); err == nil {
			s.state.Delete(oldAttachKey)
		}
	}

	// Migrate inline policies
	oldInlineKey := fmt.Sprintf("iam:user-inline-policies:%s", oldUserName)
	newInlineKey := fmt.Sprintf("iam:user-inline-policies:%s", newUserName)
	var inlinePolicies UserInlinePolicies
	if err := s.state.Get(oldInlineKey, &inlinePolicies); err == nil {
		if err := s.state.Set(newInlineKey, &inlinePolicies); err == nil {
			s.state.Delete(oldInlineKey)
		}
	}

	// Migrate login profile
	oldLoginKey := fmt.Sprintf("iam:user-login-profile:%s", oldUserName)
	newLoginKey := fmt.Sprintf("iam:user-login-profile:%s", newUserName)
	var loginProfile UserLoginProfile
	if err := s.state.Get(oldLoginKey, &loginProfile); err == nil {
		if err := s.state.Set(newLoginKey, &loginProfile); err == nil {
			s.state.Delete(oldLoginKey)
		}
	}
}

// parseTagKeys parses tag keys from request parameters
func (s *IAMService) parseTagKeys(params map[string]interface{}) []string {
	var keys []string
	tagIndex := 1

	for {
		keyParam := fmt.Sprintf("TagKeys.member.%d", tagIndex)
		key, hasKey := params[keyParam].(string)

		if !hasKey {
			// Try TagKeys.TagKey.N format
			keyParam = fmt.Sprintf("TagKeys.TagKey.%d", tagIndex)
			key, hasKey = params[keyParam].(string)
		}

		if !hasKey {
			break
		}

		keys = append(keys, key)
		tagIndex++
	}

	return keys
}
