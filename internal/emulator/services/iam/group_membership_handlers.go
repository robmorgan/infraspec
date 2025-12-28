package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// ============================================================================
// Group Membership Operations
// ============================================================================

func (s *IAMService) addUserToGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Add user to group members
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		members = GroupMembers{UserNames: []string{}}
	}

	// Check if already a member
	for _, un := range members.UserNames {
		if un == userName {
			// Already a member, idempotent success
			return s.successResponse("AddUserToGroup", EmptyResult{})
		}
	}

	members.UserNames = append(members.UserNames, userName)
	if err := s.state.Set(membersKey, &members); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to add user to group"), nil
	}

	// Add group to user's groups list
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err != nil {
		userGroups = UserGroups{GroupNames: []string{}}
	}
	userGroups.GroupNames = append(userGroups.GroupNames, groupName)
	if err := s.state.Set(userGroupsKey, &userGroups); err != nil {
		// Rollback member addition
		members.UserNames = members.UserNames[:len(members.UserNames)-1]
		s.state.Set(membersKey, &members)
		return s.errorResponse(500, "InternalFailure", "Failed to update user groups"), nil
	}

	// Add relationship in graph: user -> group
	if err := s.addRelationship("user", userName, "group", groupName, graph.RelAssociatedWith); err != nil {
		// Non-critical, just log
	}

	return s.successResponse("AddUserToGroup", EmptyResult{})
}

func (s *IAMService) removeUserFromGroup(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	groupName := getStringValue(params, "GroupName")
	if groupName == "" {
		return s.errorResponse(400, "ValidationError", "GroupName is required"), nil
	}

	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify group exists
	groupKey := fmt.Sprintf("iam:group:%s", groupName)
	if !s.state.Exists(groupKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The group with name %s cannot be found.", groupName)), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get group members
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user %s is not in group %s.", userName, groupName)), nil
	}

	// Find and remove the user
	found := false
	newUserNames := make([]string, 0, len(members.UserNames))
	for _, un := range members.UserNames {
		if un == userName {
			found = true
		} else {
			newUserNames = append(newUserNames, un)
		}
	}

	if !found {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user %s is not in group %s.", userName, groupName)), nil
	}

	members.UserNames = newUserNames
	if err := s.state.Set(membersKey, &members); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to remove user from group"), nil
	}

	// Remove group from user's groups list
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err == nil {
		newGroupNames := make([]string, 0, len(userGroups.GroupNames))
		for _, gn := range userGroups.GroupNames {
			if gn != groupName {
				newGroupNames = append(newGroupNames, gn)
			}
		}
		userGroups.GroupNames = newGroupNames
		s.state.Set(userGroupsKey, &userGroups)
	}

	// Remove relationship in graph
	s.removeRelationship("user", userName, "group", groupName, graph.RelAssociatedWith)

	return s.successResponse("RemoveUserFromGroup", EmptyResult{})
}

func (s *IAMService) listGroupsForUser(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get user's groups
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err != nil {
		userGroups = UserGroups{GroupNames: []string{}}
	}

	// Get group details for each group
	var groups []XMLGroupListItem
	for _, groupName := range userGroups.GroupNames {
		groupKey := fmt.Sprintf("iam:group:%s", groupName)
		var group XMLGroup
		if err := s.state.Get(groupKey, &group); err == nil {
			groups = append(groups, groupToListItem(group))
		}
	}

	result := ListGroupsForUserResult{
		Groups:      groups,
		IsTruncated: false,
	}
	return s.successResponse("ListGroupsForUser", result)
}

// cleanupUserGroupMemberships removes a user from all groups when the user is deleted
// This should be called from deleteUser before deleting the user
func (s *IAMService) cleanupUserGroupMemberships(userName string) {
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err != nil {
		return // No groups to clean up
	}

	// Remove user from each group
	for _, groupName := range userGroups.GroupNames {
		membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
		var members GroupMembers
		if err := s.state.Get(membersKey, &members); err == nil {
			newUserNames := make([]string, 0, len(members.UserNames))
			for _, un := range members.UserNames {
				if un != userName {
					newUserNames = append(newUserNames, un)
				}
			}
			members.UserNames = newUserNames
			s.state.Set(membersKey, &members)
		}
	}

	// Delete the user's groups list
	s.state.Delete(userGroupsKey)
}

// getGroupMembersCount returns the number of members in a group
func (s *IAMService) getGroupMembersCount(groupName string) int {
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		return 0
	}
	return len(members.UserNames)
}

// isUserInGroup checks if a user is a member of a group
func (s *IAMService) isUserInGroup(userName, groupName string) bool {
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		return false
	}
	for _, un := range members.UserNames {
		if un == userName {
			return true
		}
	}
	return false
}

// listGroupMemberNames returns a list of user names in a group
func (s *IAMService) listGroupMemberNames(groupName string) []string {
	membersKey := fmt.Sprintf("iam:group-members:%s", groupName)
	var members GroupMembers
	if err := s.state.Get(membersKey, &members); err != nil {
		return []string{}
	}
	return members.UserNames
}

// listUserGroupNames returns a list of group names for a user
func (s *IAMService) listUserGroupNames(userName string) []string {
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err != nil {
		return []string{}
	}
	return userGroups.GroupNames
}

// hasGroupMemberships checks if a user has any group memberships
func (s *IAMService) hasGroupMemberships(userName string) bool {
	userGroupsKey := fmt.Sprintf("iam:user-groups:%s", userName)
	var userGroups UserGroups
	if err := s.state.Get(userGroupsKey, &userGroups); err != nil {
		return false
	}
	return len(userGroups.GroupNames) > 0
}

// getAllGroupNames returns all group names matching an optional path prefix
func (s *IAMService) getAllGroupNames(pathPrefix string) []string {
	keys, err := s.state.List("iam:group:")
	if err != nil {
		return []string{}
	}

	var groupNames []string
	for _, key := range keys {
		// Skip non-group keys
		if strings.Contains(key, "group-policies:") || strings.Contains(key, "group-members:") || strings.Contains(key, "group-inline-policies:") {
			continue
		}
		var group XMLGroup
		if err := s.state.Get(key, &group); err == nil {
			if pathPrefix == "" || strings.HasPrefix(group.Path, pathPrefix) {
				groupNames = append(groupNames, group.GroupName)
			}
		}
	}
	return groupNames
}
