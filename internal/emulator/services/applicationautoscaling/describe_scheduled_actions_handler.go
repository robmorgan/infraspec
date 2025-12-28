package applicationautoscaling

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *ApplicationAutoScalingService) describeScheduledActions(ctx context.Context, input *DescribeScheduledActionsRequest) (*emulator.AWSResponse, error) {
	// Extract required parameter
	if input.ServiceNamespace == "" {
		return s.errorResponse(400, "ValidationException", "ServiceNamespace is required"), nil
	}

	// Build prefix for listing scheduled actions
	prefix := fmt.Sprintf("autoscaling:scheduled-action:%s:", input.ServiceNamespace)

	// List all scheduled actions for this service namespace
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServiceException", "Failed to list scheduled actions"), nil
	}

	var scheduledActions []ScheduledAction
	for _, key := range keys {
		var action ScheduledAction
		if err := s.state.Get(key, &action); err == nil {
			// Apply filters
			if len(input.ScheduledActionNames) > 0 {
				found := false
				for _, name := range input.ScheduledActionNames {
					if action.ScheduledActionName != nil && name == *action.ScheduledActionName {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if input.ResourceId != nil && *input.ResourceId != "" {
				if action.ResourceId == nil || *action.ResourceId != *input.ResourceId {
					continue
				}
			}

			if input.ScalableDimension != "" {
				if action.ScalableDimension != input.ScalableDimension {
					continue
				}
			}

			scheduledActions = append(scheduledActions, action)
		}
	}

	// Return response
	response := &DescribeScheduledActionsResponse{
		ScheduledActions: scheduledActions,
	}

	return s.jsonResponse(200, response)
}
