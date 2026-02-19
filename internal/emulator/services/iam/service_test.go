package iam

import (
	"context"
	"strings"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

// Helper function to create a role for testing
func createTestRole(t *testing.T, service *IAMService, roleName string) {
	t.Helper()
	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateRole&RoleName=" + roleName + "&AssumeRolePolicyDocument=" + trustPolicy),
		Action: "CreateRole",
	}
	_, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}
}

// ============================================================================
// PutRolePolicy Tests
// ============================================================================

func TestPutRolePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role first
	createTestRole(t, service, "test-role")

	// Put an inline policy
	policyDocument := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyName=test-policy&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "PutRolePolicyResponse")
}

func TestPutRolePolicy_RoleNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=nonexistent-role&PolicyName=test-policy&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

func TestPutRolePolicy_MissingRoleName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&PolicyName=test-policy&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidInput", emulator.ProtocolQuery)
}

func TestPutRolePolicy_MissingPolicyName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	createTestRole(t, service, "test-role")

	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidInput", emulator.ProtocolQuery)
}

func TestPutRolePolicy_MissingPolicyDocument(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	createTestRole(t, service, "test-role")

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyName=test-policy"),
		Action: "PutRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidInput", emulator.ProtocolQuery)
}

// ============================================================================
// GetRolePolicy Tests
// ============================================================================

func TestGetRolePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role and put an inline policy
	createTestRole(t, service, "test-role")

	policyDocument := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`
	putReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyName=test-policy&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get the inline policy
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetRolePolicy&RoleName=test-role&PolicyName=test-policy"),
		Action: "GetRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "GetRolePolicyResponse")

	// Verify the response contains the policy data
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<RoleName>test-role</RoleName>") {
		t.Error("Response should contain RoleName")
	}
	if !strings.Contains(bodyStr, "<PolicyName>test-policy</PolicyName>") {
		t.Error("Response should contain PolicyName")
	}
	if !strings.Contains(bodyStr, "<PolicyDocument>") {
		t.Error("Response should contain PolicyDocument")
	}
}

func TestGetRolePolicy_RoleNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetRolePolicy&RoleName=nonexistent-role&PolicyName=test-policy"),
		Action: "GetRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

func TestGetRolePolicy_PolicyNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	createTestRole(t, service, "test-role")

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetRolePolicy&RoleName=test-role&PolicyName=nonexistent-policy"),
		Action: "GetRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

// ============================================================================
// DeleteRolePolicy Tests
// ============================================================================

func TestDeleteRolePolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role and put an inline policy
	createTestRole(t, service, "test-role")

	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	putReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyName=test-policy&PolicyDocument=" + policyDocument),
		Action: "PutRolePolicy",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Delete the inline policy
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteRolePolicy&RoleName=test-role&PolicyName=test-policy"),
		Action: "DeleteRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "DeleteRolePolicyResponse")

	// Verify the policy is actually deleted
	getReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetRolePolicy&RoleName=test-role&PolicyName=test-policy"),
		Action: "GetRolePolicy",
	}
	getResp, _ := service.HandleRequest(context.Background(), getReq)
	if getResp.StatusCode != 404 {
		t.Error("Expected policy to be deleted")
	}
}

func TestDeleteRolePolicy_RoleNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteRolePolicy&RoleName=nonexistent-role&PolicyName=test-policy"),
		Action: "DeleteRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

func TestDeleteRolePolicy_PolicyNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	createTestRole(t, service, "test-role")

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteRolePolicy&RoleName=test-role&PolicyName=nonexistent-policy"),
		Action: "DeleteRolePolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

// ============================================================================
// ListRolePolicies Tests
// ============================================================================

func TestListRolePolicies_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role and put multiple inline policies
	createTestRole(t, service, "test-role")

	policyDocument := `{"Version":"2012-10-17","Statement":[]}`
	for _, policyName := range []string{"policy-1", "policy-2", "policy-3"} {
		putReq := &emulator.AWSRequest{
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
			Body:   []byte("Action=PutRolePolicy&RoleName=test-role&PolicyName=" + policyName + "&PolicyDocument=" + policyDocument),
			Action: "PutRolePolicy",
		}
		_, _ = service.HandleRequest(context.Background(), putReq)
	}

	// List inline policies
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListRolePolicies&RoleName=test-role"),
		Action: "ListRolePolicies",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "ListRolePoliciesResponse")

	// Verify the response contains the policy names
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<member>policy-1</member>") {
		t.Error("Response should contain policy-1")
	}
	if !strings.Contains(bodyStr, "<member>policy-2</member>") {
		t.Error("Response should contain policy-2")
	}
	if !strings.Contains(bodyStr, "<member>policy-3</member>") {
		t.Error("Response should contain policy-3")
	}
	if !strings.Contains(bodyStr, "<IsTruncated>false</IsTruncated>") {
		t.Error("Response should contain IsTruncated false")
	}
}

func TestListRolePolicies_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create a role without any inline policies
	createTestRole(t, service, "test-role")

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListRolePolicies&RoleName=test-role"),
		Action: "ListRolePolicies",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "ListRolePoliciesResponse")

	// Verify empty list
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<IsTruncated>false</IsTruncated>") {
		t.Error("Response should contain IsTruncated false")
	}
}

func TestListRolePolicies_RoleNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListRolePolicies&RoleName=nonexistent-role"),
		Action: "ListRolePolicies",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NoSuchEntity", emulator.ProtocolQuery)
}

func TestListRolePolicies_MissingRoleName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListRolePolicies"),
		Action: "ListRolePolicies",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidInput", emulator.ProtocolQuery)
}
