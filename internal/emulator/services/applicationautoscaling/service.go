package applicationautoscaling

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

type ApplicationAutoScalingService struct {
	state     emulator.StateManager
	validator emulator.Validator
}

func NewApplicationAutoScalingService(state emulator.StateManager, validator emulator.Validator) *ApplicationAutoScalingService {
	return &ApplicationAutoScalingService{
		state:     state,
		validator: validator,
	}
}

func (s *ApplicationAutoScalingService) ServiceName() string {
	return "anyscalefrontendservice"
}

func (s *ApplicationAutoScalingService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	switch action {
	case "RegisterScalableTarget":
		input, err := emulator.ParseJSONRequest[RegisterScalableTargetRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.registerScalableTarget(ctx, input)
	case "DeregisterScalableTarget":
		input, err := emulator.ParseJSONRequest[DeregisterScalableTargetRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deregisterScalableTarget(ctx, input)
	case "DescribeScalableTargets":
		input, err := emulator.ParseJSONRequest[DescribeScalableTargetsRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeScalableTargets(ctx, input)
	case "PutScalingPolicy":
		input, err := emulator.ParseJSONRequest[PutScalingPolicyRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.putScalingPolicy(ctx, input)
	case "DeleteScalingPolicy":
		input, err := emulator.ParseJSONRequest[DeleteScalingPolicyRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteScalingPolicy(ctx, input)
	case "DescribeScalingPolicies":
		input, err := emulator.ParseJSONRequest[DescribeScalingPoliciesRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeScalingPolicies(ctx, input)
	case "DescribeScalingActivities":
		input, err := emulator.ParseJSONRequest[DescribeScalingActivitiesRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeScalingActivities(ctx, input)
	case "ListTagsForResource":
		input, err := emulator.ParseJSONRequest[ListTagsForResourceRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listTagsForResource(ctx, input)
	case "TagResource":
		input, err := emulator.ParseJSONRequest[TagResourceRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.tagResource(ctx, input)
	case "UntagResource":
		input, err := emulator.ParseJSONRequest[UntagResourceRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.untagResource(ctx, input)
	case "DeleteScheduledAction":
		input, err := emulator.ParseJSONRequest[DeleteScheduledActionRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteScheduledAction(ctx, input)
	case "DescribeScheduledActions":
		input, err := emulator.ParseJSONRequest[DescribeScheduledActionsRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeScheduledActions(ctx, input)
	case "GetPredictiveScalingForecast":
		input, err := emulator.ParseJSONRequest[GetPredictiveScalingForecastRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.getPredictiveScalingForecast(ctx, input)
	case "PutScheduledAction":
		input, err := emulator.ParseJSONRequest[PutScheduledActionRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.putScheduledAction(ctx, input)
	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *ApplicationAutoScalingService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	// Application Auto Scaling uses X-Amz-Target header: "AnyScaleFrontendService.RegisterScalableTarget"
	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *ApplicationAutoScalingService) registerScalableTarget(ctx context.Context, input *RegisterScalableTargetRequest) (*emulator.AWSResponse, error) {
	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	if input.ResourceId == nil || *input.ResourceId == "" {
		return s.errorResponse(400, "ValidationException", "ResourceId is required"), nil
	}

	if input.ScalableDimension == "" {
		return s.errorResponse(400, "ValidationException", "ScalableDimension is required"), nil
	}

	serviceNamespace := string(input.ServiceNamespace)
	resourceId := *input.ResourceId
	scalableDimension := string(input.ScalableDimension)

	// Build target key
	key := fmt.Sprintf("autoscaling:target:%s:%s:%s", serviceNamespace, resourceId, scalableDimension)

	// Generate ARN
	targetARN := fmt.Sprintf("arn:aws:application-autoscaling:us-east-1:000000000000:scalable-target/%s", uuid.New().String())

	// Create suspended state with defaults
	suspendedState := &SuspendedState{
		DynamicScalingInSuspended:  boolPtr(false),
		DynamicScalingOutSuspended: boolPtr(false),
		ScheduledScalingSuspended:  boolPtr(false),
	}
	if input.SuspendedState != nil {
		if input.SuspendedState.DynamicScalingInSuspended != nil {
			suspendedState.DynamicScalingInSuspended = input.SuspendedState.DynamicScalingInSuspended
		}
		if input.SuspendedState.DynamicScalingOutSuspended != nil {
			suspendedState.DynamicScalingOutSuspended = input.SuspendedState.DynamicScalingOutSuspended
		}
		if input.SuspendedState.ScheduledScalingSuspended != nil {
			suspendedState.ScheduledScalingSuspended = input.SuspendedState.ScheduledScalingSuspended
		}
	}

	// Set RoleARN default
	roleARN := "arn:aws:iam::000000000000:role/aws-service-role/dynamodb.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_DynamoDBTable"
	if input.RoleARN != nil && *input.RoleARN != "" {
		roleARN = *input.RoleARN
	}

	// Create scalable target
	now := UnixTimestamp(time.Now())
	target := &ScalableTarget{
		ServiceNamespace:  input.ServiceNamespace,
		ResourceId:        input.ResourceId,
		ScalableDimension: input.ScalableDimension,
		MinCapacity:       input.MinCapacity,
		MaxCapacity:       input.MaxCapacity,
		RoleARN:           &roleARN,
		CreationTime:      &now,
		SuspendedState:    suspendedState,
		ScalableTargetARN: &targetARN,
	}

	// Save to state
	if err := s.state.Set(key, target); err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to register scalable target"), nil
	}

	// Return response
	response := &RegisterScalableTargetResponse{
		ScalableTargetARN: &targetARN,
	}

	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) deregisterScalableTarget(ctx context.Context, input *DeregisterScalableTargetRequest) (*emulator.AWSResponse, error) {
	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	if input.ResourceId == nil || *input.ResourceId == "" {
		return s.errorResponse(400, "ValidationException", "ResourceId is required"), nil
	}

	if input.ScalableDimension == "" {
		return s.errorResponse(400, "ValidationException", "ScalableDimension is required"), nil
	}

	key := fmt.Sprintf("autoscaling:target:%s:%s:%s", input.ServiceNamespace, *input.ResourceId, input.ScalableDimension)

	// Delete from state
	if err := s.state.Delete(key); err != nil {
		return s.errorResponse(404, "ObjectNotFoundException", "Scalable target not found"), nil
	}

	return s.jsonResponse(200, &DeregisterScalableTargetResponse{})
}

func (s *ApplicationAutoScalingService) describeScalableTargets(ctx context.Context, input *DescribeScalableTargetsRequest) (*emulator.AWSResponse, error) {
	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	// List all targets for this service namespace
	prefix := fmt.Sprintf("autoscaling:target:%s:", input.ServiceNamespace)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to list scalable targets"), nil
	}

	var targets []ScalableTarget
	for _, key := range keys {
		var target ScalableTarget
		if err := s.state.Get(key, &target); err == nil {
			// Apply filters
			if len(input.ResourceIds) > 0 {
				found := false
				for _, rid := range input.ResourceIds {
					if target.ResourceId != nil && rid == *target.ResourceId {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if input.ScalableDimension != "" {
				if target.ScalableDimension != input.ScalableDimension {
					continue
				}
			}

			targets = append(targets, target)
		}
	}

	response := &DescribeScalableTargetsResponse{
		ScalableTargets: targets,
	}

	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) putScalingPolicy(ctx context.Context, input *PutScalingPolicyRequest) (*emulator.AWSResponse, error) {
	if input.PolicyName == nil || *input.PolicyName == "" {
		return s.errorResponse(400, "ValidationException", "PolicyName is required"), nil
	}

	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	if input.ResourceId == nil || *input.ResourceId == "" {
		return s.errorResponse(400, "ValidationException", "ResourceId is required"), nil
	}

	if input.ScalableDimension == "" {
		return s.errorResponse(400, "ValidationException", "ScalableDimension is required"), nil
	}

	policyType := PolicyType("TargetTrackingScaling")
	if input.PolicyType != "" {
		policyType = input.PolicyType
	}

	// Build policy key
	key := fmt.Sprintf("autoscaling:policy:%s:%s:%s:%s", input.ServiceNamespace, *input.ResourceId, input.ScalableDimension, *input.PolicyName)

	// Create scaling policy
	now := UnixTimestamp(time.Now())
	policyARN := fmt.Sprintf("arn:aws:autoscaling:us-east-1:000000000000:scalingPolicy:%s:resource/%s/%s:policyName/%s", uuid.New().String(), input.ServiceNamespace, *input.ResourceId, *input.PolicyName)
	policy := &ScalingPolicy{
		PolicyName:                               input.PolicyName,
		ServiceNamespace:                         input.ServiceNamespace,
		ResourceId:                               input.ResourceId,
		ScalableDimension:                        input.ScalableDimension,
		PolicyType:                               policyType,
		PolicyARN:                                &policyARN,
		CreationTime:                             &now,
		Alarms:                                   []Alarm{},
		TargetTrackingScalingPolicyConfiguration: input.TargetTrackingScalingPolicyConfiguration,
		StepScalingPolicyConfiguration:           input.StepScalingPolicyConfiguration,
		PredictiveScalingPolicyConfiguration:     input.PredictiveScalingPolicyConfiguration,
	}

	// Save to state
	if err := s.state.Set(key, policy); err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to create scaling policy"), nil
	}

	// Return response
	response := &PutScalingPolicyResponse{
		PolicyARN: &policyARN,
		Alarms:    []Alarm{},
	}

	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) deleteScalingPolicy(ctx context.Context, input *DeleteScalingPolicyRequest) (*emulator.AWSResponse, error) {
	if input.PolicyName == nil || *input.PolicyName == "" {
		return s.errorResponse(400, "ValidationException", "PolicyName is required"), nil
	}

	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	if input.ResourceId == nil || *input.ResourceId == "" {
		return s.errorResponse(400, "ValidationException", "ResourceId is required"), nil
	}

	if input.ScalableDimension == "" {
		return s.errorResponse(400, "ValidationException", "ScalableDimension is required"), nil
	}

	key := fmt.Sprintf("autoscaling:policy:%s:%s:%s:%s", input.ServiceNamespace, *input.ResourceId, input.ScalableDimension, *input.PolicyName)

	// Delete from state
	if err := s.state.Delete(key); err != nil {
		return s.errorResponse(404, "ObjectNotFoundException", "Scaling policy not found"), nil
	}

	return s.jsonResponse(200, &DeleteScalingPolicyResponse{})
}

func (s *ApplicationAutoScalingService) describeScalingPolicies(ctx context.Context, input *DescribeScalingPoliciesRequest) (*emulator.AWSResponse, error) {
	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	// List all policies for this service namespace
	prefix := fmt.Sprintf("autoscaling:policy:%s:", input.ServiceNamespace)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to list scaling policies"), nil
	}

	var policies []ScalingPolicy
	for _, key := range keys {
		var policy ScalingPolicy
		if err := s.state.Get(key, &policy); err == nil {
			// Apply filters
			if len(input.PolicyNames) > 0 {
				found := false
				for _, name := range input.PolicyNames {
					if policy.PolicyName != nil && name == *policy.PolicyName {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if input.ResourceId != nil && *input.ResourceId != "" {
				if policy.ResourceId == nil || *policy.ResourceId != *input.ResourceId {
					continue
				}
			}

			if input.ScalableDimension != "" {
				if policy.ScalableDimension != input.ScalableDimension {
					continue
				}
			}

			policies = append(policies, policy)
		}
	}

	response := &DescribeScalingPoliciesResponse{
		ScalingPolicies: policies,
	}

	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) describeScalingActivities(ctx context.Context, input *DescribeScalingActivitiesRequest) (*emulator.AWSResponse, error) {
	// Return empty activities for now
	response := &DescribeScalingActivitiesResponse{
		ScalingActivities: []ScalingActivity{},
	}

	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) jsonResponse(statusCode int, data interface{}) (*emulator.AWSResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to marshal response"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/x-amz-json-1.1",
			"x-amzn-RequestId": uuid.New().String(),
		},
		Body: body,
	}, nil
}

func (s *ApplicationAutoScalingService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	errorData := map[string]interface{}{
		"__type":  code,
		"message": message,
	}

	body, _ := json.Marshal(errorData)

	return &emulator.AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/x-amz-json-1.1",
			"x-amzn-RequestId": uuid.New().String(),
			"x-amzn-ErrorType": code,
		},
		Body: body,
	}
}

func (s *ApplicationAutoScalingService) listTagsForResource(ctx context.Context, input *ListTagsForResourceRequest) (*emulator.AWSResponse, error) {
	// ResourceARN is optional - if not provided, return empty tags
	response := &ListTagsForResourceResponse{
		Tags: map[string]string{},
	}
	return s.jsonResponse(200, response)
}

func (s *ApplicationAutoScalingService) tagResource(ctx context.Context, input *TagResourceRequest) (*emulator.AWSResponse, error) {
	// Accept tags but don't store them (we don't persist tags in the emulator)
	return s.jsonResponse(200, &TagResourceResponse{})
}

func (s *ApplicationAutoScalingService) untagResource(ctx context.Context, input *UntagResourceRequest) (*emulator.AWSResponse, error) {
	// Accept untag requests
	return s.jsonResponse(200, &UntagResourceResponse{})
}

func boolPtr(b bool) *bool {
	return &b
}
