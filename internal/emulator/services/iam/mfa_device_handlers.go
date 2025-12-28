package iam

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// Virtual MFA Device Operations
// ============================================================================

// createVirtualMFADevice creates a new virtual MFA device
func (s *IAMService) createVirtualMFADevice(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	virtualMFADeviceName := emulator.GetStringParam(params, "VirtualMFADeviceName", "")
	if virtualMFADeviceName == "" {
		return s.errorResponse(400, "ValidationError", "VirtualMFADeviceName is required"), nil
	}

	// Validate name format
	if !isValidMFADeviceName(virtualMFADeviceName) {
		return s.errorResponse(400, "ValidationError", "Invalid virtual MFA device name. Must be alphanumeric with +=,.@_- characters, 1-128 chars"), nil
	}

	path := emulator.GetStringParam(params, "Path", "/")
	if !isValidPath(path) {
		return s.errorResponse(400, "ValidationError", "Invalid path"), nil
	}

	// Generate serial number
	serialNumber := fmt.Sprintf("arn:aws:iam::%s:mfa/%s%s", defaultAccountID, path[1:], virtualMFADeviceName)

	// Check if device already exists
	stateKey := fmt.Sprintf("iam:mfa-device:%s", serialNumber)
	var existing VirtualMFADeviceData
	if err := s.state.Get(stateKey, &existing); err == nil {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Virtual MFA device %s already exists", virtualMFADeviceName)), nil
	}

	// Generate base32 seed (20 bytes = 160 bits)
	seed := make([]byte, 20)
	if _, err := rand.Read(seed); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to generate MFA seed"), nil
	}
	base32Seed := base32.StdEncoding.EncodeToString(seed)

	// Generate a placeholder QR code PNG (in real AWS this would be an actual QR code)
	// We'll just base64 encode a small placeholder
	qrCodePNG := base64.StdEncoding.EncodeToString([]byte("QR_CODE_PLACEHOLDER"))

	// Parse tags if provided
	tags := s.parseTags(params)

	now := time.Now().UTC()
	device := VirtualMFADeviceData{
		SerialNumber:     serialNumber,
		Base32StringSeed: base32Seed,
		QRCodePNG:        qrCodePNG,
		Tags:             tags,
		CreateDate:       now,
	}

	if err := s.state.Set(stateKey, device); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to create virtual MFA device"), nil
	}

	result := CreateVirtualMFADeviceResult{
		VirtualMFADevice: XMLVirtualMFADevice{
			SerialNumber:     serialNumber,
			Base32StringSeed: base32Seed,
			QRCodePNG:        qrCodePNG,
			Tags:             tags,
		},
	}

	return s.successResponse("CreateVirtualMFADevice", result)
}

// deleteVirtualMFADevice deletes a virtual MFA device
func (s *IAMService) deleteVirtualMFADevice(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	serialNumber := emulator.GetStringParam(params, "SerialNumber", "")
	if serialNumber == "" {
		return s.errorResponse(400, "ValidationError", "SerialNumber is required"), nil
	}

	stateKey := fmt.Sprintf("iam:mfa-device:%s", serialNumber)
	var device VirtualMFADeviceData
	if err := s.state.Get(stateKey, &device); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Virtual MFA device %s not found", serialNumber)), nil
	}

	// Cannot delete if still assigned to a user
	if device.UserName != "" {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete virtual MFA device %s because it is still assigned to user %s", serialNumber, device.UserName)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to delete virtual MFA device"), nil
	}

	return s.successResponse("DeleteVirtualMFADevice", EmptyResult{})
}

// listVirtualMFADevices lists all virtual MFA devices
func (s *IAMService) listVirtualMFADevices(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	assignmentStatus := emulator.GetStringParam(params, "AssignmentStatus", "")

	keys, err := s.state.List("iam:mfa-device:")
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list virtual MFA devices"), nil
	}

	var devices []VirtualMFADeviceListItem
	for _, key := range keys {
		var device VirtualMFADeviceData
		if err := s.state.Get(key, &device); err != nil {
			continue
		}

		// Filter by assignment status
		if assignmentStatus == "Assigned" && device.UserName == "" {
			continue
		}
		if assignmentStatus == "Unassigned" && device.UserName != "" {
			continue
		}

		item := VirtualMFADeviceListItem{
			SerialNumber: device.SerialNumber,
			EnableDate:   device.EnableDate,
			Tags:         device.Tags,
		}

		// Include user info if assigned
		if device.UserName != "" {
			var user XMLUser
			userKey := fmt.Sprintf("iam:user:%s", device.UserName)
			if err := s.state.Get(userKey, &user); err == nil {
				item.User = &user
			}
		}

		devices = append(devices, item)
	}

	result := ListVirtualMFADevicesResult{
		VirtualMFADevices: devices,
		IsTruncated:       false,
	}

	return s.successResponse("ListVirtualMFADevices", result)
}

// ============================================================================
// MFA Device Operations (for users)
// ============================================================================

// enableMFADevice enables an MFA device for a user
func (s *IAMService) enableMFADevice(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	serialNumber := emulator.GetStringParam(params, "SerialNumber", "")
	if serialNumber == "" {
		return s.errorResponse(400, "ValidationError", "SerialNumber is required"), nil
	}

	authenticationCode1 := emulator.GetStringParam(params, "AuthenticationCode1", "")
	authenticationCode2 := emulator.GetStringParam(params, "AuthenticationCode2", "")
	if authenticationCode1 == "" || authenticationCode2 == "" {
		return s.errorResponse(400, "ValidationError", "AuthenticationCode1 and AuthenticationCode2 are required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	// Verify MFA device exists
	deviceKey := fmt.Sprintf("iam:mfa-device:%s", serialNumber)
	var device VirtualMFADeviceData
	if err := s.state.Get(deviceKey, &device); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("MFA device %s not found", serialNumber)), nil
	}

	// Check if device is already assigned
	if device.UserName != "" {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("MFA device %s is already assigned to user %s", serialNumber, device.UserName)), nil
	}

	// In a real implementation, we would validate the authentication codes against the TOTP seed
	// For the emulator, we just accept any 6-digit codes
	if !isValidTOTPCode(authenticationCode1) || !isValidTOTPCode(authenticationCode2) {
		return s.errorResponse(400, "InvalidAuthenticationCode", "Authentication codes must be 6-digit numbers"), nil
	}

	// Assign device to user
	device.UserName = userName
	device.EnableDate = time.Now().UTC()

	if err := s.state.Set(deviceKey, device); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to enable MFA device"), nil
	}

	return s.successResponse("EnableMFADevice", EmptyResult{})
}

// deactivateMFADevice deactivates an MFA device for a user
func (s *IAMService) deactivateMFADevice(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	serialNumber := emulator.GetStringParam(params, "SerialNumber", "")
	if serialNumber == "" {
		return s.errorResponse(400, "ValidationError", "SerialNumber is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	// Verify MFA device exists
	deviceKey := fmt.Sprintf("iam:mfa-device:%s", serialNumber)
	var device VirtualMFADeviceData
	if err := s.state.Get(deviceKey, &device); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("MFA device %s not found", serialNumber)), nil
	}

	// Verify device is assigned to this user
	if device.UserName != userName {
		return s.errorResponse(400, "ValidationError", fmt.Sprintf("MFA device %s is not assigned to user %s", serialNumber, userName)), nil
	}

	// Deactivate the device (unassign from user)
	device.UserName = ""
	device.EnableDate = time.Time{}

	if err := s.state.Set(deviceKey, device); err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to deactivate MFA device"), nil
	}

	return s.successResponse("DeactivateMFADevice", EmptyResult{})
}

// listMFADevices lists MFA devices for a user
func (s *IAMService) listMFADevices(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	// Find all MFA devices assigned to this user
	keys, err := s.state.List("iam:mfa-device:")
	if err != nil {
		return s.errorResponse(500, "ServiceFailure", "Failed to list MFA devices"), nil
	}

	var devices []MFADeviceListItem
	for _, key := range keys {
		var device VirtualMFADeviceData
		if err := s.state.Get(key, &device); err != nil {
			continue
		}

		if device.UserName == userName {
			devices = append(devices, MFADeviceListItem{
				SerialNumber: device.SerialNumber,
				UserName:     userName,
				EnableDate:   device.EnableDate,
			})
		}
	}

	result := ListMFADevicesResult{
		MFADevices:  devices,
		IsTruncated: false,
	}

	return s.successResponse("ListMFADevices", result)
}

// resyncMFADevice resynchronizes an MFA device
func (s *IAMService) resyncMFADevice(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := emulator.GetStringParam(params, "UserName", "")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	serialNumber := emulator.GetStringParam(params, "SerialNumber", "")
	if serialNumber == "" {
		return s.errorResponse(400, "ValidationError", "SerialNumber is required"), nil
	}

	authenticationCode1 := emulator.GetStringParam(params, "AuthenticationCode1", "")
	authenticationCode2 := emulator.GetStringParam(params, "AuthenticationCode2", "")
	if authenticationCode1 == "" || authenticationCode2 == "" {
		return s.errorResponse(400, "ValidationError", "AuthenticationCode1 and AuthenticationCode2 are required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	var user XMLUser
	if err := s.state.Get(userKey, &user); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("User %s not found", userName)), nil
	}

	// Verify MFA device exists
	deviceKey := fmt.Sprintf("iam:mfa-device:%s", serialNumber)
	var device VirtualMFADeviceData
	if err := s.state.Get(deviceKey, &device); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("MFA device %s not found", serialNumber)), nil
	}

	// Verify device is assigned to this user
	if device.UserName != userName {
		return s.errorResponse(400, "ValidationError", fmt.Sprintf("MFA device %s is not assigned to user %s", serialNumber, userName)), nil
	}

	// Validate codes
	if !isValidTOTPCode(authenticationCode1) || !isValidTOTPCode(authenticationCode2) {
		return s.errorResponse(400, "InvalidAuthenticationCode", "Authentication codes must be 6-digit numbers"), nil
	}

	// In a real implementation, we would resync the TOTP counter
	// For the emulator, we just accept valid codes

	return s.successResponse("ResyncMFADevice", EmptyResult{})
}

// ============================================================================
// Helper functions
// ============================================================================

// isValidMFADeviceName validates the MFA device name format
func isValidMFADeviceName(name string) bool {
	if len(name) < 1 || len(name) > 128 {
		return false
	}
	// Must be alphanumeric with +=,.@_- allowed
	matched, _ := regexp.MatchString(`^[\w+=,.@-]+$`, name)
	return matched
}

// isValidPath validates an IAM path
func isValidPath(path string) bool {
	if len(path) < 1 || len(path) > 512 {
		return false
	}
	if path[0] != '/' {
		return false
	}
	if path[len(path)-1] != '/' && len(path) > 1 {
		return false
	}
	// Path can only contain alphanumeric chars and /
	matched, _ := regexp.MatchString(`^/[\w/]*$`, path)
	return matched
}

// isValidTOTPCode validates a 6-digit TOTP code
func isValidTOTPCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	matched, _ := regexp.MatchString(`^[0-9]{6}$`, code)
	return matched
}
