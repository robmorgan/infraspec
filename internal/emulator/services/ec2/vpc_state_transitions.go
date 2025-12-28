package ec2

import (
	"fmt"
	"time"
)

// transitionVpcState atomically transitions a VPC to a new state with validation
func (s *EC2Service) transitionVpcState(vpcId string, newState VpcState) error {
	resourceKey := "vpcs:" + vpcId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	key := fmt.Sprintf("ec2:vpcs:%s", vpcId)
	var vpc Vpc

	// Use atomic Update to prevent race conditions between Get and Set
	return s.state.Update(key, &vpc, func() error {
		currentState := vpc.State
		if !IsValidVpcTransition(currentState, newState) {
			return fmt.Errorf("invalid VPC transition from %s to %s", currentState, newState)
		}

		vpc.State = newState
		return nil
	})
}

// scheduleVpcTransition schedules an async VPC state transition
func (s *EC2Service) scheduleVpcTransition(vpcId string, targetState VpcState, delay time.Duration) {
	resourceKey := "vpcs:" + vpcId
	cancelCh := s.stateMachine.SetPendingTransition(resourceKey, string(targetState))

	go func() {
		select {
		case <-s.shutdownCtx.Done():
			return
		case <-cancelCh:
			return
		case <-time.After(delay):
			s.transitionVpcState(vpcId, targetState)
			s.stateMachine.ClearPendingTransition(resourceKey)
		}
	}()
}
