package iam

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// Account Alias Operations
// ============================================================================

// createAccountAlias creates an account alias
func (s *IAMService) createAccountAlias(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	accountAlias := emulator.GetStringParam(params, "AccountAlias", "")
	if accountAlias == "" {
		return s.errorResponse(400, "ValidationError", "AccountAlias is required"), nil
	}

	// Validate alias format (3-63 lowercase alphanumeric or hyphens, must start with letter)
	if !isValidAccountAlias(accountAlias) {
		return s.errorResponse(400, "ValidationError", "Invalid account alias. Must be 3-63 lowercase alphanumeric characters or hyphens, starting with a letter"), nil
	}

	// Check if alias already exists (only one alias allowed per account)
	stateKey := "iam:account-alias"
	var existing string
	if err := s.state.Get(stateKey, &existing); err == nil && existing != "" {
		return s.errorResponse(409, "EntityAlreadyExists", "An account alias already exists"), nil
	}

	if err := s.state.Set(stateKey, accountAlias); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to create account alias"), nil
	}

	return s.successResponse("CreateAccountAlias", EmptyResult{})
}

// deleteAccountAlias deletes the account alias
func (s *IAMService) deleteAccountAlias(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	accountAlias := emulator.GetStringParam(params, "AccountAlias", "")
	if accountAlias == "" {
		return s.errorResponse(400, "ValidationError", "AccountAlias is required"), nil
	}

	stateKey := "iam:account-alias"
	var existing string
	if err := s.state.Get(stateKey, &existing); err != nil || existing == "" {
		return s.errorResponse(404, "NoSuchEntity", "Account alias does not exist"), nil
	}

	// Verify the alias matches
	if existing != accountAlias {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Account alias %s does not exist", accountAlias)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete account alias"), nil
	}

	return s.successResponse("DeleteAccountAlias", EmptyResult{})
}

// listAccountAliases lists account aliases (at most one per account)
func (s *IAMService) listAccountAliases(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	stateKey := "iam:account-alias"
	var alias string
	aliases := []string{}

	if err := s.state.Get(stateKey, &alias); err == nil && alias != "" {
		aliases = append(aliases, alias)
	}

	result := ListAccountAliasesResult{
		AccountAliases: aliases,
		IsTruncated:    false,
	}

	return s.successResponse("ListAccountAliases", result)
}

// ============================================================================
// Account Password Policy Operations
// ============================================================================

// updateAccountPasswordPolicy updates or creates the account password policy
func (s *IAMService) updateAccountPasswordPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Parse all password policy parameters
	policy := PasswordPolicyData{
		MinimumPasswordLength:      getIntParam(params, "MinimumPasswordLength", 6),
		RequireSymbols:             getBoolParam(params, "RequireSymbols", false),
		RequireNumbers:             getBoolParam(params, "RequireNumbers", false),
		RequireUppercaseCharacters: getBoolParam(params, "RequireUppercaseCharacters", false),
		RequireLowercaseCharacters: getBoolParam(params, "RequireLowercaseCharacters", false),
		AllowUsersToChangePassword: getBoolParam(params, "AllowUsersToChangePassword", false),
		ExpirePasswords:            getBoolParam(params, "ExpirePasswords", false),
		MaxPasswordAge:             getIntParam(params, "MaxPasswordAge", 0),
		PasswordReusePrevention:    getIntParam(params, "PasswordReusePrevention", 0),
		HardExpiry:                 getBoolParam(params, "HardExpiry", false),
	}

	// Validate password length (6-128)
	if policy.MinimumPasswordLength < 6 || policy.MinimumPasswordLength > 128 {
		return s.errorResponse(400, "ValidationError", "MinimumPasswordLength must be between 6 and 128"), nil
	}

	// Validate max password age (1-1095 if set)
	if policy.MaxPasswordAge != 0 && (policy.MaxPasswordAge < 1 || policy.MaxPasswordAge > 1095) {
		return s.errorResponse(400, "ValidationError", "MaxPasswordAge must be between 1 and 1095"), nil
	}

	// Validate password reuse prevention (1-24 if set)
	if policy.PasswordReusePrevention != 0 && (policy.PasswordReusePrevention < 1 || policy.PasswordReusePrevention > 24) {
		return s.errorResponse(400, "ValidationError", "PasswordReusePrevention must be between 1 and 24"), nil
	}

	stateKey := "iam:password-policy"
	if err := s.state.Set(stateKey, policy); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to update password policy"), nil
	}

	return s.successResponse("UpdateAccountPasswordPolicy", EmptyResult{})
}

// getAccountPasswordPolicy retrieves the account password policy
func (s *IAMService) getAccountPasswordPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	stateKey := "iam:password-policy"
	var policy PasswordPolicyData
	if err := s.state.Get(stateKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", "Password policy has not been set"), nil
	}

	result := GetAccountPasswordPolicyResult{
		PasswordPolicy: XMLPasswordPolicy{
			MinimumPasswordLength:      policy.MinimumPasswordLength,
			RequireSymbols:             policy.RequireSymbols,
			RequireNumbers:             policy.RequireNumbers,
			RequireUppercaseCharacters: policy.RequireUppercaseCharacters,
			RequireLowercaseCharacters: policy.RequireLowercaseCharacters,
			AllowUsersToChangePassword: policy.AllowUsersToChangePassword,
			ExpirePasswords:            policy.ExpirePasswords,
			MaxPasswordAge:             policy.MaxPasswordAge,
			PasswordReusePrevention:    policy.PasswordReusePrevention,
			HardExpiry:                 policy.HardExpiry,
		},
	}

	return s.successResponse("GetAccountPasswordPolicy", result)
}

// deleteAccountPasswordPolicy deletes the account password policy
func (s *IAMService) deleteAccountPasswordPolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	stateKey := "iam:password-policy"
	var policy PasswordPolicyData
	if err := s.state.Get(stateKey, &policy); err != nil {
		return s.errorResponse(404, "NoSuchEntity", "Password policy has not been set"), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete password policy"), nil
	}

	return s.successResponse("DeleteAccountPasswordPolicy", EmptyResult{})
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidAccountAlias validates an account alias format
func isValidAccountAlias(alias string) bool {
	if len(alias) < 3 || len(alias) > 63 {
		return false
	}
	// Must start with a letter, contain only lowercase alphanumeric or hyphens
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9-]*$`, alias)
	return matched
}

// getIntParam gets an integer parameter with a default value
func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// getBoolParam gets a boolean parameter with a default value
func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "True" || v == "1"
		}
	}
	return defaultValue
}
