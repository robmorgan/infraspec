package ec2

import (
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// transitionInstanceState atomically transitions an instance to a new state with validation
func (s *EC2Service) transitionInstanceState(instanceId string, newState InstanceStateName) error {
	resourceKey := "instances:" + instanceId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	key := fmt.Sprintf("ec2:instances:%s", instanceId)
	var instance Instance

	// Use atomic Update to prevent race conditions between Get and Set
	return s.state.Update(key, &instance, func() error {
		if instance.State == nil {
			return fmt.Errorf("instance has no state: %s", instanceId)
		}

		currentState := instance.State.Name

		// Validate the transition
		if !IsValidInstanceTransition(currentState, newState) {
			return NewInstanceStateError(instanceId, currentState, newState)
		}

		// Apply the transition
		instance.State = &InstanceState{
			Code: helpers.Int32Ptr(InstanceStateCode(newState)),
			Name: newState,
		}

		return nil
	})
}

// scheduleInstanceTransition schedules an async state transition with cancellation support
func (s *EC2Service) scheduleInstanceTransition(instanceId string, targetState InstanceStateName, delay time.Duration) {
	resourceKey := "instances:" + instanceId
	cancelCh := s.stateMachine.SetPendingTransition(resourceKey, string(targetState))

	go func() {
		select {
		case <-s.shutdownCtx.Done():
			// Service is shutting down
			return
		case <-cancelCh:
			// Transition was cancelled
			return
		case <-time.After(delay):
			// Apply the transition
			s.transitionInstanceState(instanceId, targetState)
			s.stateMachine.ClearPendingTransition(resourceKey)
		}
	}()
}

// removeInstanceAfterDelay removes an instance after a delay (for terminated instances)
// Uses a separate resource key ("removal:instances:xxx") to track the removal operation
// so it doesn't conflict with state transition tracking
func (s *EC2Service) removeInstanceAfterDelay(instanceId string, delay time.Duration) {
	// Use separate key for removal tracking to avoid conflicting with state transitions
	removalKey := "removal:instances:" + instanceId
	instanceKey := "instances:" + instanceId

	cancelCh := s.stateMachine.SetPendingTransition(removalKey, "removed")

	go func() {
		select {
		case <-s.shutdownCtx.Done():
			// Service is shutting down
			return
		case <-cancelCh:
			// Removal was cancelled
			return
		case <-time.After(delay):
			// Acquire lock on the instance resource for the actual deletion
			rs := s.stateMachine.GetOrCreateResourceState(instanceKey)
			rs.mu.Lock()
			defer rs.mu.Unlock()

			s.state.Delete(fmt.Sprintf("ec2:instances:%s", instanceId))
			s.stateMachine.ClearPendingTransition(removalKey)
			s.stateMachine.RemoveResourceState(removalKey)
			s.stateMachine.RemoveResourceState(instanceKey)
		}
	}()
}
