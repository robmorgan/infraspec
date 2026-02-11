package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// SAML Provider Operations
// ============================================================================

// createSAMLProvider creates a new SAML identity provider
func (s *IAMService) createSAMLProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	name := emulator.GetStringParam(params, "Name", "")
	if name == "" {
		return s.errorResponse(400, "ValidationError", "Name is required"), nil
	}

	// Validate name format (alphanumeric, ._- allowed, 1-128 chars)
	if !isValidSAMLProviderName(name) {
		return s.errorResponse(400, "ValidationError", "Invalid SAML provider name. Must be alphanumeric with ._- characters, 1-128 chars"), nil
	}

	samlMetadataDocument := emulator.GetStringParam(params, "SAMLMetadataDocument", "")
	if samlMetadataDocument == "" {
		return s.errorResponse(400, "ValidationError", "SAMLMetadataDocument is required"), nil
	}

	// Check if provider already exists
	stateKey := fmt.Sprintf("iam:saml-provider:%s", name)
	var existing SAMLProviderData
	if err := s.state.Get(stateKey, &existing); err == nil {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("SAML provider with name %s already exists", name)), nil
	}

	// Parse tags if provided
	tags := s.parseTags(params)

	arn := fmt.Sprintf("arn:aws:iam::%s:saml-provider/%s", defaultAccountID, name)
	now := time.Now().UTC()

	// Calculate ValidUntil from metadata (simplified - in real AWS this parses the XML)
	validUntil := now.AddDate(1, 0, 0) // Default to 1 year from now

	provider := SAMLProviderData{
		Name:                 name,
		Arn:                  arn,
		SAMLMetadataDocument: samlMetadataDocument,
		CreateDate:           now,
		ValidUntil:           validUntil,
		Tags:                 tags,
	}

	if err := s.state.Set(stateKey, provider); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to create SAML provider"), nil
	}

	result := CreateSAMLProviderResult{
		SAMLProviderArn: arn,
		Tags:            tags,
	}

	return s.successResponse("CreateSAMLProvider", result)
}

// getSAMLProvider retrieves details about a SAML provider
func (s *IAMService) getSAMLProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	samlProviderArn := emulator.GetStringParam(params, "SAMLProviderArn", "")
	if samlProviderArn == "" {
		return s.errorResponse(400, "ValidationError", "SAMLProviderArn is required"), nil
	}

	// Extract name from ARN
	name := extractSAMLProviderNameFromArn(samlProviderArn)
	if name == "" {
		return s.errorResponse(400, "ValidationError", "Invalid SAMLProviderArn format"), nil
	}

	stateKey := fmt.Sprintf("iam:saml-provider:%s", name)
	var provider SAMLProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SAML provider %s not found", samlProviderArn)), nil
	}

	result := GetSAMLProviderResult{
		CreateDate:           provider.CreateDate,
		ValidUntil:           provider.ValidUntil,
		SAMLMetadataDocument: provider.SAMLMetadataDocument,
		Tags:                 provider.Tags,
	}

	return s.successResponse("GetSAMLProvider", result)
}

// updateSAMLProvider updates an existing SAML provider's metadata document
func (s *IAMService) updateSAMLProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	samlProviderArn := emulator.GetStringParam(params, "SAMLProviderArn", "")
	if samlProviderArn == "" {
		return s.errorResponse(400, "ValidationError", "SAMLProviderArn is required"), nil
	}

	samlMetadataDocument := emulator.GetStringParam(params, "SAMLMetadataDocument", "")
	if samlMetadataDocument == "" {
		return s.errorResponse(400, "ValidationError", "SAMLMetadataDocument is required"), nil
	}

	// Extract name from ARN
	name := extractSAMLProviderNameFromArn(samlProviderArn)
	if name == "" {
		return s.errorResponse(400, "ValidationError", "Invalid SAMLProviderArn format"), nil
	}

	stateKey := fmt.Sprintf("iam:saml-provider:%s", name)
	var provider SAMLProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SAML provider %s not found", samlProviderArn)), nil
	}

	// Update the metadata document
	provider.SAMLMetadataDocument = samlMetadataDocument
	// Update ValidUntil (simplified)
	provider.ValidUntil = time.Now().UTC().AddDate(1, 0, 0)

	if err := s.state.Set(stateKey, provider); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to update SAML provider"), nil
	}

	result := UpdateSAMLProviderResult{
		SAMLProviderArn: samlProviderArn,
	}

	return s.successResponse("UpdateSAMLProvider", result)
}

// deleteSAMLProvider deletes a SAML provider
func (s *IAMService) deleteSAMLProvider(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	samlProviderArn := emulator.GetStringParam(params, "SAMLProviderArn", "")
	if samlProviderArn == "" {
		return s.errorResponse(400, "ValidationError", "SAMLProviderArn is required"), nil
	}

	// Extract name from ARN
	name := extractSAMLProviderNameFromArn(samlProviderArn)
	if name == "" {
		return s.errorResponse(400, "ValidationError", "Invalid SAMLProviderArn format"), nil
	}

	stateKey := fmt.Sprintf("iam:saml-provider:%s", name)
	var provider SAMLProviderData
	if err := s.state.Get(stateKey, &provider); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SAML provider %s not found", samlProviderArn)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete SAML provider"), nil
	}

	return s.successResponse("DeleteSAMLProvider", EmptyResult{})
}

// listSAMLProviders lists all SAML providers in the account
func (s *IAMService) listSAMLProviders(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// List all SAML providers
	keys, err := s.state.List("iam:saml-provider:")
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list SAML providers"), nil
	}

	var providers []SAMLProviderListEntry
	for _, key := range keys {
		var provider SAMLProviderData
		if err := s.state.Get(key, &provider); err != nil {
			continue
		}
		arn := provider.Arn
		validUntil := provider.ValidUntil
		createDate := provider.CreateDate
		providers = append(providers, SAMLProviderListEntry{
			Arn:        &arn,
			ValidUntil: &validUntil,
			CreateDate: &createDate,
		})
	}

	result := ListSAMLProvidersResult{
		SAMLProviderList: providers,
	}

	return s.successResponse("ListSAMLProviders", result)
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidSAMLProviderName validates the SAML provider name format
func isValidSAMLProviderName(name string) bool {
	if len(name) < 1 || len(name) > 128 {
		return false
	}
	// Must be alphanumeric with ._- allowed
	matched, _ := regexp.MatchString(`^[\w._-]+$`, name)
	return matched
}

// extractSAMLProviderNameFromArn extracts the provider name from an ARN
func extractSAMLProviderNameFromArn(arn string) string {
	// ARN format: arn:aws:iam::123456789012:saml-provider/ProviderName
	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// generateOIDCProviderArn generates an ARN for an OIDC provider based on URL
func generateOIDCProviderArn(url string) string {
	// Remove protocol prefix
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	return fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", defaultAccountID, url)
}

// generateOIDCProviderStateKey generates a state key from the URL (hashed for safety)
func generateOIDCProviderStateKey(url string) string {
	// Normalize and hash the URL for use as a state key
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("iam:oidc-provider:%s", hex.EncodeToString(hash[:8]))
}

// extractOIDCProviderUrlFromArn extracts the URL from an OIDC provider ARN
func extractOIDCProviderUrlFromArn(arn string) string {
	// ARN format: arn:aws:iam::123456789012:oidc-provider/provider.example.com
	parts := strings.SplitN(arn, ":oidc-provider/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
