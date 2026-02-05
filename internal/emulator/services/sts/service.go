package sts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

type StsService struct {
	state     emulator.StateManager
	validator emulator.Validator
}

func NewStsService(state emulator.StateManager, validator emulator.Validator) *StsService {
	return &StsService{
		state:     state,
		validator: validator,
	}
}

func (s *StsService) ServiceName() string {
	return "sts"
}

// SupportedActions returns the list of AWS API actions this service handles.
// Used by the router to determine which service handles a given Query Protocol request.
func (s *StsService) SupportedActions() []string {
	return []string{
		"AssumeRole",
		"AssumeRoleWithSAML",
		"AssumeRoleWithWebIdentity",
		"AssumeRoot",
		"DecodeAuthorizationMessage",
		"GetAccessKeyInfo",
		"GetCallerIdentity",
		"GetDelegatedAccessToken",
		"GetFederationToken",
		"GetSessionToken",
	}
}

func (s *StsService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	params, err := s.parseParameters(req)
	if err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	if err := s.validator.ValidateAction(action, params); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	switch action {
	case "AssumeRole":
		return s.assumeRole(ctx, params)
	case "AssumeRoleWithSAML":
		return s.assumeRoleWithSAML(ctx, params)
	case "AssumeRoleWithWebIdentity":
		return s.assumeRoleWithWebIdentity(ctx, params)
	case "AssumeRoot":
		return s.assumeRoot(ctx, params)
	case "DecodeAuthorizationMessage":
		return s.decodeAuthorizationMessage(ctx, params)
	case "GetAccessKeyInfo":
		return s.getAccessKeyInfo(ctx, params)
	case "GetCallerIdentity":
		return s.getCallerIdentity(ctx, params)
	case "GetDelegatedAccessToken":
		return s.getDelegatedAccessToken(ctx, params)
	case "GetFederationToken":
		return s.getFederationToken(ctx, params)
	case "GetSessionToken":
		return s.getSessionToken(ctx, params)
	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *StsService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *StsService) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
	if req.Parameters != nil {
		return req.Parameters, nil
	}

	contentType := req.Headers["Content-Type"]
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return s.parseFormData(string(req.Body))
	}

	if strings.Contains(contentType, "application/json") {
		var params map[string]interface{}
		if err := json.Unmarshal(req.Body, &params); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %w", err)
		}
		return params, nil
	}

	return make(map[string]interface{}), nil
}

func (s *StsService) parseFormData(body string) (map[string]interface{}, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			params[key] = vals[0]
		} else {
			params[key] = vals
		}
	}

	return params, nil
}

func (s *StsService) assumeRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement AssumeRole
	// Required parameter: AssumeRole (map[string]interface{}) - Input for AssumeRole

	return s.errorResponse(501, "NotImplemented", "AssumeRole is not yet implemented"), nil
}

func (s *StsService) assumeRoleWithSAML(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement AssumeRoleWithSAML
	// Required parameter: AssumeRoleWithSAML (map[string]interface{}) - Input for AssumeRoleWithSAML

	return s.errorResponse(501, "NotImplemented", "AssumeRoleWithSAML is not yet implemented"), nil
}

func (s *StsService) assumeRoleWithWebIdentity(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement AssumeRoleWithWebIdentity
	// Required parameter: AssumeRoleWithWebIdentity (map[string]interface{}) - Input for AssumeRoleWithWebIdentity

	return s.errorResponse(501, "NotImplemented", "AssumeRoleWithWebIdentity is not yet implemented"), nil
}

func (s *StsService) assumeRoot(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement AssumeRoot
	// Required parameter: AssumeRoot (map[string]interface{}) - Input for AssumeRoot

	return s.errorResponse(501, "NotImplemented", "AssumeRoot is not yet implemented"), nil
}

func (s *StsService) decodeAuthorizationMessage(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement DecodeAuthorizationMessage
	// Required parameter: DecodeAuthorizationMessage (map[string]interface{}) - Input for DecodeAuthorizationMessage

	return s.errorResponse(501, "NotImplemented", "DecodeAuthorizationMessage is not yet implemented"), nil
}

func (s *StsService) getAccessKeyInfo(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement GetAccessKeyInfo
	// Required parameter: GetAccessKeyInfo (map[string]interface{}) - Input for GetAccessKeyInfo

	return s.errorResponse(501, "NotImplemented", "GetAccessKeyInfo is not yet implemented"), nil
}

func (s *StsService) getCallerIdentity(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// GetCallerIdentity returns information about the IAM identity making the request
	// This is used by Terraform and AWS SDK to validate credentials

	// Create a mock caller identity response
	accountID := "123456789012"                  // Mock AWS account ID
	userID := "AIDAI" + uuid.New().String()[:13] // Mock user ID
	arn := fmt.Sprintf("arn:aws:iam::%s:user/infraspec-emulator", accountID)

	return s.successResponse("GetCallerIdentity", GetCallerIdentityResponse{
		UserId:  &userID,
		Account: &accountID,
		Arn:     &arn,
	})
}

func (s *StsService) getDelegatedAccessToken(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement GetDelegatedAccessToken
	// Required parameter: GetDelegatedAccessToken (map[string]interface{}) - Input for GetDelegatedAccessToken

	return s.errorResponse(501, "NotImplemented", "GetDelegatedAccessToken is not yet implemented"), nil
}

func (s *StsService) getFederationToken(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement GetFederationToken
	// Required parameter: GetFederationToken (map[string]interface{}) - Input for GetFederationToken

	return s.errorResponse(501, "NotImplemented", "GetFederationToken is not yet implemented"), nil
}

func (s *StsService) getSessionToken(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement GetSessionToken
	// Required parameter: GetSessionToken (map[string]interface{}) - Input for GetSessionToken

	return s.errorResponse(501, "NotImplemented", "GetSessionToken is not yet implemented"), nil
}

func (s *StsService) successResponse(action string, data interface{}) (*emulator.AWSResponse, error) {
	return emulator.BuildQueryResponse(action, data, emulator.ResponseBuilderConfig{
		ServiceName: "sts",
		Version:     "2011-06-15",
	})
}

func (s *StsService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildErrorResponse("sts", statusCode, code, message)
}

// Helper functions
func getStringParam(params map[string]interface{}, key, defaultValue string) *string {
	if val, ok := params[key].(string); ok {
		return &val
	}
	if defaultValue != "" {
		return &defaultValue
	}
	return nil
}

func getInt32Param(params map[string]interface{}, key string, defaultValue int32) *int32 {
	if val, ok := params[key].(float64); ok {
		result := int32(val)
		return &result
	}
	if val, ok := params[key].(int); ok {
		result := int32(val)
		return &result
	}
	if val, ok := params[key].(int32); ok {
		return &val
	}
	if defaultValue != 0 {
		return &defaultValue
	}
	return nil
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) *bool {
	if val, ok := params[key].(bool); ok {
		return &val
	}
	if val, ok := params[key].(string); ok {
		result := val == "true"
		return &result
	}
	return &defaultValue
}
