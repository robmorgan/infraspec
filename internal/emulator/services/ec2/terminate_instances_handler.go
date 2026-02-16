package ec2

import (
	"context"
	"fmt"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) terminateInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
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

		// Check if already terminated
		if instance.State != nil && instance.State.Name == InstanceStateName("terminated") {
			rs.mu.Unlock()
			return s.errorResponse(400, "IncorrectInstanceState", fmt.Sprintf("The instance '%s' is already terminated", instanceId)), nil
		}

		previousState := instance.State
		instance.State = &InstanceState{
			Code: helpers.Int32Ptr(32),
			Name: InstanceStateName("shutting-down"),
		}

		s.state.Set(fmt.Sprintf("ec2:instances:%s", instanceId), &instance)
		rs.mu.Unlock()

		stateChanges = append(stateChanges, InstanceStateChange{
			InstanceId:    &instanceId,
			CurrentState:  instance.State,
			PreviousState: previousState,
		})

		// Schedule transition to terminated and then remove after delay
		s.scheduleInstanceTransition(instanceId, InstanceStateName("terminated"), 5*time.Second)
		s.removeInstanceAfterDelay(instanceId, 30*time.Second)
	}

	return s.instanceStateChangeResponse("TerminateInstances", stateChanges)
}
