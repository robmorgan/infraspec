package iam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// SSH Public Key Operations
// ============================================================================

// uploadSSHPublicKey uploads a new SSH public key for a user
func (s *IAMService) uploadSSHPublicKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	sshPublicKeyBody := emulator.GetStringParam(params, "SSHPublicKeyBody", "")
	if sshPublicKeyBody == "" {
		return s.errorResponse(400, "ValidationError", "SSHPublicKeyBody is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	// Generate SSH public key ID
	keyId := generateSSHPublicKeyId()

	// Generate fingerprint from key body (simplified - real AWS uses MD5 of key)
	fingerprint := generateSSHKeyFingerprint(sshPublicKeyBody)

	now := time.Now().UTC()
	sshKey := SSHPublicKeyData{
		UserName:         userName,
		SSHPublicKeyId:   keyId,
		Fingerprint:      fingerprint,
		SSHPublicKeyBody: sshPublicKeyBody,
		Status:           "Active",
		UploadDate:       now,
	}

	stateKey := fmt.Sprintf("iam:ssh-public-key:%s:%s", userName, keyId)
	if err := s.state.Set(stateKey, sshKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to upload SSH public key"), nil
	}

	result := UploadSSHPublicKeyResult{
		SSHPublicKey: XMLSSHPublicKey{
			UserName:         userName,
			SSHPublicKeyId:   keyId,
			Fingerprint:      fingerprint,
			SSHPublicKeyBody: sshPublicKeyBody,
			Status:           "Active",
			UploadDate:       now,
		},
	}

	return s.successResponse("UploadSSHPublicKey", result)
}

// getSSHPublicKey retrieves an SSH public key
func (s *IAMService) getSSHPublicKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	sshPublicKeyId := emulator.GetStringParam(params, "SSHPublicKeyId", "")
	if sshPublicKeyId == "" {
		return s.errorResponse(400, "ValidationError", "SSHPublicKeyId is required"), nil
	}

	encoding := emulator.GetStringParam(params, "Encoding", "SSH")

	stateKey := fmt.Sprintf("iam:ssh-public-key:%s:%s", userName, sshPublicKeyId)
	var sshKey SSHPublicKeyData
	if err := s.state.Get(stateKey, &sshKey); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SSH public key %s not found for user %s", sshPublicKeyId, userName)), nil
	}

	// Note: encoding parameter (SSH or PEM) would affect key format in real AWS
	// For emulation, we just return the key as-is
	_ = encoding

	result := GetSSHPublicKeyResult{
		SSHPublicKey: XMLSSHPublicKey{
			UserName:         sshKey.UserName,
			SSHPublicKeyId:   sshKey.SSHPublicKeyId,
			Fingerprint:      sshKey.Fingerprint,
			SSHPublicKeyBody: sshKey.SSHPublicKeyBody,
			Status:           sshKey.Status,
			UploadDate:       sshKey.UploadDate,
		},
	}

	return s.successResponse("GetSSHPublicKey", result)
}

// updateSSHPublicKey updates the status of an SSH public key
func (s *IAMService) updateSSHPublicKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	sshPublicKeyId := emulator.GetStringParam(params, "SSHPublicKeyId", "")
	if sshPublicKeyId == "" {
		return s.errorResponse(400, "ValidationError", "SSHPublicKeyId is required"), nil
	}

	status := emulator.GetStringParam(params, "Status", "")
	if status == "" {
		return s.errorResponse(400, "ValidationError", "Status is required"), nil
	}

	// Validate status
	if status != "Active" && status != "Inactive" {
		return s.errorResponse(400, "ValidationError", "Status must be Active or Inactive"), nil
	}

	stateKey := fmt.Sprintf("iam:ssh-public-key:%s:%s", userName, sshPublicKeyId)
	var sshKey SSHPublicKeyData
	if err := s.state.Get(stateKey, &sshKey); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SSH public key %s not found for user %s", sshPublicKeyId, userName)), nil
	}

	sshKey.Status = status

	if err := s.state.Set(stateKey, sshKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to update SSH public key"), nil
	}

	return s.successResponse("UpdateSSHPublicKey", EmptyResult{})
}

// deleteSSHPublicKey deletes an SSH public key
func (s *IAMService) deleteSSHPublicKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	sshPublicKeyId := emulator.GetStringParam(params, "SSHPublicKeyId", "")
	if sshPublicKeyId == "" {
		return s.errorResponse(400, "ValidationError", "SSHPublicKeyId is required"), nil
	}

	stateKey := fmt.Sprintf("iam:ssh-public-key:%s:%s", userName, sshPublicKeyId)
	var sshKey SSHPublicKeyData
	if err := s.state.Get(stateKey, &sshKey); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("SSH public key %s not found for user %s", sshPublicKeyId, userName)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete SSH public key"), nil
	}

	return s.successResponse("DeleteSSHPublicKey", EmptyResult{})
}

// listSSHPublicKeys lists SSH public keys for a user
func (s *IAMService) listSSHPublicKeys(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	prefix := fmt.Sprintf("iam:ssh-public-key:%s:", userName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list SSH public keys"), nil
	}

	var sshKeys []SSHPublicKeyMetadataListItem
	for _, key := range keys {
		var sshKey SSHPublicKeyData
		if err := s.state.Get(key, &sshKey); err != nil {
			continue
		}

		sshKeys = append(sshKeys, SSHPublicKeyMetadataListItem{
			UserName:       sshKey.UserName,
			SSHPublicKeyId: sshKey.SSHPublicKeyId,
			Status:         sshKey.Status,
			UploadDate:     sshKey.UploadDate,
		})
	}

	result := ListSSHPublicKeysResult{
		SSHPublicKeys: sshKeys,
		IsTruncated:   false,
	}

	return s.successResponse("ListSSHPublicKeys", result)
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidSSHPublicKeyId validates an SSH public key ID format
func isValidSSHPublicKeyId(id string) bool {
	if len(id) < 1 || len(id) > 128 {
		return false
	}
	matched, _ := regexp.MatchString(`^APKA[A-Z0-9]+$`, id)
	return matched
}

// generateSSHPublicKeyId generates a unique SSH public key ID
func generateSSHPublicKeyId() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "APKA" + strings.ToUpper(hex.EncodeToString(b))[:16]
}

// generateSSHKeyFingerprint generates a fingerprint from the SSH public key body
func generateSSHKeyFingerprint(keyBody string) string {
	// Simplified fingerprint - real AWS uses MD5 of the decoded key
	hash := sha256.Sum256([]byte(keyBody))
	hexStr := hex.EncodeToString(hash[:])
	// Format as colon-separated pairs (first 32 chars / 16 bytes)
	var parts []string
	for i := 0; i < 32; i += 2 {
		parts = append(parts, hexStr[i:i+2])
	}
	return strings.Join(parts, ":")
}
