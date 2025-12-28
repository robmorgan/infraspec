package ec2

import (
	"fmt"
	"time"
)

// transitionSubnetState atomically transitions a subnet to a new state with validation
func (s *EC2Service) transitionSubnetState(subnetId string, newState SubnetState) error {
	resourceKey := "subnets:" + subnetId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	key := fmt.Sprintf("ec2:subnets:%s", subnetId)
	var subnet Subnet

	// Use atomic Update to prevent race conditions between Get and Set
	return s.state.Update(key, &subnet, func() error {
		currentState := subnet.State
		if !IsValidSubnetTransition(currentState, newState) {
			return fmt.Errorf("invalid subnet transition from %s to %s", currentState, newState)
		}

		subnet.State = newState
		return nil
	})
}

// scheduleSubnetTransition schedules an async subnet state transition
func (s *EC2Service) scheduleSubnetTransition(subnetId string, targetState SubnetState, delay time.Duration) {
	resourceKey := "subnets:" + subnetId
	cancelCh := s.stateMachine.SetPendingTransition(resourceKey, string(targetState))

	go func() {
		select {
		case <-s.shutdownCtx.Done():
			return
		case <-cancelCh:
			return
		case <-time.After(delay):
			s.transitionSubnetState(subnetId, targetState)
			s.stateMachine.ClearPendingTransition(resourceKey)
		}
	}()
}
