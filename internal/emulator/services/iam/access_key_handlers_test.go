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
// Access Key Tests
// ============================================================================

func TestCreateAccessKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a user first
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=TestUser"),
		Action:  "CreateUser",
	}
	_, err := service.HandleRequest(context.Background(), createUserReq)
	require.NoError(t, err)

	// Create access key
	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=TestUser"),
		Action:  "CreateAccessKey",
	}
	resp, err := service.HandleRequest(context.Background(), createKeyReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<UserName>TestUser</UserName>")
	require.Contains(t, body, "<AccessKeyId>AKIA")
	require.Contains(t, body, "<SecretAccessKey>")
	require.Contains(t, body, "<Status>Active</Status>")
}

func TestCreateAccessKey_UserNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=NonexistentUser"),
		Action:  "CreateAccessKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestCreateAccessKey_LimitExceeded(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=LimitTestUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create two access keys (the maximum)
	for i := 0; i < 2; i++ {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateAccessKey&UserName=LimitTestUser"),
			Action:  "CreateAccessKey",
		}
		resp, _ := service.HandleRequest(context.Background(), req)
		require.Equal(t, 200, resp.StatusCode)
	}

	// Try to create a third - should fail
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=LimitTestUser"),
		Action:  "CreateAccessKey",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "LimitExceeded")
}

func TestDeleteAccessKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and access key
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=DeleteKeyUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=DeleteKeyUser"),
		Action:  "CreateAccessKey",
	}
	resp, _ := service.HandleRequest(context.Background(), createKeyReq)

	// Extract access key ID
	body := string(resp.Body)
	start := strings.Index(body, "<AccessKeyId>") + 13
	end := strings.Index(body, "</AccessKeyId>")
	accessKeyId := body[start:end]

	// Delete the access key
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccessKey&UserName=DeleteKeyUser&AccessKeyId=" + accessKeyId),
		Action:  "DeleteAccessKey",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify it's deleted by listing
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccessKeys&UserName=DeleteKeyUser"),
		Action:  "ListAccessKeys",
	}
	resp, _ = service.HandleRequest(context.Background(), listReq)
	require.NotContains(t, string(resp.Body), accessKeyId)
}

func TestDeleteAccessKey_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=TestUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccessKey&UserName=TestUser&AccessKeyId=AKIANONEXISTENT"),
		Action:  "DeleteAccessKey",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestUpdateAccessKey_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and access key
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=UpdateKeyUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=UpdateKeyUser"),
		Action:  "CreateAccessKey",
	}
	resp, _ := service.HandleRequest(context.Background(), createKeyReq)

	// Extract access key ID
	body := string(resp.Body)
	start := strings.Index(body, "<AccessKeyId>") + 13
	end := strings.Index(body, "</AccessKeyId>")
	accessKeyId := body[start:end]

	// Update to Inactive
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccessKey&UserName=UpdateKeyUser&AccessKeyId=" + accessKeyId + "&Status=Inactive"),
		Action:  "UpdateAccessKey",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify status changed by listing
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccessKeys&UserName=UpdateKeyUser"),
		Action:  "ListAccessKeys",
	}
	resp, _ = service.HandleRequest(context.Background(), listReq)
	require.Contains(t, string(resp.Body), "<Status>Inactive</Status>")
}

func TestUpdateAccessKey_InvalidStatus(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and access key
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=InvalidStatusUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=InvalidStatusUser"),
		Action:  "CreateAccessKey",
	}
	resp, _ := service.HandleRequest(context.Background(), createKeyReq)

	// Extract access key ID
	body := string(resp.Body)
	start := strings.Index(body, "<AccessKeyId>") + 13
	end := strings.Index(body, "</AccessKeyId>")
	accessKeyId := body[start:end]

	// Try to update with invalid status
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccessKey&UserName=InvalidStatusUser&AccessKeyId=" + accessKeyId + "&Status=InvalidStatus"),
		Action:  "UpdateAccessKey",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Status must be Active or Inactive")
}

func TestListAccessKeys_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and two access keys
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=ListKeysUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	for i := 0; i < 2; i++ {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateAccessKey&UserName=ListKeysUser"),
			Action:  "CreateAccessKey",
		}
		_, _ = service.HandleRequest(context.Background(), req)
	}

	// List access keys
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccessKeys&UserName=ListKeysUser"),
		Action:  "ListAccessKeys",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	body := string(resp.Body)
	// Count occurrences of AccessKeyId
	count := strings.Count(body, "<AccessKeyId>AKIA")
	require.Equal(t, 2, count)
}

func TestListAccessKeys_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user without access keys
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=NoKeysUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// List access keys
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccessKeys&UserName=NoKeysUser"),
		Action:  "ListAccessKeys",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.NotContains(t, string(resp.Body), "<AccessKeyId>")
}

func TestGetAccessKeyLastUsed_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and access key
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=LastUsedUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=LastUsedUser"),
		Action:  "CreateAccessKey",
	}
	resp, _ := service.HandleRequest(context.Background(), createKeyReq)

	// Extract access key ID
	body := string(resp.Body)
	start := strings.Index(body, "<AccessKeyId>") + 13
	end := strings.Index(body, "</AccessKeyId>")
	accessKeyId := body[start:end]

	// Get last used info
	lastUsedReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetAccessKeyLastUsed&AccessKeyId=" + accessKeyId),
		Action:  "GetAccessKeyLastUsed",
	}
	resp, err := service.HandleRequest(context.Background(), lastUsedReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<UserName>LastUsedUser</UserName>")
}

func TestGetAccessKeyLastUsed_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetAccessKeyLastUsed&AccessKeyId=AKIANONEXISTENT"),
		Action:  "GetAccessKeyLastUsed",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestAccessKeyIdFormat(t *testing.T) {
	// Test that generated access key IDs have the correct format
	keyId := generateAccessKeyId()
	require.True(t, strings.HasPrefix(keyId, "AKIA"))
	require.Equal(t, 20, len(keyId))
}

func TestSecretAccessKeyFormat(t *testing.T) {
	// Test that generated secret access keys have the correct length
	secret := generateSecretAccessKey()
	require.Equal(t, 40, len(secret))
}
