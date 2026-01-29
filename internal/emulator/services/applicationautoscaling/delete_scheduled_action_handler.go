package applicationautoscaling

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *ApplicationAutoScalingService) deleteScheduledAction(ctx context.Context, input *DeleteScheduledActionRequest) (*emulator.AWSResponse, error) {
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

	// Check if scheduled action exists
	if !s.state.Exists(key) {
		return s.errorResponse(404, "ObjectNotFoundException",
			fmt.Sprintf("Scheduled action %s not found", *input.ScheduledActionName)), nil
	}

	// Delete the scheduled action
	if err := s.state.Delete(key); err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to delete scheduled action"), nil
	}

	// Return empty response on success
	return s.jsonResponse(200, &DeleteScheduledActionResponse{})
}
