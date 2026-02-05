package applicationautoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *ApplicationAutoScalingService) putScheduledAction(ctx context.Context, input *PutScheduledActionRequest) (*emulator.AWSResponse, error) {
	// Extract and validate required parameters
	if input.ScheduledActionName == nil || *input.ScheduledActionName == "" {
		return s.errorResponse(400, "ValidationException", "ScheduledActionName is required"), nil
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

	// Build state key for the scheduled action
	key := fmt.Sprintf("autoscaling:scheduled-action:%s:%s:%s:%s",
		input.ServiceNamespace, *input.ResourceId, input.ScalableDimension, *input.ScheduledActionName)

	// Create the scheduled action
	now := UnixTimestamp(time.Now())
	scheduledActionARN := fmt.Sprintf("arn:aws:autoscaling:us-east-1:000000000000:scheduledAction:%s:resource/%s/%s:scheduledActionName/%s",
		uuid.New().String(), input.ServiceNamespace, *input.ResourceId, *input.ScheduledActionName)

	// Convert input timestamps to UnixTimestamp
	var startTime, endTime *UnixTimestamp
	if input.StartTime != nil {
		st := UnixTimestamp(*input.StartTime)
		startTime = &st
	}
	if input.EndTime != nil {
		et := UnixTimestamp(*input.EndTime)
		endTime = &et
	}

	scheduledAction := &ScheduledAction{
		ScheduledActionName:  input.ScheduledActionName,
		ServiceNamespace:     input.ServiceNamespace,
		ResourceId:           input.ResourceId,
		ScalableDimension:    input.ScalableDimension,
		CreationTime:         &now,
		ScheduledActionARN:   &scheduledActionARN,
		Schedule:             input.Schedule,
		Timezone:             input.Timezone,
		StartTime:            startTime,
		EndTime:              endTime,
		ScalableTargetAction: input.ScalableTargetAction,
	}

	// Save to state
	if err := s.state.Set(key, scheduledAction); err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to create scheduled action"), nil
	}

	// Return empty response on success
	return s.jsonResponse(200, &PutScheduledActionResponse{})
}
