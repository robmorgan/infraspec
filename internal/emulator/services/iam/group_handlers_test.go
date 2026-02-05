package iam

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

// ============================================================================
// Group CRUD Tests
// ============================================================================

func TestCreateGroup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=Developers&Path=/engineering/"),
		Action:  "CreateGroup",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	require.Contains(t, string(resp.Body), "<GroupName>Developers</GroupName>")
	require.Contains(t, string(resp.Body), "<Path>/engineering/</Path>")
	require.Contains(t, string(resp.Body), "AGPA")
}

func TestCreateGroup_Duplicate(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create first group
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=Admins"),
		Action:  "CreateGroup",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Try to create duplicate
	resp, err = service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestCreateGroup_MissingGroupName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup"),
		Action:  "CreateGroup",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "GroupName is required")
}

func TestGetGroup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a group first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=TestGroup"),
		Action:  "CreateGroup",
	}
	_, err := service.HandleRequest(context.Background(), createReq)
	require.NoError(t, err)

	// Get the group
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=TestGroup"),
		Action:  "GetGroup",
	}
	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<GroupName>TestGroup</GroupName>")
}

func TestGetGroup_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=NonexistentGroup"),
		Action:  "GetGroup",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestDeleteGroup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a group
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=ToDelete"),
		Action:  "CreateGroup",
	}
	_, err := service.HandleRequest(context.Background(), createReq)
	require.NoError(t, err)

	// Delete the group
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteGroup&GroupName=ToDelete"),
		Action:  "DeleteGroup",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify it's deleted
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=ToDelete"),
		Action:  "GetGroup",
	}
	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

func TestDeleteGroup_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteGroup&GroupName=NonexistentGroup"),
		Action:  "DeleteGroup",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListGroups_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create multiple groups
	for _, name := range []string{"Group1", "Group2", "Group3"} {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateGroup&GroupName=" + name),
			Action:  "CreateGroup",
		}
		_, err := service.HandleRequest(context.Background(), req)
		require.NoError(t, err)
	}

	// List groups
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListGroups"),
		Action:  "ListGroups",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "Group1")
	require.Contains(t, body, "Group2")
	require.Contains(t, body, "Group3")
}

func TestUpdateGroup_RenameSuccess(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a group
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=OldName"),
		Action:  "CreateGroup",
	}
	_, err := service.HandleRequest(context.Background(), createReq)
	require.NoError(t, err)

	// Rename the group
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateGroup&GroupName=OldName&NewGroupName=NewName"),
		Action:  "UpdateGroup",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify old name doesn't exist
	getOldReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=OldName"),
		Action:  "GetGroup",
	}
	resp, err = service.HandleRequest(context.Background(), getOldReq)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	// Verify new name exists
	getNewReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=NewName"),
		Action:  "GetGroup",
	}
	resp, err = service.HandleRequest(context.Background(), getNewReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<GroupName>NewName</GroupName>")
}

// ============================================================================
// Group Membership Tests
// ============================================================================

func TestAddUserToGroup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a user and a group
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=TestUser"),
		Action:  "CreateUser",
	}
	_, err := service.HandleRequest(context.Background(), createUserReq)
	require.NoError(t, err)

	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=TestGroup"),
		Action:  "CreateGroup",
	}
	_, err = service.HandleRequest(context.Background(), createGroupReq)
	require.NoError(t, err)

	// Add user to group
	addReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AddUserToGroup&GroupName=TestGroup&UserName=TestUser"),
		Action:  "AddUserToGroup",
	}
	resp, err := service.HandleRequest(context.Background(), addReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify by getting group
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=TestGroup"),
		Action:  "GetGroup",
	}
	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	require.Contains(t, string(resp.Body), "TestUser")
}

func TestRemoveUserFromGroup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and group
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=TestUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=TestGroup"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	// Add user to group
	addReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AddUserToGroup&GroupName=TestGroup&UserName=TestUser"),
		Action:  "AddUserToGroup",
	}
	_, _ = service.HandleRequest(context.Background(), addReq)

	// Remove user from group
	removeReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=RemoveUserFromGroup&GroupName=TestGroup&UserName=TestUser"),
		Action:  "RemoveUserFromGroup",
	}
	resp, err := service.HandleRequest(context.Background(), removeReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify by getting group - user should not be in the list
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroup&GroupName=TestGroup"),
		Action:  "GetGroup",
	}
	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	require.NotContains(t, string(resp.Body), "<UserName>TestUser</UserName>")
}

func TestListGroupsForUser_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and groups
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=MultiGroupUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	for _, groupName := range []string{"Engineering", "QA"} {
		createGroupReq := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateGroup&GroupName=" + groupName),
			Action:  "CreateGroup",
		}
		_, _ = service.HandleRequest(context.Background(), createGroupReq)

		addReq := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=AddUserToGroup&GroupName=" + groupName + "&UserName=MultiGroupUser"),
			Action:  "AddUserToGroup",
		}
		_, _ = service.HandleRequest(context.Background(), addReq)
	}

	// List groups for user
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListGroupsForUser&UserName=MultiGroupUser"),
		Action:  "ListGroupsForUser",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "Engineering")
	require.Contains(t, body, "QA")
}

func TestDeleteGroup_WithMembers_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user and group
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=Member"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=GroupWithMembers"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	// Add user to group
	addReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AddUserToGroup&GroupName=GroupWithMembers&UserName=Member"),
		Action:  "AddUserToGroup",
	}
	_, _ = service.HandleRequest(context.Background(), addReq)

	// Try to delete group - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteGroup&GroupName=GroupWithMembers"),
		Action:  "DeleteGroup",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "DeleteConflict")
}

// ============================================================================
// Group Inline Policy Tests
// ============================================================================

func TestGroupInlinePolicy_CRUD(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a group
	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=PolicyTestGroup"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`

	// Put inline policy
	putReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=PutGroupPolicy&GroupName=PolicyTestGroup&PolicyName=S3ReadPolicy&PolicyDocument=" + policyDoc),
		Action:  "PutGroupPolicy",
	}
	resp, err := service.HandleRequest(context.Background(), putReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Get inline policy
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetGroupPolicy&GroupName=PolicyTestGroup&PolicyName=S3ReadPolicy"),
		Action:  "GetGroupPolicy",
	}
	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "S3ReadPolicy")

	// List inline policies
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListGroupPolicies&GroupName=PolicyTestGroup"),
		Action:  "ListGroupPolicies",
	}
	resp, err = service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "S3ReadPolicy")

	// Delete inline policy
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteGroupPolicy&GroupName=PolicyTestGroup&PolicyName=S3ReadPolicy"),
		Action:  "DeleteGroupPolicy",
	}
	resp, err = service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deleted
	resp, err = service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

// ============================================================================
// Group Policy Attachment Tests
// ============================================================================

func TestAttachGroupPolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a group
	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=AttachTestGroup"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	// Create a policy
	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`
	createPolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreatePolicy&PolicyName=TestManagedPolicy&PolicyDocument=" + policyDoc),
		Action:  "CreatePolicy",
	}
	resp, err := service.HandleRequest(context.Background(), createPolicyReq)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Extract the policy ARN
	body := string(resp.Body)
	arnStart := strings.Index(body, "<Arn>") + 5
	arnEnd := strings.Index(body, "</Arn>")
	policyArn := body[arnStart:arnEnd]

	// Attach policy to group
	attachReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AttachGroupPolicy&GroupName=AttachTestGroup&PolicyArn=" + policyArn),
		Action:  "AttachGroupPolicy",
	}
	resp, err = service.HandleRequest(context.Background(), attachReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// List attached policies
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAttachedGroupPolicies&GroupName=AttachTestGroup"),
		Action:  "ListAttachedGroupPolicies",
	}
	resp, err = service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "TestManagedPolicy")

	// Detach policy
	detachReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DetachGroupPolicy&GroupName=AttachTestGroup&PolicyArn=" + policyArn),
		Action:  "DetachGroupPolicy",
	}
	resp, err = service.HandleRequest(context.Background(), detachReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify detached
	resp, err = service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.NotContains(t, string(resp.Body), "TestManagedPolicy")
}

func TestDeleteGroup_WithAttachedPolicy_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create group
	createGroupReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateGroup&GroupName=GroupWithPolicy"),
		Action:  "CreateGroup",
	}
	_, _ = service.HandleRequest(context.Background(), createGroupReq)

	// Create and attach a policy
	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`
	createPolicyReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreatePolicy&PolicyName=BlockingPolicy&PolicyDocument=" + policyDoc),
		Action:  "CreatePolicy",
	}
	resp, _ := service.HandleRequest(context.Background(), createPolicyReq)
	body := string(resp.Body)
	arnStart := strings.Index(body, "<Arn>") + 5
	arnEnd := strings.Index(body, "</Arn>")
	policyArn := body[arnStart:arnEnd]

	attachReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=AttachGroupPolicy&GroupName=GroupWithPolicy&PolicyArn=" + policyArn),
		Action:  "AttachGroupPolicy",
	}
	_, _ = service.HandleRequest(context.Background(), attachReq)

	// Try to delete group - should fail
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteGroup&GroupName=GroupWithPolicy"),
		Action:  "DeleteGroup",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "DeleteConflict")
}
