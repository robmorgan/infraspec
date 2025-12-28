package iam

import (
	"context"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SSH Public Key Tests
// ============================================================================

func TestUploadSSHPublicKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user first
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ+test@example.com"),
		Action:  "UploadSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<SSHPublicKeyId>APKA")
	require.Contains(t, body, "<UserName>sshuser</UserName>")
	require.Contains(t, body, "<Status>Active</Status>")
	require.Contains(t, body, "<Fingerprint>")
}

func TestUploadSSHPublicKey_MissingUserName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ"),
		Action:  "UploadSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "UserName is required")
}

func TestUploadSSHPublicKey_MissingKeyBody(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser"),
		Action:  "UploadSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "SSHPublicKeyBody is required")
}

func TestUploadSSHPublicKey_UserNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=nonexistent&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ"),
		Action:  "UploadSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestGetSSHPublicKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser2"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Upload SSH key
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser2&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ"),
		Action:  "UploadSSHPublicKey",
	}
	uploadResp, _ := service.HandleRequest(context.Background(), uploadReq)
	body := string(uploadResp.Body)

	// Extract SSH key ID from response
	start := strings.Index(body, "<SSHPublicKeyId>") + len("<SSHPublicKeyId>")
	end := strings.Index(body[start:], "</SSHPublicKeyId>") + start
	keyId := body[start:end]

	// Get the key
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSSHPublicKey&UserName=sshuser2&SSHPublicKeyId=" + keyId),
		Action:  "GetSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	respBody := string(resp.Body)
	require.Contains(t, respBody, "<SSHPublicKeyId>"+keyId+"</SSHPublicKeyId>")
	require.Contains(t, respBody, "<UserName>sshuser2</UserName>")
}

func TestGetSSHPublicKey_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSSHPublicKey&UserName=someuser&SSHPublicKeyId=APKAFAKEKEY12345678"),
		Action:  "GetSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestUpdateSSHPublicKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser3"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Upload SSH key
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser3&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ"),
		Action:  "UploadSSHPublicKey",
	}
	uploadResp, _ := service.HandleRequest(context.Background(), uploadReq)
	body := string(uploadResp.Body)

	// Extract SSH key ID
	start := strings.Index(body, "<SSHPublicKeyId>") + len("<SSHPublicKeyId>")
	end := strings.Index(body[start:], "</SSHPublicKeyId>") + start
	keyId := body[start:end]

	// Update the key status
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateSSHPublicKey&UserName=sshuser3&SSHPublicKeyId=" + keyId + "&Status=Inactive"),
		Action:  "UpdateSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify status changed
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSSHPublicKey&UserName=sshuser3&SSHPublicKeyId=" + keyId),
		Action:  "GetSSHPublicKey",
	}
	getResp, _ := service.HandleRequest(context.Background(), getReq)
	require.Contains(t, string(getResp.Body), "<Status>Inactive</Status>")
}

func TestUpdateSSHPublicKey_InvalidStatus(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateSSHPublicKey&UserName=someuser&SSHPublicKeyId=APKAFAKEKEY&Status=Invalid"),
		Action:  "UpdateSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Status must be Active or Inactive")
}

func TestDeleteSSHPublicKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser4"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Upload SSH key
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser4&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQ"),
		Action:  "UploadSSHPublicKey",
	}
	uploadResp, _ := service.HandleRequest(context.Background(), uploadReq)
	body := string(uploadResp.Body)

	// Extract SSH key ID
	start := strings.Index(body, "<SSHPublicKeyId>") + len("<SSHPublicKeyId>")
	end := strings.Index(body[start:], "</SSHPublicKeyId>") + start
	keyId := body[start:end]

	// Delete the key
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteSSHPublicKey&UserName=sshuser4&SSHPublicKeyId=" + keyId),
		Action:  "DeleteSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSSHPublicKey&UserName=sshuser4&SSHPublicKeyId=" + keyId),
		Action:  "GetSSHPublicKey",
	}
	getResp, _ := service.HandleRequest(context.Background(), getReq)
	require.Equal(t, 404, getResp.StatusCode)
}

func TestDeleteSSHPublicKey_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteSSHPublicKey&UserName=someuser&SSHPublicKeyId=APKAFAKEKEY12345678"),
		Action:  "DeleteSSHPublicKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListSSHPublicKeys_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser5"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Upload two SSH keys
	for i := 0; i < 2; i++ {
		uploadReq := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=UploadSSHPublicKey&UserName=sshuser5&SSHPublicKeyBody=ssh-rsa+AAAAB3NzaC1yc2EAAAADAQABAAABAQtest" + string(rune('0'+i))),
			Action:  "UploadSSHPublicKey",
		}
		_, _ = service.HandleRequest(context.Background(), uploadReq)
	}

	// List keys
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListSSHPublicKeys&UserName=sshuser5"),
		Action:  "ListSSHPublicKeys",
	}

	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "<SSHPublicKeys>")
	// Should have 2 keys
	require.Equal(t, 2, strings.Count(body, "<SSHPublicKeyId>"))
}

func TestListSSHPublicKeys_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user with no SSH keys
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=sshuser6"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// List keys
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListSSHPublicKeys&UserName=sshuser6"),
		Action:  "ListSSHPublicKeys",
	}

	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<SSHPublicKeys")
}

func TestListSSHPublicKeys_UserNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListSSHPublicKeys&UserName=nonexistent"),
		Action:  "ListSSHPublicKeys",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestGenerateSSHPublicKeyId(t *testing.T) {
	id := generateSSHPublicKeyId()
	require.True(t, strings.HasPrefix(id, "APKA"))
	require.Equal(t, 20, len(id))
}

func TestGenerateSSHKeyFingerprint(t *testing.T) {
	fingerprint := generateSSHKeyFingerprint("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test@example.com")
	require.NotEmpty(t, fingerprint)
	// Check fingerprint format (colon-separated hex pairs)
	require.Contains(t, fingerprint, ":")
	parts := strings.Split(fingerprint, ":")
	require.Equal(t, 16, len(parts))
}
