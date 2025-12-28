package iam

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// generateIAMId generates an AWS-style IAM resource ID with the given prefix
func generateIAMId(prefix string) string {
	// AWS IDs are 21 characters: 4-char prefix + 17 alphanumeric chars
	b := make([]byte, 13) // Will give us ~17 chars after base64
	rand.Read(b)
	encoded := base64.RawURLEncoding.EncodeToString(b)
	// Replace URL-safe chars with uppercase letters to match AWS style
	encoded = strings.ReplaceAll(encoded, "-", "X")
	encoded = strings.ReplaceAll(encoded, "_", "Y")
	encoded = strings.ToUpper(encoded)
	if len(encoded) > 17 {
		encoded = encoded[:17]
	}
	return prefix + encoded
}

// extractPolicyNameFromArn extracts the policy name from an ARN like
// arn:aws:iam::123456789012:policy/path/PolicyName
func extractPolicyNameFromArn(arn string) string {
	parts := strings.Split(arn, ":policy")
	if len(parts) < 2 {
		return ""
	}
	pathAndName := parts[1]
	// Remove leading slash and get the last part after any path
	pathAndName = strings.TrimPrefix(pathAndName, "/")
	nameParts := strings.Split(pathAndName, "/")
	if len(nameParts) == 0 {
		return ""
	}
	return nameParts[len(nameParts)-1]
}

func getStringValue(params map[string]interface{}, key string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return ""
}

func getInt32Value(params map[string]interface{}, key string, defaultValue int32) int32 {
	if val, ok := params[key].(float64); ok {
		return int32(val)
	}
	if val, ok := params[key].(int); ok {
		return int32(val)
	}
	if val, ok := params[key].(int32); ok {
		return val
	}
	if val, ok := params[key].(string); ok {
		var parsed int32
		if _, err := fmt.Sscanf(val, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// roleToListItem converts an XMLRole to XMLRoleListItem for list responses
func roleToListItem(r XMLRole) XMLRoleListItem {
	return XMLRoleListItem{
		RoleName:                 r.RoleName,
		RoleId:                   r.RoleId,
		Arn:                      r.Arn,
		Path:                     r.Path,
		AssumeRolePolicyDocument: r.AssumeRolePolicyDocument,
		Description:              r.Description,
		MaxSessionDuration:       r.MaxSessionDuration,
		CreateDate:               r.CreateDate,
		Tags:                     r.Tags,
	}
}

// instanceProfileToListItem converts an XMLInstanceProfile to XMLInstanceProfileListItem for list responses
func instanceProfileToListItem(p XMLInstanceProfile) XMLInstanceProfileListItem {
	return XMLInstanceProfileListItem{
		InstanceProfileName: p.InstanceProfileName,
		InstanceProfileId:   p.InstanceProfileId,
		Arn:                 p.Arn,
		Path:                p.Path,
		CreateDate:          p.CreateDate,
		Roles:               p.Roles,
		Tags:                p.Tags,
	}
}

// userToListItem converts an XMLUser to XMLUserListItem for list responses
func userToListItem(u XMLUser) XMLUserListItem {
	return XMLUserListItem{
		UserName:         u.UserName,
		UserId:           u.UserId,
		Arn:              u.Arn,
		Path:             u.Path,
		CreateDate:       u.CreateDate,
		PasswordLastUsed: u.PasswordLastUsed,
		Tags:             u.Tags,
	}
}

// groupToListItem converts an XMLGroup to XMLGroupListItem for list responses
func groupToListItem(g XMLGroup) XMLGroupListItem {
	return XMLGroupListItem{
		GroupName:  g.GroupName,
		GroupId:    g.GroupId,
		Arn:        g.Arn,
		Path:       g.Path,
		CreateDate: g.CreateDate,
	}
}
