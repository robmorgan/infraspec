package iam

import (
	"context"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser&Path=/test/"),
		Action:  "CreateUser",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<UserName>testuser</UserName>")
	assert.Contains(t, string(resp.Body), "<Path>/test/</Path>")
	assert.Contains(t, string(resp.Body), "arn:aws:iam::123456789012:user/test/testuser")
}

func TestCreateUser_Duplicate(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create first user
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create duplicate
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestCreateUser_MissingUserName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser"),
		Action:  "CreateUser",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "ValidationError")
}

func TestGetUser_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Get user
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUser&UserName=testuser"),
		Action:  "GetUser",
	}

	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<UserName>testuser</UserName>")
}

func TestGetUser_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUser&UserName=nonexistent"),
		Action:  "GetUser",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestDeleteUser_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete user
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify user is gone
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUser&UserName=testuser"),
		Action:  "GetUser",
	}

	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestDeleteUser_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=nonexistent"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestListUsers_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create multiple users
	users := []string{"user1", "user2", "user3"}
	for _, u := range users {
		createReq := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateUser&UserName=" + u),
			Action:  "CreateUser",
		}
		_, _ = service.HandleRequest(context.Background(), createReq)
	}

	// List users
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListUsers"),
		Action:  "ListUsers",
	}

	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body := string(resp.Body)
	for _, u := range users {
		assert.Contains(t, body, "<UserName>"+u+"</UserName>")
	}
}

func TestUpdateUser_RenameSuccess(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=oldname"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Update (rename) user
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateUser&UserName=oldname&NewUserName=newname"),
		Action:  "UpdateUser",
	}

	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify old name doesn't exist
	getOldReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUser&UserName=oldname"),
		Action:  "GetUser",
	}
	resp, _ = service.HandleRequest(context.Background(), getOldReq)
	assert.Equal(t, 404, resp.StatusCode)

	// Verify new name exists
	getNewReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUser&UserName=newname"),
		Action:  "GetUser",
	}
	resp, err = service.HandleRequest(context.Background(), getNewReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<UserName>newname</UserName>")
}

func TestTagUser_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Tag user
	tagReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=TagUser&UserName=testuser&Tags.member.1.Key=Environment&Tags.member.1.Value=Production"),
		Action:  "TagUser",
	}

	resp, err := service.HandleRequest(context.Background(), tagReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// List tags to verify
	listTagsReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListUserTags&UserName=testuser"),
		Action:  "ListUserTags",
	}

	resp, err = service.HandleRequest(context.Background(), listTagsReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<Key>Environment</Key>")
	assert.Contains(t, string(resp.Body), "<Value>Production</Value>")
}

func TestLoginProfile_CRUD(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create login profile
	createProfileReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateLoginProfile&UserName=testuser&Password=MyPassword123!"),
		Action:  "CreateLoginProfile",
	}

	resp, err := service.HandleRequest(context.Background(), createProfileReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<UserName>testuser</UserName>")

	// Get login profile
	getProfileReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetLoginProfile&UserName=testuser"),
		Action:  "GetLoginProfile",
	}

	resp, err = service.HandleRequest(context.Background(), getProfileReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<UserName>testuser</UserName>")

	// Delete login profile
	deleteProfileReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteLoginProfile&UserName=testuser"),
		Action:  "DeleteLoginProfile",
	}

	resp, err = service.HandleRequest(context.Background(), deleteProfileReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify login profile is gone
	resp, _ = service.HandleRequest(context.Background(), getProfileReq)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestUserInlinePolicy_CRUD(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`

	// Put inline policy
	putPolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=PutUserPolicy&UserName=testuser&PolicyName=TestPolicy&PolicyDocument=" + policyDoc),
		Action:  "PutUserPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), putPolicyReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Get inline policy
	getPolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetUserPolicy&UserName=testuser&PolicyName=TestPolicy"),
		Action:  "GetUserPolicy",
	}

	resp, err = service.HandleRequest(context.Background(), getPolicyReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "<PolicyName>TestPolicy</PolicyName>")

	// List inline policies
	listPoliciesReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListUserPolicies&UserName=testuser"),
		Action:  "ListUserPolicies",
	}

	resp, err = service.HandleRequest(context.Background(), listPoliciesReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "TestPolicy")

	// Delete inline policy
	deletePolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUserPolicy&UserName=testuser&PolicyName=TestPolicy"),
		Action:  "DeleteUserPolicy",
	}

	resp, err = service.HandleRequest(context.Background(), deletePolicyReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAttachUserPolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create policy
	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`
	createPolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreatePolicy&PolicyName=TestPolicy&PolicyDocument=" + policyDoc),
		Action:  "CreatePolicy",
	}
	_, _ = service.HandleRequest(context.Background(), createPolicyReq)

	// Attach policy to user
	attachReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AttachUserPolicy&UserName=testuser&PolicyArn=arn:aws:iam::123456789012:policy/TestPolicy"),
		Action:  "AttachUserPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), attachReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// List attached policies
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAttachedUserPolicies&UserName=testuser"),
		Action:  "ListAttachedUserPolicies",
	}

	resp, err = service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "TestPolicy")
}

func TestDeleteUser_WithLoginProfile_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create login profile
	createProfileReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateLoginProfile&UserName=testuser&Password=MyPassword123!"),
		Action:  "CreateLoginProfile",
	}
	_, _ = service.HandleRequest(context.Background(), createProfileReq)

	// Try to delete user - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.True(t, strings.Contains(string(resp.Body), "DeleteConflict") || strings.Contains(string(resp.Body), "login profile"))
}

func TestDeleteUser_WithAccessKey_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create access key
	createKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccessKey&UserName=testuser"),
		Action:  "CreateAccessKey",
	}
	_, _ = service.HandleRequest(context.Background(), createKeyReq)

	// Try to delete user - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "DeleteConflict")
	assert.Contains(t, string(resp.Body), "access keys")
}

func TestDeleteUser_WithSSHPublicKey_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Upload SSH public key
	sshKeyBody := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... test@example.com"
	uploadKeyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UploadSSHPublicKey&UserName=testuser&SSHPublicKeyBody=" + sshKeyBody),
		Action:  "UploadSSHPublicKey",
	}
	_, _ = service.HandleRequest(context.Background(), uploadKeyReq)

	// Try to delete user - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "DeleteConflict")
	assert.Contains(t, string(resp.Body), "SSH public keys")
}

func TestDeleteUser_WithMFADevice_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create virtual MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=testmfa"),
		Action:  "CreateVirtualMFADevice",
	}
	resp, _ := service.HandleRequest(context.Background(), createMFAReq)
	// Extract serial number from response
	body := string(resp.Body)
	serialStart := strings.Index(body, "<SerialNumber>") + len("<SerialNumber>")
	serialEnd := strings.Index(body, "</SerialNumber>")
	serialNumber := body[serialStart:serialEnd]

	// Enable MFA device for user
	enableMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=testuser&SerialNumber=" + serialNumber + "&AuthenticationCode1=123456&AuthenticationCode2=654321"),
		Action:  "EnableMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), enableMFAReq)

	// Try to delete user - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "DeleteConflict")
	assert.Contains(t, string(resp.Body), "MFA devices")
}

func TestDeleteUser_WithGroupMembership_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=testuser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create group
	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=testgroup"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	// Add user to group
	addToGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AddUserToGroup&UserName=testuser&GroupName=testgroup"),
		Action:  "AddUserToGroup",
	}
	_, _ = service.HandleRequest(context.Background(), addToGroupReq)

	// Try to delete user - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteUser&UserName=testuser"),
		Action:  "DeleteUser",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "DeleteConflict")
	assert.Contains(t, string(resp.Body), "groups")
}
