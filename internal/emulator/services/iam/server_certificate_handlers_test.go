package iam

import (
	"context"
	"strings"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Server Certificate Tests
// ============================================================================

func TestUploadServerCertificate_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=TestCert&CertificateBody=-----BEGIN CERTIFICATE-----&PrivateKey=-----BEGIN PRIVATE KEY-----"),
		Action:  "UploadServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<ServerCertificateName>TestCert</ServerCertificateName>")
	require.Contains(t, body, "<ServerCertificateId>ASCA")
	require.Contains(t, body, "arn:aws:iam::123456789012:server-certificate/TestCert")
}

func TestUploadServerCertificate_MissingName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&CertificateBody=cert&PrivateKey=key"),
		Action:  "UploadServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "ServerCertificateName is required")
}

func TestUploadServerCertificate_MissingCertificateBody(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=TestCert&PrivateKey=key"),
		Action:  "UploadServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "CertificateBody is required")
}

func TestUploadServerCertificate_MissingPrivateKey(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=TestCert&CertificateBody=cert"),
		Action:  "UploadServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "PrivateKey is required")
}

func TestUploadServerCertificate_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=DuplicateCert&CertificateBody=cert&PrivateKey=key"),
		Action:  "UploadServerCertificate",
	}

	// Create first
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create again
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestGetServerCertificate_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Upload certificate
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=GetTestCert&CertificateBody=-----BEGIN CERTIFICATE-----&PrivateKey=-----BEGIN PRIVATE KEY-----&CertificateChain=-----BEGIN CERTIFICATE CHAIN-----"),
		Action:  "UploadServerCertificate",
	}
	_, _ = service.HandleRequest(context.Background(), uploadReq)

	// Get certificate
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServerCertificate&ServerCertificateName=GetTestCert"),
		Action:  "GetServerCertificate",
	}
	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "<ServerCertificateName>GetTestCert</ServerCertificateName>")
	require.Contains(t, body, "<CertificateBody>-----BEGIN CERTIFICATE-----</CertificateBody>")
	require.Contains(t, body, "<CertificateChain>-----BEGIN CERTIFICATE CHAIN-----</CertificateChain>")
}

func TestGetServerCertificate_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServerCertificate&ServerCertificateName=NonexistentCert"),
		Action:  "GetServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestDeleteServerCertificate_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Upload certificate
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=DeleteTestCert&CertificateBody=cert&PrivateKey=key"),
		Action:  "UploadServerCertificate",
	}
	_, _ = service.HandleRequest(context.Background(), uploadReq)

	// Delete certificate
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteServerCertificate&ServerCertificateName=DeleteTestCert"),
		Action:  "DeleteServerCertificate",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServerCertificate&ServerCertificateName=DeleteTestCert"),
		Action:  "GetServerCertificate",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	require.Equal(t, 404, resp.StatusCode)
}

func TestDeleteServerCertificate_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteServerCertificate&ServerCertificateName=NonexistentCert"),
		Action:  "DeleteServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListServerCertificates_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Upload two certificates
	for _, name := range []string{"Cert1", "Cert2"} {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=" + name + "&CertificateBody=cert&PrivateKey=key"),
			Action:  "UploadServerCertificate",
		}
		_, _ = service.HandleRequest(context.Background(), req)
	}

	// List certificates
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListServerCertificates"),
		Action:  "ListServerCertificates",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "Cert1")
	require.Contains(t, body, "Cert2")
}

func TestListServerCertificates_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListServerCertificates"),
		Action:  "ListServerCertificates",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<ServerCertificateMetadataList")
}

func TestUpdateServerCertificate_Rename(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Upload certificate
	uploadReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadServerCertificate&ServerCertificateName=OldName&CertificateBody=cert&PrivateKey=key"),
		Action:  "UploadServerCertificate",
	}
	_, _ = service.HandleRequest(context.Background(), uploadReq)

	// Rename certificate
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateServerCertificate&ServerCertificateName=OldName&NewServerCertificateName=NewName"),
		Action:  "UpdateServerCertificate",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify old name is gone
	getOldReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServerCertificate&ServerCertificateName=OldName"),
		Action:  "GetServerCertificate",
	}
	resp, _ = service.HandleRequest(context.Background(), getOldReq)
	require.Equal(t, 404, resp.StatusCode)

	// Verify new name exists
	getNewReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServerCertificate&ServerCertificateName=NewName"),
		Action:  "GetServerCertificate",
	}
	resp, _ = service.HandleRequest(context.Background(), getNewReq)
	require.Equal(t, 200, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NewName")
}

func TestUpdateServerCertificate_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateServerCertificate&ServerCertificateName=NonexistentCert&NewServerCertificateName=NewName"),
		Action:  "UpdateServerCertificate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestIsValidServerCertificateName(t *testing.T) {
	// Valid names
	require.True(t, isValidServerCertificateName("MyCert"))
	require.True(t, isValidServerCertificateName("My_Cert"))
	require.True(t, isValidServerCertificateName("My-Cert"))
	require.True(t, isValidServerCertificateName("cert@domain.com"))
	require.True(t, isValidServerCertificateName("a"))

	// Invalid names
	require.False(t, isValidServerCertificateName(""))
	require.False(t, isValidServerCertificateName(strings.Repeat("a", 129)))
}

func TestGenerateServerCertificateId(t *testing.T) {
	id := generateServerCertificateId()
	require.True(t, strings.HasPrefix(id, "ASCA"))
	require.Equal(t, 20, len(id))
}
