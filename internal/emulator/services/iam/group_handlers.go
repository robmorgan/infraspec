package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// Group CRUD Operations
// ============================================================================

func (s *IAMService) createGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	// Check if group already exists
	stateKey := fmt.Sprintf("iam:group:%s", groupName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Group with name %s already exists.", groupName)), nil
	}

	path := getStringValue(params, "Path")
	if path == "" {
		path = "/"
	}

	group := XMLGroup{
		GroupName:  groupName,
		GroupId:    generateIAMId("AGPA"),
		Arn:        fmt.Sprintf("arn:aws:iam::%s:group%s%s", defaultAccountID, path, groupName),
		Path:       path,
		CreateDate: time.Now().UTC(),
	}

	if err := s.state.Set(stateKey, &group); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store group"), nil
	}

	// Initialize empty policy attachments and members
	attachKey := fmt.Sprintf("iam:group-policies:%s", groupName)
	if err := s.state.Set(attachKey, &GroupAttachments{PolicyArns: []string{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to initialize group attachments"), nil
	}

	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	if err := s.state.Set(membersKey, &GroupMembers{UserNames: []string{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to initialize group members"), nil
	}

	// Register group in the relationship graph
	s.registerResource("group", groupName, map[string]string{
		"arn":  group.Arn,
		"path": group.Path,
	})

	result := CreateGroupResult{Group: group}
	return s.successResponse("CreateGroup", result)
}

func (s *IAMService) getGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	var group XMLGroup
	stateKey := fmt.Sprintf("iam:group:%s", groupName)
	if err := s.state.Get(stateKey, &group); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Get group members
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		members = GroupMembers{UserNames: []string{}}
	}

	// Get user details for each member
	var users []XMLUserListItem
	for _, userName := range members.UserNames {
		userKey := fmt.Sprintf("iam:user:%s", userName)
		var user XMLUser
		if err := s.state.Get(userKey, &user); err == nil {
			users = append(users, userToListItem(user))
		}
	}

	result := GetGroupResult{
		Group:       group,
		Users:       users,
		IsTruncated: false,
	}
	return s.successResponse("GetGroup", result)
}

func (s *IAMService) deleteGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Check for group members - must be removed first
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err == nil && len(members.UserNames) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must remove all users from group first."), nil
	}

	// Check for attached policies - must be detached first
	attachKey := fmt.Sprintf("iam:group-policies:%s", groupName)
	var attachments GroupAttachments
	if err := s.state.Get(attachKey, &attachments); err == nil && len(attachments.PolicyArns) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must detach all policies first."), nil
	}

	// Check for inline policies - must be deleted first
	inlineKey := fmt.Sprintf("iam:group-inline-policies:%s", groupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(inlineKey, &inlinePolicies); err == nil && len(inlinePolicies.Policies) > 0 {
		return s.errorResponse(409, "DeleteConflict", "Cannot delete entity, must delete inline policies first."), nil
	}

	// Unregister from graph
	if err := s.unregisterResource("group", groupName); err != nil {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete group: %v", err)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete group"), nil
	}

	// Clean up related state
	s.state.Delete(attachKey)
	s.state.Delete(membersKey)
	s.state.Delete(inlineKey)

	return s.successResponse("DeleteGroup", EmptyResult{})
}

func (s *IAMService) updateGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:group:%s", groupName)
	var group XMLGroup
	if err := s.state.Get(stateKey, &group); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	newGroupName := getStringValue(params, "NewGroupName")
	newPath := getStringValue(params, "NewPath")

	// If renaming group
	if newGroupName != "" && newGroupName != groupName {
		newStateKey := fmt.Sprintf("iam:group:%s", newGroupName)
		if s.state.Exists(newStateKey) {
			return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Group with name %s already exists.", newGroupName)), nil
		}

		// Update group fields
		group.GroupName = newGroupName
		if newPath != "" {
			group.Path = newPath
		}
		group.Arn = fmt.Sprintf("arn:aws:iam::%s:group%s%s", defaultAccountID, group.Path, newGroupName)

		// Store with new key first (safer order - new key exists before old is deleted)
		if err := s.state.Set(newStateKey, &group); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to update group"), nil
		}

		// Only delete old key after new key is successfully created
		s.state.Delete(stateKey)

		// Migrate related state
		s.migrateGroupState(groupName, newGroupName)

		// Update graph registration
		s.unregisterResource("group", groupName)
		s.registerResource("group", newGroupName, map[string]string{
			"arn":  group.Arn,
			"path": group.Path,
		})
	} else if newPath != "" {
		// Just updating path
		group.Path = newPath
		group.Arn = fmt.Sprintf("arn:aws:iam::%s:group%s%s", defaultAccountID, group.Path, groupName)
		if err := s.state.Set(stateKey, &group); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to update group"), nil
		}
	}

	return s.successResponse("UpdateGroup", EmptyResult{})
}

func (s *IAMService) listGroups(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := getStringValue(params, "PathPrefix")

	keys, err := s.state.List("iam:group:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list groups"), nil
	}

	var groups []XMLGroupListItem
	for _, key := range keys {
		// Skip non-group keys (like iam:group-policies:, iam:group-members:, iam:group-inline-policies:)
		if strings.Contains(key, "group-policies:") || strings.Contains(key, "group-members:") || strings.Contains(key, "group-inline-policies:") {
			continue
		}
		var group XMLGroup
		if err := s.state.Get(key, &group); err == nil {
			if pathPrefix == "" || strings.HasPrefix(group.Path, pathPrefix) {
				groups = append(groups, groupToListItem(group))
			}
		}
	}

	result := ListGroupsResult{
		Groups:      groups,
		IsTruncated: false,
	}
	return s.successResponse("ListGroups", result)
}

// ============================================================================
// Helper Functions
// ============================================================================

// migrateGroupState migrates group-related state when a group is renamed.
// Note: This is best-effort migration for the in-memory emulator. In a production
// system, this would require transactional semantics. For testing purposes,
// partial migration failures are acceptable as state resets on restart.
func (s *IAMService) migrateGroupState(oldGroupName, newGroupName string) {
	// Migrate policy attachments
	oldAttachKey := fmt.Sprintf("iam:group-policies:%s", oldGroupName)
	newAttachKey := fmt.Sprintf("iam:group-policies:%s", newGroupName)
	var attachments GroupAttachments
	if err := s.state.Get(oldAttachKey, &attachments); err == nil {
		if err := s.state.Set(newAttachKey, &attachments); err == nil {
			s.state.Delete(oldAttachKey)
		}
	}

	// Migrate inline policies
	oldInlineKey := fmt.Sprintf("iam:group-inline-policies:%s", oldGroupName)
	newInlineKey := fmt.Sprintf("iam:group-inline-policies:%s", newGroupName)
	var inlinePolicies GroupInlinePolicies
	if err := s.state.Get(oldInlineKey, &inlinePolicies); err == nil {
		if err := s.state.Set(newInlineKey, &inlinePolicies); err == nil {
			s.state.Delete(oldInlineKey)
		}
	}

	// Migrate group members
	oldMembersKey := fmt.Sprintf("iam:group-members:%s", oldGroupName)
	newMembersKey := fmt.Sprintf("iam:group-members:%s", newGroupName)
	var members GroupMembers
	if err := s.state.Get(oldMembersKey, &members); err == nil {
		if err := s.state.Set(newMembersKey, &members); err == nil {
			s.state.Delete(oldMembersKey)

			// Update user-groups mappings to reflect the new group name
			for _, userName := range members.UserNames {
				userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
				var userGroups UserGroups
				if err := s.state.Get(userGroupsKey, &userGroups); err == nil {
					for i, gn := range userGroups.GroupNames {
						if gn == oldGroupName {
							userGroups.GroupNames[i] = newGroupName
							break
						}
					}
					s.state.Set(userGroupsKey, &userGroups)
				}
			}
		}
	}
}
