package iam

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// setupIntegrationTest creates a test server with the IAM service and returns an AWS SDK client
func setupIntegrationTest(t *testing.T) (*iam.Client, *httptest.Server, func()) {
	t.Helper()

	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create ResourceManager for graph-based dependency tracking
	rmConfig := graph.DefaultResourceManagerConfig()
	rm := graph.NewResourceManager(state, rmConfig)

	service := NewIAMServiceWithGraph(state, validator, rm)

	// Create test server that handles IAM requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		// Extract Action from form data
		var action string
		if values, err := url.ParseQuery(string(body)); err == nil {
			action = values.Get("Action")
		}

		awsReq := &emulator.AWSRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: make(map[string]string),
			Body:    body,
			Action:  action,
		}

		for key := range r.Header {
			awsReq.Headers[key] = r.Header.Get(key)
		}

		resp, err := service.HandleRequest(r.Context(), awsReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(resp.Body)
	}))

	// Create AWS SDK client pointing to test server
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	client := iam.NewFromConfig(cfg, func(o *iam.Options) {
		o.BaseEndpoint = aws.String(server.URL)
	})

	cleanup := func() {
		server.Close()
	}

	return client, server, cleanup
}

func TestIntegration_CreateAndGetRole(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	trustPolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "ec2.amazonaws.com"},
			"Action": "sts:AssumeRole"
		}]
	}`

	// Create a role
	createResult, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("test-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String("Test role for integration testing"),
		MaxSessionDuration:       aws.Int32(7200),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if createResult.Role == nil {
		t.Fatal("Expected Role in response")
	}

	if createResult.Role.RoleName == nil || *createResult.Role.RoleName != "test-role" {
		t.Errorf("Expected role name 'test-role', got %v", createResult.Role.RoleName)
	}

	if createResult.Role.Arn == nil || *createResult.Role.Arn == "" {
		t.Error("Expected Arn to be set")
	}

	if createResult.Role.RoleId == nil || *createResult.Role.RoleId == "" {
		t.Error("Expected RoleId to be set")
	}

	t.Logf("Created role: %s with ARN: %s", *createResult.Role.RoleName, *createResult.Role.Arn)

	// Get the role
	getResult, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("test-role"),
	})
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}

	if getResult.Role == nil {
		t.Fatal("Expected Role in GetRole response")
	}

	if *getResult.Role.RoleName != "test-role" {
		t.Errorf("Expected role name 'test-role', got %s", *getResult.Role.RoleName)
	}
}

func TestIntegration_CreateAndDeleteRole(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	trustPolicy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`

	// Create a role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("delete-test-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Delete the role
	_, err = client.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String("delete-test-role"),
	})
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	// Verify role is deleted
	_, err = client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("delete-test-role"),
	})
	if err == nil {
		t.Error("Expected error when getting deleted role")
	}
}

func TestIntegration_ListRoles(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`

	// Create multiple roles
	for i := 1; i <= 3; i++ {
		_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
			RoleName:                 aws.String("list-test-role-" + string(rune('0'+i))),
			AssumeRolePolicyDocument: aws.String(trustPolicy),
		})
		if err != nil {
			t.Fatalf("CreateRole %d failed: %v", i, err)
		}
	}

	// List roles
	listResult, err := client.ListRoles(ctx, &iam.ListRolesInput{})
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}

	if len(listResult.Roles) < 3 {
		t.Errorf("Expected at least 3 roles, got %d", len(listResult.Roles))
	}
}

func TestIntegration_CreateAndGetPolicy(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "s3:GetObject",
			"Resource": "*"
		}]
	}`

	// Create a policy
	createResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String("test-policy"),
		PolicyDocument: aws.String(policyDocument),
		Description:    aws.String("Test policy for integration testing"),
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}

	if createResult.Policy == nil {
		t.Fatal("Expected Policy in response")
	}

	if createResult.Policy.PolicyName == nil || *createResult.Policy.PolicyName != "test-policy" {
		t.Errorf("Expected policy name 'test-policy', got %v", createResult.Policy.PolicyName)
	}

	policyArn := createResult.Policy.Arn
	t.Logf("Created policy: %s with ARN: %s", *createResult.Policy.PolicyName, *policyArn)

	// Get the policy
	getResult, err := client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("GetPolicy failed: %v", err)
	}

	if *getResult.Policy.PolicyName != "test-policy" {
		t.Errorf("Expected policy name 'test-policy', got %s", *getResult.Policy.PolicyName)
	}
}

func TestIntegration_AttachDetachRolePolicy(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a role
	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("attach-test-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Create a policy
	policyDocument := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "*", "Resource": "*"}]}`
	createPolicyResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String("attach-test-policy"),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}

	policyArn := createPolicyResult.Policy.Arn

	// Attach policy to role
	_, err = client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String("attach-test-role"),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("AttachRolePolicy failed: %v", err)
	}

	// List attached policies
	listResult, err := client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String("attach-test-role"),
	})
	if err != nil {
		t.Fatalf("ListAttachedRolePolicies failed: %v", err)
	}

	if len(listResult.AttachedPolicies) != 1 {
		t.Errorf("Expected 1 attached policy, got %d", len(listResult.AttachedPolicies))
	}

	// Detach policy
	_, err = client.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String("attach-test-role"),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("DetachRolePolicy failed: %v", err)
	}

	// Verify policy is detached
	listResult, err = client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String("attach-test-role"),
	})
	if err != nil {
		t.Fatalf("ListAttachedRolePolicies failed: %v", err)
	}

	if len(listResult.AttachedPolicies) != 0 {
		t.Errorf("Expected 0 attached policies after detach, got %d", len(listResult.AttachedPolicies))
	}
}

func TestIntegration_CreateAndGetInstanceProfile(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create an instance profile
	createResult, err := client.CreateInstanceProfile(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String("test-instance-profile"),
	})
	if err != nil {
		t.Fatalf("CreateInstanceProfile failed: %v", err)
	}

	if createResult.InstanceProfile == nil {
		t.Fatal("Expected InstanceProfile in response")
	}

	if createResult.InstanceProfile.InstanceProfileName == nil || *createResult.InstanceProfile.InstanceProfileName != "test-instance-profile" {
		t.Errorf("Expected profile name 'test-instance-profile', got %v", createResult.InstanceProfile.InstanceProfileName)
	}

	t.Logf("Created instance profile: %s", *createResult.InstanceProfile.InstanceProfileName)

	// Get the instance profile
	getResult, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String("test-instance-profile"),
	})
	if err != nil {
		t.Fatalf("GetInstanceProfile failed: %v", err)
	}

	if *getResult.InstanceProfile.InstanceProfileName != "test-instance-profile" {
		t.Errorf("Expected profile name 'test-instance-profile', got %s", *getResult.InstanceProfile.InstanceProfileName)
	}
}

func TestIntegration_AddRemoveRoleFromInstanceProfile(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a role
	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("profile-test-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Create an instance profile
	_, err = client.CreateInstanceProfile(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String("profile-test-profile"),
	})
	if err != nil {
		t.Fatalf("CreateInstanceProfile failed: %v", err)
	}

	// Add role to instance profile
	_, err = client.AddRoleToInstanceProfile(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String("profile-test-profile"),
		RoleName:            aws.String("profile-test-role"),
	})
	if err != nil {
		t.Fatalf("AddRoleToInstanceProfile failed: %v", err)
	}

	// Verify role is added
	getResult, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String("profile-test-profile"),
	})
	if err != nil {
		t.Fatalf("GetInstanceProfile failed: %v", err)
	}

	if len(getResult.InstanceProfile.Roles) != 1 {
		t.Errorf("Expected 1 role in instance profile, got %d", len(getResult.InstanceProfile.Roles))
	}

	// Remove role from instance profile
	_, err = client.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String("profile-test-profile"),
		RoleName:            aws.String("profile-test-role"),
	})
	if err != nil {
		t.Fatalf("RemoveRoleFromInstanceProfile failed: %v", err)
	}

	// Verify role is removed
	getResult, err = client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String("profile-test-profile"),
	})
	if err != nil {
		t.Fatalf("GetInstanceProfile failed: %v", err)
	}

	if len(getResult.InstanceProfile.Roles) != 0 {
		t.Errorf("Expected 0 roles in instance profile after removal, got %d", len(getResult.InstanceProfile.Roles))
	}
}

func TestIntegration_DeleteRoleWithAttachedPolicies(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a role
	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("delete-conflict-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Create and attach a policy
	policyDocument := `{"Version": "2012-10-17", "Statement": []}`
	createPolicyResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String("delete-conflict-policy"),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}

	_, err = client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String("delete-conflict-role"),
		PolicyArn: createPolicyResult.Policy.Arn,
	})
	if err != nil {
		t.Fatalf("AttachRolePolicy failed: %v", err)
	}

	// Try to delete role (should fail due to attached policy)
	_, err = client.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String("delete-conflict-role"),
	})
	if err == nil {
		t.Error("Expected error when deleting role with attached policies")
	}
}

func TestIntegration_RoleNotFound(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Try to get a non-existent role
	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("non-existent-role"),
	})
	if err == nil {
		t.Error("Expected error when getting non-existent role")
	}
}

func TestIntegration_DuplicateRole(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	trustPolicy := `{"Version": "2012-10-17", "Statement": []}`

	// Create a role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("duplicate-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err != nil {
		t.Fatalf("First CreateRole failed: %v", err)
	}

	// Try to create the same role again
	_, err = client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("duplicate-role"),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	})
	if err == nil {
		t.Error("Expected error when creating duplicate role")
	}
}

func TestIntegration_GetPolicyVersion(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	policyDocument := `{"Version": "2012-10-17", "Statement": []}`

	// Create a policy
	createResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String("version-test-policy"),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}

	// Get the policy version
	versionResult, err := client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
		PolicyArn: createResult.Policy.Arn,
		VersionId: aws.String("v1"),
	})
	if err != nil {
		t.Fatalf("GetPolicyVersion failed: %v", err)
	}

	if versionResult.PolicyVersion == nil {
		t.Fatal("Expected PolicyVersion in response")
	}

	if *versionResult.PolicyVersion.VersionId != "v1" {
		t.Errorf("Expected version 'v1', got %s", *versionResult.PolicyVersion.VersionId)
	}

	if !versionResult.PolicyVersion.IsDefaultVersion {
		t.Error("Expected v1 to be default version")
	}
}

func TestIntegration_UpdateAssumeRolePolicy(t *testing.T) {
	client, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	originalPolicy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "ec2.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`
	updatedPolicy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`

	// Create a role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("update-policy-role"),
		AssumeRolePolicyDocument: aws.String(originalPolicy),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Update the assume role policy
	_, err = client.UpdateAssumeRolePolicy(ctx, &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String("update-policy-role"),
		PolicyDocument: aws.String(updatedPolicy),
	})
	if err != nil {
		t.Fatalf("UpdateAssumeRolePolicy failed: %v", err)
	}

	// Verify the update
	getResult, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("update-policy-role"),
	})
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}

	if getResult.Role.AssumeRolePolicyDocument == nil {
		t.Error("Expected AssumeRolePolicyDocument to be set")
	}
}
