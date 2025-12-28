package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) startInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	instanceIds := s.parseInstanceIds(params)
	if len(instanceIds) == 0 {
		return s.errorResponse(400, "MissingParameter", "InstanceId is required"), nil
	}

	stateChanges := make([]InstanceStateChange, 0)

	for _, instanceId := range instanceIds {
		resourceKey := "instances:" + instanceId

		// Cancel any pending transitions first (before acquiring lock)
		s.stateMachine.CancelPendingTransition(resourceKey)

		rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
		rs.mu.Lock()

		var instance Instance
		if err := s.state.Get(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
			rs.mu.Unlock()
			return s.errorResponse(400, "InvalidInstanceID.NotFound", fmt.Sprintf("The instance ID '%s' does not exist", instanceId)), nil
		}

		// Validate instance is in stopped state
		if instance.State == nil || instance.State.Name != InstanceStateName("stopped") {
			currentState := "unknown"
			if instance.State != nil {
				currentState = string(instance.State.Name)
			}
			rs.mu.Unlock()
			return s.errorResponse(400, "IncorrectInstanceState", fmt.Sprintf("The instance '%s' is not in the 'stopped' state. Current state: %s", instanceId, currentState)), nil
		}

		previousState := instance.State
		instance.State = &InstanceState{
			Code: helpers.Int32Ptr(0),
			Name: InstanceStateName("pending"),
		}

		s.state.Set(fmt.Sprintf("ec2:instances:%s", instanceId), &instance)
		rs.mu.Unlock()

		stateChanges = append(stateChanges, InstanceStateChange{
			InstanceId:    &instanceId,
			CurrentState:  instance.State,
			PreviousState: previousState,
		})

		// Schedule transition to running
		s.scheduleInstanceTransition(instanceId, InstanceStateName("running"), 5*time.Second)
	}

	return s.instanceStateChangeResponse("StartInstances", stateChanges)
}
