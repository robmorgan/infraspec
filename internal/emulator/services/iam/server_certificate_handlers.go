package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// Server Certificate Operations
// ============================================================================

// uploadServerCertificate uploads a new server certificate
func (s *IAMService) uploadServerCertificate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	serverCertificateName := emulator.GetStringParam(params, "ServerCertificateName", "")
	if serverCertificateName == "" {
		return s.errorResponse(400, "ValidationError", "ServerCertificateName is required"), nil
	}

	// Validate name format
	if !isValidServerCertificateName(serverCertificateName) {
		return s.errorResponse(400, "ValidationError", "Invalid server certificate name"), nil
	}

	certificateBody := emulator.GetStringParam(params, "CertificateBody", "")
	if certificateBody == "" {
		return s.errorResponse(400, "ValidationError", "CertificateBody is required"), nil
	}

	privateKey := emulator.GetStringParam(params, "PrivateKey", "")
	if privateKey == "" {
		return s.errorResponse(400, "ValidationError", "PrivateKey is required"), nil
	}

	path := emulator.GetStringParam(params, "Path", "/")
	certificateChain := emulator.GetStringParam(params, "CertificateChain", "")

	// Check if certificate already exists
	stateKey := fmt.Sprintf("iam:server-certificate:%s", serverCertificateName)
	var existing ServerCertificateData
	if err := s.state.Get(stateKey, &existing); err == nil {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Server certificate %s already exists", serverCertificateName)), nil
	}

	// Generate certificate ID
	certId := generateServerCertificateId()
	arn := fmt.Sprintf("arn:aws:iam::%s:server-certificate%s%s", defaultAccountID, path, serverCertificateName)
	now := time.Now().UTC()

	// Parse expiration from certificate (simplified - in real AWS this parses the X.509 cert)
	expiration := now.AddDate(1, 0, 0) // Default to 1 year from now

	// Parse tags if provided
	tags := s.parseTags(params)

	cert := ServerCertificateData{
		ServerCertificateName: serverCertificateName,
		ServerCertificateId:   certId,
		Arn:                   arn,
		Path:                  path,
		CertificateBody:       certificateBody,
		CertificateChain:      certificateChain,
		PrivateKey:            privateKey,
		UploadDate:            now,
		Expiration:            expiration,
		Tags:                  tags,
	}

	if err := s.state.Set(stateKey, cert); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to upload server certificate"), nil
	}

	result := UploadServerCertificateResult{
		ServerCertificateMetadata: XMLServerCertificateMetadata{
			ServerCertificateName: serverCertificateName,
			ServerCertificateId:   certId,
			Arn:                   arn,
			Path:                  path,
			UploadDate:            now,
			Expiration:            expiration,
		},
		Tags: tags,
	}

	return s.successResponse("UploadServerCertificate", result)
}

// getServerCertificate retrieves a server certificate
func (s *IAMService) getServerCertificate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	serverCertificateName := emulator.GetStringParam(params, "ServerCertificateName", "")
	if serverCertificateName == "" {
		return s.errorResponse(400, "ValidationError", "ServerCertificateName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:server-certificate:%s", serverCertificateName)
	var cert ServerCertificateData
	if err := s.state.Get(stateKey, &cert); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Server certificate %s not found", serverCertificateName)), nil
	}

	result := GetServerCertificateResult{
		ServerCertificate: XMLServerCertificate{
			ServerCertificateMetadata: XMLServerCertificateMetadata{
				ServerCertificateName: cert.ServerCertificateName,
				ServerCertificateId:   cert.ServerCertificateId,
				Arn:                   cert.Arn,
				Path:                  cert.Path,
				UploadDate:            cert.UploadDate,
				Expiration:            cert.Expiration,
			},
			CertificateBody:  cert.CertificateBody,
			CertificateChain: cert.CertificateChain,
			Tags:             cert.Tags,
		},
	}

	return s.successResponse("GetServerCertificate", result)
}

// deleteServerCertificate deletes a server certificate
func (s *IAMService) deleteServerCertificate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	serverCertificateName := emulator.GetStringParam(params, "ServerCertificateName", "")
	if serverCertificateName == "" {
		return s.errorResponse(400, "ValidationError", "ServerCertificateName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:server-certificate:%s", serverCertificateName)
	var cert ServerCertificateData
	if err := s.state.Get(stateKey, &cert); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Server certificate %s not found", serverCertificateName)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete server certificate"), nil
	}

	return s.successResponse("DeleteServerCertificate", EmptyResult{})
}

// listServerCertificates lists all server certificates
func (s *IAMService) listServerCertificates(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := emulator.GetStringParam(params, "PathPrefix", "/")

	keys, err := s.state.List("iam:server-certificate:")
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list server certificates"), nil
	}

	var certs []ServerCertificateMetadataListItem
	for _, key := range keys {
		var cert ServerCertificateData
		if err := s.state.Get(key, &cert); err != nil {
			continue
		}

		// Filter by path prefix
		if !strings.HasPrefix(cert.Path, pathPrefix) {
			continue
		}

		certs = append(certs, ServerCertificateMetadataListItem{
			ServerCertificateName: cert.ServerCertificateName,
			ServerCertificateId:   cert.ServerCertificateId,
			Arn:                   cert.Arn,
			Path:                  cert.Path,
			UploadDate:            cert.UploadDate,
			Expiration:            cert.Expiration,
		})
	}

	result := ListServerCertificatesResult{
		ServerCertificateMetadataList: certs,
		IsTruncated:                   false,
	}

	return s.successResponse("ListServerCertificates", result)
}

// updateServerCertificate updates a server certificate (rename or change path)
func (s *IAMService) updateServerCertificate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	serverCertificateName := emulator.GetStringParam(params, "ServerCertificateName", "")
	if serverCertificateName == "" {
		return s.errorResponse(400, "ValidationError", "ServerCertificateName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:server-certificate:%s", serverCertificateName)
	var cert ServerCertificateData
	if err := s.state.Get(stateKey, &cert); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Server certificate %s not found", serverCertificateName)), nil
	}

	newName := emulator.GetStringParam(params, "NewServerCertificateName", "")
	newPath := emulator.GetStringParam(params, "NewPath", "")

	// Update name if provided
	if newName != "" && newName != serverCertificateName {
		if !isValidServerCertificateName(newName) {
			return s.errorResponse(400, "ValidationError", "Invalid new server certificate name"), nil
		}

		// Check if new name already exists
		newStateKey := fmt.Sprintf("iam:server-certificate:%s", newName)
		var existing ServerCertificateData
		if err := s.state.Get(newStateKey, &existing); err == nil {
			return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Server certificate %s already exists", newName)), nil
		}

		cert.ServerCertificateName = newName
	}

	// Update path if provided
	if newPath != "" {
		cert.Path = newPath
	}

	// Update ARN
	cert.Arn = fmt.Sprintf("arn:aws:iam::%s:server-certificate%s%s", defaultAccountID, cert.Path, cert.ServerCertificateName)

	// Handle rename: create new key first, then delete old (safer order)
	if newName != "" && newName != serverCertificateName {
		newStateKey := fmt.Sprintf("iam:server-certificate:%s", newName)
		if err := s.state.Set(newStateKey, cert); err != nil {
			return s.errorResponse(500, "ServiceFailure", "Failed to update server certificate"), nil
		}
		// Only delete old key after new key is successfully created
		s.state.Delete(stateKey)
	} else {
		// Just updating path or other fields, use same key
		if err := s.state.Set(stateKey, cert); err != nil {
			return s.errorResponse(500, "ServiceFailure", "Failed to update server certificate"), nil
		}
	}

	return s.successResponse("UpdateServerCertificate", EmptyResult{})
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidServerCertificateName validates the server certificate name format
func isValidServerCertificateName(name string) bool {
	if len(name) < 1 || len(name) > 128 {
		return false
	}
	// Must be alphanumeric with +=,.@_- allowed
	matched, _ := regexp.MatchString(`^[\w+=,.@-]+$`, name)
	return matched
}

// generateServerCertificateId generates a unique server certificate ID
func generateServerCertificateId() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "ASCA" + strings.ToUpper(hex.EncodeToString(b))[:16]
}
