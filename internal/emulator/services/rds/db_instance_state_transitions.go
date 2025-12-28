package rds

import (
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// transitionDBInstanceState atomically transitions a DB instance to a new state with validation
func (s *RDSService) transitionDBInstanceState(identifier string, newState DBInstanceState) error {
	resourceKey := "db-instance:" + identifier
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	key := fmt.Sprintf("rds:db-instance:%s", identifier)
	var instance DBInstance

	// Use atomic Update to prevent race conditions between Get and Set
	return s.state.Update(key, &instance, func() error {
		if instance.DBInstanceStatus == nil {
			return fmt.Errorf("DB instance has no status: %s", identifier)
		}

		currentState := DBInstanceState(*instance.DBInstanceStatus)

		// Validate the transition
		if !IsValidDBInstanceTransition(currentState, newState) {
			return NewDBInstanceStateError(identifier, currentState, newState)
		}

		// Apply the transition
		newStatus := string(newState)
		instance.DBInstanceStatus = helpers.StringPtr(newStatus)

		return nil
	})
}

// scheduleDBInstanceTransition schedules an async state transition with cancellation support
func (s *RDSService) scheduleDBInstanceTransition(identifier string, targetState DBInstanceState, delay time.Duration) {
	resourceKey := "db-instance:" + identifier
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
			if err := s.transitionDBInstanceState(identifier, targetState); err != nil {
				// Log the error but don't fail - state might have changed
				// In production, this could be sent to a logger
			}
			s.stateMachine.ClearPendingTransition(resourceKey)
		}
	}()
}

// removeDBInstanceAfterDelay removes a DB instance after a delay (for deleted instances)
// Uses a separate resource key ("removal:db-instance:xxx") to track the removal operation
// so it doesn't conflict with state transition tracking
func (s *RDSService) removeDBInstanceAfterDelay(identifier string, delay time.Duration) {
	// Use separate key for removal tracking to avoid conflicting with state transitions
	removalKey := "removal:db-instance:" + identifier
	instanceKey := "db-instance:" + identifier

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

			key := fmt.Sprintf("rds:db-instance:%s", identifier)
			var instance DBInstance
			if err := s.state.Get(key, &instance); err != nil {
				// Instance already gone
				s.stateMachine.ClearPendingTransition(removalKey)
				s.stateMachine.RemoveResourceState(removalKey)
				return
			}

			// Only delete if still in "deleting" state
			if instance.DBInstanceStatus != nil && DBInstanceState(*instance.DBInstanceStatus) == DBInstanceStateDeleting {
				s.state.Delete(key)
				// Also delete the resource ID index
				if instance.DbiResourceId != nil {
					s.state.Delete(fmt.Sprintf("rds:db-instance-by-resource-id:%s", *instance.DbiResourceId))
				}
			}

			s.stateMachine.ClearPendingTransition(removalKey)
			s.stateMachine.RemoveResourceState(removalKey)
			s.stateMachine.RemoveResourceState(instanceKey)
		}
	}()
}
