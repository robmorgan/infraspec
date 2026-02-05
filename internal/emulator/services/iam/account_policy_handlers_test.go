package iam

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

// ============================================================================
// Account Alias Tests
// ============================================================================

func TestCreateAccountAlias_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=my-company"),
		Action:  "CreateAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
}

func TestCreateAccountAlias_MissingAlias(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias"),
		Action:  "CreateAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "AccountAlias is required")
}

func TestCreateAccountAlias_InvalidFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Test with uppercase (invalid)
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=MyCompany"),
		Action:  "CreateAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Invalid account alias")
}

func TestCreateAccountAlias_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=my-company"),
		Action:  "CreateAccountAlias",
	}

	// Create first alias
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create another
	req2 := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=another-company"),
		Action:  "CreateAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), req2)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestDeleteAccountAlias_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create alias first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=delete-test"),
		Action:  "CreateAccountAlias",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete it
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccountAlias&AccountAlias=delete-test"),
		Action:  "DeleteAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccountAliases"),
		Action:  "ListAccountAliases",
	}
	listResp, _ := service.HandleRequest(context.Background(), listReq)
	require.NotContains(t, string(listResp.Body), "delete-test")
}

func TestDeleteAccountAlias_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccountAlias&AccountAlias=nonexistent"),
		Action:  "DeleteAccountAlias",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListAccountAliases_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create alias
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateAccountAlias&AccountAlias=list-test-alias"),
		Action:  "CreateAccountAlias",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// List aliases
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccountAliases"),
		Action:  "ListAccountAliases",
	}

	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "<AccountAliases>")
	require.Contains(t, body, "list-test-alias")
}

func TestListAccountAliases_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListAccountAliases"),
		Action:  "ListAccountAliases",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<AccountAliases")
}

// ============================================================================
// Password Policy Tests
// ============================================================================

func TestUpdateAccountPasswordPolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccountPasswordPolicy&MinimumPasswordLength=12&RequireSymbols=true&RequireNumbers=true&RequireUppercaseCharacters=true&RequireLowercaseCharacters=true"),
		Action:  "UpdateAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestUpdateAccountPasswordPolicy_InvalidLength(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Too short
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccountPasswordPolicy&MinimumPasswordLength=3"),
		Action:  "UpdateAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "MinimumPasswordLength must be between 6 and 128")
}

func TestUpdateAccountPasswordPolicy_InvalidMaxAge(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccountPasswordPolicy&MaxPasswordAge=2000"),
		Action:  "UpdateAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "MaxPasswordAge must be between 1 and 1095")
}

func TestGetAccountPasswordPolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create policy first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccountPasswordPolicy&MinimumPasswordLength=14&RequireSymbols=true"),
		Action:  "UpdateAccountPasswordPolicy",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Get policy
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetAccountPasswordPolicy"),
		Action:  "GetAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "<MinimumPasswordLength>14</MinimumPasswordLength>")
	require.Contains(t, body, "<RequireSymbols>true</RequireSymbols>")
}

func TestGetAccountPasswordPolicy_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetAccountPasswordPolicy"),
		Action:  "GetAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestDeleteAccountPasswordPolicy_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create policy first
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateAccountPasswordPolicy&MinimumPasswordLength=10"),
		Action:  "UpdateAccountPasswordPolicy",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete policy
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccountPasswordPolicy"),
		Action:  "DeleteAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetAccountPasswordPolicy"),
		Action:  "GetAccountPasswordPolicy",
	}
	getResp, _ := service.HandleRequest(context.Background(), getReq)
	require.Equal(t, 404, getResp.StatusCode)
}

func TestDeleteAccountPasswordPolicy_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteAccountPasswordPolicy"),
		Action:  "DeleteAccountPasswordPolicy",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestIsValidAccountAlias(t *testing.T) {
	// Valid aliases
	require.True(t, isValidAccountAlias("mycompany"))
	require.True(t, isValidAccountAlias("my-company"))
	require.True(t, isValidAccountAlias("company123"))
	require.True(t, isValidAccountAlias("a12"))

	// Invalid aliases
	require.False(t, isValidAccountAlias("ab"))           // Too short
	require.False(t, isValidAccountAlias("MyCompany"))    // Uppercase
	require.False(t, isValidAccountAlias("123company"))   // Starts with number
	require.False(t, isValidAccountAlias("-company"))     // Starts with hyphen
	require.False(t, isValidAccountAlias("company_name")) // Underscore not allowed
}
