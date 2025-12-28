package rds

import (
	"fmt"
	"sync"
	"time"
)

// ResourceState represents state tracking for a single resource
type ResourceState struct {
	mu                sync.Mutex
	pendingTransition *PendingTransition
}

// PendingTransition tracks an in-flight async state change
type PendingTransition struct {
	targetState string
	cancelCh    chan struct{}
	scheduledAt time.Time
}

// ResourceStateManager manages per-resource state machines
type ResourceStateManager struct {
	mu        sync.RWMutex
	resources map[string]*ResourceState // key: "db-instance:mydb", "db-cluster:mycluster"
}

// NewResourceStateManager creates a new resource state manager
func NewResourceStateManager() *ResourceStateManager {
	return &ResourceStateManager{
		resources: make(map[string]*ResourceState),
	}
}

// GetOrCreateResourceState gets or creates state tracking for a resource
func (m *ResourceStateManager) GetOrCreateResourceState(resourceKey string) *ResourceState {
	m.mu.RLock()
	rs, exists := m.resources[resourceKey]
	m.mu.RUnlock()

	if exists {
		return rs
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if rs, exists = m.resources[resourceKey]; exists {
		return rs
	}

	rs = &ResourceState{}
	m.resources[resourceKey] = rs
	return rs
}

// RemoveResourceState removes state tracking for a resource (cleanup after deletion)
func (m *ResourceStateManager) RemoveResourceState(resourceKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.resources, resourceKey)
}

// CancelPendingTransition cancels any pending transition for a resource
func (m *ResourceStateManager) CancelPendingTransition(resourceKey string) {
	rs := m.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.pendingTransition != nil {
		close(rs.pendingTransition.cancelCh)
		rs.pendingTransition = nil
	}
}

// SetPendingTransition sets a pending transition for a resource
// Returns a cancel channel that will be closed if the transition is cancelled
func (m *ResourceStateManager) SetPendingTransition(resourceKey string, targetState string) chan struct{} {
	rs := m.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Cancel any existing pending transition
	if rs.pendingTransition != nil {
		close(rs.pendingTransition.cancelCh)
	}

	cancelCh := make(chan struct{})
	rs.pendingTransition = &PendingTransition{
		targetState: targetState,
		cancelCh:    cancelCh,
		scheduledAt: time.Now(),
	}

	return cancelCh
}

// ClearPendingTransition clears the pending transition after it completes
func (m *ResourceStateManager) ClearPendingTransition(resourceKey string) {
	rs := m.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.pendingTransition = nil
}

// StateTransitionError represents an invalid state transition
type StateTransitionError struct {
	ResourceType string
	ResourceID   string
	FromState    string
	ToState      string
}

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("invalid %s state transition from %s to %s for %s",
		e.ResourceType, e.FromState, e.ToState, e.ResourceID)
}
