package sts

import (
	"context"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

// ============================================================================
// GetCallerIdentity Tests
// ============================================================================

func TestGetCallerIdentity_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetCallerIdentity"),
		Action: "GetCallerIdentity",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "GetCallerIdentityResponse")

	// Verify response contains expected fields
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<UserId>") {
		t.Error("Response should contain UserId")
	}
	if !strings.Contains(bodyStr, "<Account>") {
		t.Error("Response should contain Account")
	}
	if !strings.Contains(bodyStr, "<Arn>") {
		t.Error("Response should contain Arn")
	}
	if !strings.Contains(bodyStr, "123456789012") {
		t.Error("Response should contain mock account ID")
	}
}

func TestGetCallerIdentity_XMLNamespace(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetCallerIdentity"),
		Action: "GetCallerIdentity",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	bodyStr := string(resp.Body)
	// Verify proper XML namespace for STS
	if !strings.Contains(bodyStr, "xmlns=") {
		t.Error("Response should contain XML namespace")
	}
}

// ============================================================================
// AssumeRole Tests (Not Implemented - returns 501)
// ============================================================================

func TestAssumeRole_NotImplemented(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=AssumeRole&RoleArn=arn:aws:iam::123456789012:role/test&RoleSessionName=test"),
		Action: "AssumeRole",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 501)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NotImplemented", emulator.ProtocolQuery)
}

// ============================================================================
// GetSessionToken Tests (Not Implemented - returns 501)
// ============================================================================

func TestGetSessionToken_NotImplemented(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetSessionToken"),
		Action: "GetSessionToken",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 501)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NotImplemented", emulator.ProtocolQuery)
}

// ============================================================================
// GetFederationToken Tests (Not Implemented - returns 501)
// ============================================================================

func TestGetFederationToken_NotImplemented(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetFederationToken&Name=testuser"),
		Action: "GetFederationToken",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 501)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "NotImplemented", emulator.ProtocolQuery)
}

// ============================================================================
// Invalid Action Tests
// ============================================================================

func TestInvalidAction(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=NonExistentAction"),
		Action: "NonExistentAction",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidAction", emulator.ProtocolQuery)
}

func TestMissingAction(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte(""),
		Action: "",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	// Missing action returns ValidationException from the validator
	testhelpers.AssertErrorResponse(t, resp, "ValidationException", emulator.ProtocolQuery)
}

// ============================================================================
// Response Format Tests
// ============================================================================

func TestResponseFormat_XMLDeclaration(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetCallerIdentity"),
		Action: "GetCallerIdentity",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	bodyStr := string(resp.Body)
	if !strings.HasPrefix(bodyStr, "<?xml version=") {
		t.Error("Response should start with XML declaration")
	}
}

func TestResponseFormat_ResponseMetadata(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=GetCallerIdentity"),
		Action: "GetCallerIdentity",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<ResponseMetadata>") {
		t.Error("Response should contain ResponseMetadata")
	}
	if !strings.Contains(bodyStr, "<RequestId>") {
		t.Error("Response should contain RequestId inside ResponseMetadata")
	}
}

func TestErrorResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewStsService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=InvalidAction"),
		Action: "InvalidAction",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	bodyStr := string(resp.Body)
	// Error responses should have proper structure
	if !strings.Contains(bodyStr, "<ErrorResponse") {
		t.Error("Error response should contain ErrorResponse element")
	}
	if !strings.Contains(bodyStr, "<Error>") {
		t.Error("Error response should contain Error element")
	}
	if !strings.Contains(bodyStr, "<Code>") {
		t.Error("Error response should contain Code element")
	}
	if !strings.Contains(bodyStr, "<Message>") {
		t.Error("Error response should contain Message element")
	}
	if !strings.Contains(bodyStr, "<RequestId>") {
		t.Error("Error response should contain RequestId")
	}
}
