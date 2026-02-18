package iam

import (
	"context"
	"fmt"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// User Login Profile Operations
// ============================================================================

func (s *IAMService) createLoginProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	password := getStringValue(params, "Password")
	if password == "" {
		return s.errorResponse(400, "ValidationError", "Password is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Check if login profile already exists
	loginKey := fmt.Sprintf("iam:user-login-profile:%s", userName)
	if s.state.Exists(loginKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Login Profile for user %s already exists.", userName)), nil
	}

	passwordResetRequired := getBoolValue(params, "PasswordResetRequired", false)

	now := time.Now().UTC()
	loginProfile := UserLoginProfile{
		PasswordHash:          password, // In real AWS, this would be hashed
		CreateDate:            now,
		PasswordResetRequired: passwordResetRequired,
	}

	if err := s.state.Set(loginKey, &loginProfile); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to create login profile"), nil
	}

	result := CreateLoginProfileResult{
		LoginProfile: XMLLoginProfile{
			UserName:              userName,
			CreateDate:            now,
			PasswordResetRequired: passwordResetRequired,
		},
	}
	return s.successResponse("CreateLoginProfile", result)
}

func (s *IAMService) getLoginProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	loginKey := fmt.Sprintf("iam:user-login-profile:%s", userName)
	var loginProfile UserLoginProfile
	if err := s.state.Get(loginKey, &loginProfile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Login Profile for user %s cannot be found.", userName)), nil
	}

	result := GetLoginProfileResult{
		LoginProfile: XMLLoginProfile{
			UserName:              userName,
			CreateDate:            loginProfile.CreateDate,
			PasswordResetRequired: loginProfile.PasswordResetRequired,
		},
	}
	return s.successResponse("GetLoginProfile", result)
}

func (s *IAMService) updateLoginProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	loginKey := fmt.Sprintf("iam:user-login-profile:%s", userName)
	var loginProfile UserLoginProfile
	if err := s.state.Get(loginKey, &loginProfile); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Login Profile for user %s cannot be found.", userName)), nil
	}

	// Update password if provided
	password := getStringValue(params, "Password")
	if password != "" {
		loginProfile.PasswordHash = password // In real AWS, this would be hashed
	}

	// Update PasswordResetRequired if provided
	if val, ok := params["PasswordResetRequired"]; ok {
		if boolVal, ok := val.(bool); ok {
			loginProfile.PasswordResetRequired = boolVal
		} else if strVal, ok := val.(string); ok {
			loginProfile.PasswordResetRequired = strVal == "true"
		}
	}

	if err := s.state.Set(loginKey, &loginProfile); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update login profile"), nil
	}

	return s.successResponse("UpdateLoginProfile", EmptyResult{})
}

func (s *IAMService) deleteLoginProfile(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	loginKey := fmt.Sprintf("iam:user-login-profile:%s", userName)
	if !s.state.Exists(loginKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Login Profile for user %s cannot be found.", userName)), nil
	}

	if err := s.state.Delete(loginKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete login profile"), nil
	}

	return s.successResponse("DeleteLoginProfile", EmptyResult{})
}

// getBoolValue extracts a boolean value from params
func getBoolValue(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	if val, ok := params[key].(string); ok {
		return val == "true"
	}
	return defaultValue
}
