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
// Role Update Tests
// ============================================================================

func TestUpdateRole_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=TestRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}`),
		Action:  "CreateRole",
	}
	_, err := service.HandleRequest(context.Background(), createReq)
	require.NoError(t, err)

	// Update the role
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateRole&RoleName=TestRole&Description=Updated+description&MaxSessionDuration=7200"),
		Action:  "UpdateRole",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify changes by getting the role
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetRole&RoleName=TestRole"),
		Action:  "GetRole",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	body := string(resp.Body)
	require.Contains(t, body, "<Description>Updated description</Description>")
	require.Contains(t, body, "<MaxSessionDuration>7200</MaxSessionDuration>")
}

func TestUpdateRole_InvalidMaxSessionDuration(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=TestRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}`),
		Action:  "CreateRole",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Try to update with invalid max session duration
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateRole&RoleName=TestRole&MaxSessionDuration=100"),
		Action:  "UpdateRole",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "MaxSessionDuration must be between")
}

func TestUpdateRole_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateRole&RoleName=NonexistentRole&Description=test"),
		Action:  "UpdateRole",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestUpdateRoleDescription_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=DescRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}`),
		Action:  "CreateRole",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Update description
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateRoleDescription&RoleName=DescRole&Description=New+description"),
		Action:  "UpdateRoleDescription",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<Description>New description</Description>")
}

// ============================================================================
// Role Tag Tests
// ============================================================================

func TestTagRole_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=TagRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}`),
		Action:  "CreateRole",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Tag the role
	tagReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=TagRole&RoleName=TagRole&Tags.member.1.Key=Environment&Tags.member.1.Value=Production"),
		Action:  "TagRole",
	}
	resp, err := service.HandleRequest(context.Background(), tagReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify tags
	listTagsReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListRoleTags&RoleName=TagRole"),
		Action:  "ListRoleTags",
	}
	resp, _ = service.HandleRequest(context.Background(), listTagsReq)
	body := string(resp.Body)
	require.Contains(t, body, "<Key>Environment</Key>")
	require.Contains(t, body, "<Value>Production</Value>")
}

func TestUntagRole_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role with tags
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=UntagRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}&Tags.member.1.Key=ToRemove&Tags.member.1.Value=Value1&Tags.member.2.Key=ToKeep&Tags.member.2.Value=Value2`),
		Action:  "CreateRole",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Untag the role
	untagReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UntagRole&RoleName=UntagRole&TagKeys.member.1=ToRemove"),
		Action:  "UntagRole",
	}
	resp, err := service.HandleRequest(context.Background(), untagReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify the tag was removed
	listTagsReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListRoleTags&RoleName=UntagRole"),
		Action:  "ListRoleTags",
	}
	resp, _ = service.HandleRequest(context.Background(), listTagsReq)
	body := string(resp.Body)
	require.NotContains(t, body, "<Key>ToRemove</Key>")
	require.Contains(t, body, "<Key>ToKeep</Key>")
}

func TestListRoleTags_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role with multiple tags
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte(`Action=CreateRole&RoleName=MultiTagRole&AssumeRolePolicyDocument={"Version":"2012-10-17","Statement":[]}&Tags.member.1.Key=Env&Tags.member.1.Value=Dev&Tags.member.2.Key=Team&Tags.member.2.Value=Engineering`),
		Action:  "CreateRole",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// List tags
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListRoleTags&RoleName=MultiTagRole"),
		Action:  "ListRoleTags",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "ListRoleTagsResult")
	require.Contains(t, body, "<IsTruncated>false</IsTruncated>")
}

func TestListRoleTags_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListRoleTags&RoleName=NonexistentRole"),
		Action:  "ListRoleTags",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

// ============================================================================
// Service-Linked Role Tests
// ============================================================================

func TestCreateServiceLinkedRole_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateServiceLinkedRole&AWSServiceName=elasticmapreduce.amazonaws.com&Description=EMR+service+role"),
		Action:  "CreateServiceLinkedRole",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	body := string(resp.Body)
	require.Contains(t, body, "AWSServiceRoleFor")
	require.Contains(t, body, "/aws-service-role/elasticmapreduce.amazonaws.com/")
	require.Contains(t, body, "sts:AssumeRole")
}

func TestCreateServiceLinkedRole_WithCustomSuffix(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateServiceLinkedRole&AWSServiceName=autoscaling.amazonaws.com&CustomSuffix=MyApp"),
		Action:  "CreateServiceLinkedRole",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	body := string(resp.Body)
	require.Contains(t, body, "_MyApp")
}

func TestDeleteServiceLinkedRole_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a service-linked role
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateServiceLinkedRole&AWSServiceName=ecs.amazonaws.com"),
		Action:  "CreateServiceLinkedRole",
	}
	resp, _ := service.HandleRequest(context.Background(), createReq)

	// Extract the role name from the response
	body := string(resp.Body)
	start := strings.Index(body, "<RoleName>") + 10
	end := strings.Index(body, "</RoleName>")
	roleName := body[start:end]

	// Delete the service-linked role
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteServiceLinkedRole&RoleName=" + roleName),
		Action:  "DeleteServiceLinkedRole",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<DeletionTaskId>")
}

func TestDeleteServiceLinkedRole_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteServiceLinkedRole&RoleName=NonexistentRole"),
		Action:  "DeleteServiceLinkedRole",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestGetServiceLinkedRoleDeletionStatus_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create and delete a service-linked role to get a task ID
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateServiceLinkedRole&AWSServiceName=lambda.amazonaws.com"),
		Action:  "CreateServiceLinkedRole",
	}
	resp, _ := service.HandleRequest(context.Background(), createReq)
	body := string(resp.Body)
	start := strings.Index(body, "<RoleName>") + 10
	end := strings.Index(body, "</RoleName>")
	roleName := body[start:end]

	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteServiceLinkedRole&RoleName=" + roleName),
		Action:  "DeleteServiceLinkedRole",
	}
	resp, _ = service.HandleRequest(context.Background(), deleteReq)
	body = string(resp.Body)
	start = strings.Index(body, "<DeletionTaskId>") + 16
	end = strings.Index(body, "</DeletionTaskId>")
	taskId := body[start:end]

	// Get deletion status
	statusReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServiceLinkedRoleDeletionStatus&DeletionTaskId=" + taskId),
		Action:  "GetServiceLinkedRoleDeletionStatus",
	}
	resp, err := service.HandleRequest(context.Background(), statusReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<Status>SUCCEEDED</Status>")
}

func TestGetServiceLinkedRoleDeletionStatus_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetServiceLinkedRoleDeletionStatus&DeletionTaskId=nonexistent-task-id"),
		Action:  "GetServiceLinkedRoleDeletionStatus",
	}
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}
