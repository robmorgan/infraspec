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
// Virtual MFA Device Tests
// ============================================================================

func TestCreateVirtualMFADevice_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=TestMFA"),
		Action:  "CreateVirtualMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<SerialNumber>arn:aws:iam::123456789012:mfa/TestMFA</SerialNumber>")
	require.Contains(t, body, "<Base32StringSeed>")
	require.Contains(t, body, "<QRCodePNG>")
}

func TestCreateVirtualMFADevice_MissingName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice"),
		Action:  "CreateVirtualMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "VirtualMFADeviceName is required")
}

func TestCreateVirtualMFADevice_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=DuplicateMFA"),
		Action:  "CreateVirtualMFADevice",
	}

	// Create first
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create again
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestDeleteVirtualMFADevice_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create device
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=DeleteMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete device
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteVirtualMFADevice&SerialNumber=arn:aws:iam::123456789012:mfa/DeleteMFA"),
		Action:  "DeleteVirtualMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestDeleteVirtualMFADevice_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteVirtualMFADevice&SerialNumber=arn:aws:iam::123456789012:mfa/NonexistentMFA"),
		Action:  "DeleteVirtualMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListVirtualMFADevices_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create two devices
	for _, name := range []string{"MFA1", "MFA2"} {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=" + name),
			Action:  "CreateVirtualMFADevice",
		}
		_, _ = service.HandleRequest(context.Background(), req)
	}

	// List devices
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListVirtualMFADevices"),
		Action:  "ListVirtualMFADevices",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "MFA1")
	require.Contains(t, body, "MFA2")
}

func TestListVirtualMFADevices_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListVirtualMFADevices"),
		Action:  "ListVirtualMFADevices",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<VirtualMFADevices")
}

// ============================================================================
// MFA Device User Operations Tests
// ============================================================================

func TestEnableMFADevice_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=MFAUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=EnableTestMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	// Enable MFA device for user
	enableReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=MFAUser&SerialNumber=arn:aws:iam::123456789012:mfa/EnableTestMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), enableReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestEnableMFADevice_UserNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=NonexistentUser&SerialNumber=arn:aws:iam::123456789012:mfa/SomeMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestEnableMFADevice_DeviceNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=MFAUser2"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=MFAUser2&SerialNumber=arn:aws:iam::123456789012:mfa/NonexistentMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestEnableMFADevice_InvalidCode(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=MFAUser3"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=InvalidCodeMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=MFAUser3&SerialNumber=arn:aws:iam::123456789012:mfa/InvalidCodeMFA&AuthenticationCode1=invalid&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "InvalidAuthenticationCode")
}

func TestDeactivateMFADevice_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=DeactivateUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create and enable MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=DeactivateTestMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	enableReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=DeactivateUser&SerialNumber=arn:aws:iam::123456789012:mfa/DeactivateTestMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), enableReq)

	// Deactivate the device
	deactivateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeactivateMFADevice&UserName=DeactivateUser&SerialNumber=arn:aws:iam::123456789012:mfa/DeactivateTestMFA"),
		Action:  "DeactivateMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), deactivateReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestDeleteVirtualMFADevice_WhileAssigned_Fails(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=AssignedMFAUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create and enable MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=AssignedMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	enableReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=AssignedMFAUser&SerialNumber=arn:aws:iam::123456789012:mfa/AssignedMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), enableReq)

	// Try to delete while still assigned
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteVirtualMFADevice&SerialNumber=arn:aws:iam::123456789012:mfa/AssignedMFA"),
		Action:  "DeleteVirtualMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "DeleteConflict")
}

func TestListMFADevices_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=ListMFAUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create and enable MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=ListTestMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	enableReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=ListMFAUser&SerialNumber=arn:aws:iam::123456789012:mfa/ListTestMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), enableReq)

	// List MFA devices for user
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListMFADevices&UserName=ListMFAUser"),
		Action:  "ListMFADevices",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "ListTestMFA")
	require.Contains(t, body, "ListMFAUser")
}

func TestListMFADevices_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user without MFA
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=NoMFAUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// List MFA devices for user
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListMFADevices&UserName=NoMFAUser"),
		Action:  "ListMFADevices",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<MFADevices")
}

func TestResyncMFADevice_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=ResyncUser"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create and enable MFA device
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=ResyncTestMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	enableReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=EnableMFADevice&UserName=ResyncUser&SerialNumber=arn:aws:iam::123456789012:mfa/ResyncTestMFA&AuthenticationCode1=123456&AuthenticationCode2=789012"),
		Action:  "EnableMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), enableReq)

	// Resync device
	resyncReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ResyncMFADevice&UserName=ResyncUser&SerialNumber=arn:aws:iam::123456789012:mfa/ResyncTestMFA&AuthenticationCode1=234567&AuthenticationCode2=890123"),
		Action:  "ResyncMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), resyncReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestResyncMFADevice_NotAssigned(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create user
	createUserReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateUser&UserName=ResyncUser2"),
		Action:  "CreateUser",
	}
	_, _ = service.HandleRequest(context.Background(), createUserReq)

	// Create MFA device but don't enable it for user
	createMFAReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateVirtualMFADevice&VirtualMFADeviceName=UnassignedMFA"),
		Action:  "CreateVirtualMFADevice",
	}
	_, _ = service.HandleRequest(context.Background(), createMFAReq)

	// Try to resync
	resyncReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ResyncMFADevice&UserName=ResyncUser2&SerialNumber=arn:aws:iam::123456789012:mfa/UnassignedMFA&AuthenticationCode1=234567&AuthenticationCode2=890123"),
		Action:  "ResyncMFADevice",
	}
	resp, err := service.HandleRequest(context.Background(), resyncReq)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestIsValidMFADeviceName(t *testing.T) {
	// Valid names
	require.True(t, isValidMFADeviceName("MyDevice"))
	require.True(t, isValidMFADeviceName("My_Device"))
	require.True(t, isValidMFADeviceName("My-Device"))
	require.True(t, isValidMFADeviceName("user@example.com"))
	require.True(t, isValidMFADeviceName("device+tag"))
	require.True(t, isValidMFADeviceName("a"))

	// Invalid names
	require.False(t, isValidMFADeviceName(""))
	require.False(t, isValidMFADeviceName(strings.Repeat("a", 129)))
}

func TestIsValidTOTPCode(t *testing.T) {
	// Valid codes
	require.True(t, isValidTOTPCode("123456"))
	require.True(t, isValidTOTPCode("000000"))
	require.True(t, isValidTOTPCode("999999"))

	// Invalid codes
	require.False(t, isValidTOTPCode("12345"))   // too short
	require.False(t, isValidTOTPCode("1234567")) // too long
	require.False(t, isValidTOTPCode("abcdef"))  // not numeric
	require.False(t, isValidTOTPCode(""))        // empty
}
