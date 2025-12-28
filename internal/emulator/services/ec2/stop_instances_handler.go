package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) stopInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
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

		// Validate instance is in running state
		if instance.State == nil || instance.State.Name != InstanceStateName("running") {
			currentState := "unknown"
			if instance.State != nil {
				currentState = string(instance.State.Name)
			}
			rs.mu.Unlock()
			return s.errorResponse(400, "IncorrectInstanceState", fmt.Sprintf("The instance '%s' is not in the 'running' state. Current state: %s", instanceId, currentState)), nil
		}

		previousState := instance.State
		instance.State = &InstanceState{
			Code: helpers.Int32Ptr(64),
			Name: InstanceStateName("stopping"),
		}

		s.state.Set(fmt.Sprintf("ec2:instances:%s", instanceId), &instance)
		rs.mu.Unlock()

		stateChanges = append(stateChanges, InstanceStateChange{
			InstanceId:    &instanceId,
			CurrentState:  instance.State,
			PreviousState: previousState,
		})

		// Schedule transition to stopped
		s.scheduleInstanceTransition(instanceId, InstanceStateName("stopped"), 5*time.Second)
	}

	return s.instanceStateChangeResponse("StopInstances", stateChanges)
}
