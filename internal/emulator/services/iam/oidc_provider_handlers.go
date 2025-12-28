package iam

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// OpenID Connect Provider Operations
// ============================================================================

// createOpenIDConnectProvider creates a new OIDC identity provider
func (s *IAMService) createOpenIDConnectProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	url := emulator.GetStringParam(params, "Url", "")
	if url == "" {
		return s.errorResponse(400, "ValidationError", "Url is required"), nil
	}

	// Validate URL format (must start with https://)
	if !strings.HasPrefix(url, "https://") {
		return s.errorResponse(400, "ValidationError", "Url must start with https://"), nil
	}

	// URL must be 1-255 characters
	if len(url) > 255 {
		return s.errorResponse(400, "ValidationError", "Url must not exceed 255 characters"), nil
	}

	// Get thumbprint list (at least one thumbprint required)
	thumbprintList := parseStringList(params, "ThumbprintList")
	if len(thumbprintList) == 0 {
		return s.errorResponse(400, "ValidationError", "At least one thumbprint is required"), nil
	}

	// Validate thumbprints (40 hex characters each)
	for _, tp := range thumbprintList {
		if !isValidThumbprint(tp) {
			return s.errorResponse(400, "ValidationError", fmt.Sprintf("Invalid thumbprint: %s. Must be 40 hex characters", tp)), nil
		}
	}

	// Check if provider already exists
	stateKey := generateOIDCProviderStateKey(url)
	var existing OIDCProviderData
	if err := s.state.Get(stateKey, &existing); err == nil {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("OIDC provider with URL %s already exists", url)), nil
	}

	// Parse optional client ID list
	clientIDList := parseStringList(params, "ClientIDList")

	// Parse tags if provided
	tags := s.parseTags(params)

	arn := generateOIDCProviderArn(url)
	now := time.Now().UTC()

	provider := OIDCProviderData{
		Url:            url,
		Arn:            arn,
		CreateDate:     now,
		ThumbprintList: thumbprintList,
		ClientIDList:   clientIDList,
		Tags:           tags,
	}

	if err := s.state.Set(stateKey, provider); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to create OIDC provider"), nil
	}

	// Also store ARN -> URL mapping for lookups by ARN
	arnKey := fmt.Sprintf("iam:oidc-provider-arn:%s", arn)
	if err := s.state.Set(arnKey, url); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to create OIDC provider mapping"), nil
	}

	result := CreateOpenIDConnectProviderResult{
		OpenIDConnectProviderArn: arn,
		Tags:                     tags,
	}

	return s.successResponse("CreateOpenIDConnectProvider", result)
}

// getOpenIDConnectProvider retrieves details about an OIDC provider
func (s *IAMService) getOpenIDConnectProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	arn := emulator.GetStringParam(params, "OpenIDConnectProviderArn", "")
	if arn == "" {
		return s.errorResponse(400, "ValidationError", "OpenIDConnectProviderArn is required"), nil
	}

	// Extract URL from ARN and find the provider
	url := extractOIDCProviderUrlFromArn(arn)
	if url == "" {
		return s.errorResponse(400, "ValidationError", "Invalid OpenIDConnectProviderArn format"), nil
	}

	stateKey := generateOIDCProviderStateKey("https://" + url)
	var provider OIDCProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("OIDC provider %s not found", arn)), nil
	}

	result := GetOpenIDConnectProviderResult{
		Url:            provider.Url,
		CreateDate:     provider.CreateDate,
		ThumbprintList: provider.ThumbprintList,
		ClientIDList:   provider.ClientIDList,
		Tags:           provider.Tags,
	}

	return s.successResponse("GetOpenIDConnectProvider", result)
}

// deleteOpenIDConnectProvider deletes an OIDC provider
func (s *IAMService) deleteOpenIDConnectProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	arn := emulator.GetStringParam(params, "OpenIDConnectProviderArn", "")
	if arn == "" {
		return s.errorResponse(400, "ValidationError", "OpenIDConnectProviderArn is required"), nil
	}

	// Extract URL from ARN and find the provider
	url := extractOIDCProviderUrlFromArn(arn)
	if url == "" {
		return s.errorResponse(400, "ValidationError", "Invalid OpenIDConnectProviderArn format"), nil
	}

	stateKey := generateOIDCProviderStateKey("https://" + url)
	var provider OIDCProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("OIDC provider %s not found", arn)), nil
	}

	// Delete both the provider and the ARN mapping
	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete OIDC provider"), nil
	}

	arnKey := fmt.Sprintf("iam:oidc-provider-arn:%s", arn)
	s.state.Delete(arnKey) // Ignore error for mapping cleanup

	return s.successResponse("DeleteOpenIDConnectProvider", EmptyResult{})
}

// listOpenIDConnectProviders lists all OIDC providers in the account
func (s *IAMService) listOpenIDConnectProviders(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// List all OIDC providers (using the hash-based keys)
	keys, err := s.state.List("iam:oidc-provider:")
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list OIDC providers"), nil
	}

	var providers []OpenIDConnectProviderListEntry
	for _, key := range keys {
		// Skip the ARN mapping keys
		if strings.Contains(key, "oidc-provider-arn:") {
			continue
		}

		var provider OIDCProviderData
		if err := s.state.Get(key, &provider); err != nil {
			continue
		}
		arn := provider.Arn
		providers = append(providers, OpenIDConnectProviderListEntry{
			Arn: &arn,
		})
	}

	result := ListOpenIDConnectProvidersResult{
		OpenIDConnectProviderList: providers,
	}

	return s.successResponse("ListOpenIDConnectProviders", result)
}

// updateOpenIDConnectProviderThumbprint updates the thumbprint list for an OIDC provider
func (s *IAMService) updateOpenIDConnectProviderThumbprint(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	arn := emulator.GetStringParam(params, "OpenIDConnectProviderArn", "")
	if arn == "" {
		return s.errorResponse(400, "ValidationError", "OpenIDConnectProviderArn is required"), nil
	}

	// Get thumbprint list
	thumbprintList := parseStringList(params, "ThumbprintList")
	if len(thumbprintList) == 0 {
		return s.errorResponse(400, "ValidationError", "ThumbprintList is required"), nil
	}

	// Validate thumbprints
	for _, tp := range thumbprintList {
		if !isValidThumbprint(tp) {
			return s.errorResponse(400, "ValidationError", fmt.Sprintf("Invalid thumbprint: %s. Must be 40 hex characters", tp)), nil
		}
	}

	// Extract URL from ARN and find the provider
	url := extractOIDCProviderUrlFromArn(arn)
	if url == "" {
		return s.errorResponse(400, "ValidationError", "Invalid OpenIDConnectProviderArn format"), nil
	}

	stateKey := generateOIDCProviderStateKey("https://" + url)
	var provider OIDCProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("OIDC provider %s not found", arn)), nil
	}

	// Update thumbprint list
	provider.ThumbprintList = thumbprintList

	if err := s.state.Set(stateKey, provider); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to update OIDC provider"), nil
	}

	return s.successResponse("UpdateOpenIDConnectProviderThumbprint", EmptyResult{})
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidThumbprint validates that a thumbprint is 40 hex characters
func isValidThumbprint(thumbprint string) bool {
	if len(thumbprint) != 40 {
		return false
	}
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, thumbprint)
	return matched
}

// parseStringList parses a list of strings from various parameter formats
func parseStringList(params map[string]interface{}, key string) []string {
	var result []string

	// Try direct list
	if list, ok := params[key].([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	// Try string slice
	if list, ok := params[key].([]string); ok {
		return list
	}

	// Try AWS-style indexed parameters: Key.member.1, Key.member.2, etc.
	for i := 1; i <= 100; i++ {
		memberKey := fmt.Sprintf("%s.member.%d", key, i)
		if val, ok := params[memberKey].(string); ok {
			result = append(result, val)
		} else {
			break
		}
	}

	return result
}
