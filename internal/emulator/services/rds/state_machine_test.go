package rds

import (
	"sync"
	"testing"
	"time"
)

func TestIsValidDBInstanceTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     DBInstanceState
		to       DBInstanceState
		expected bool
	}{
		// Valid transitions from creating
		{"creating to available", DBInstanceStateCreating, DBInstanceStateAvailable, true},
		{"creating to failed", DBInstanceStateCreating, DBInstanceStateFailed, true},

		// Valid transitions from available
		{"available to modifying", DBInstanceStateAvailable, DBInstanceStateModifying, true},
		{"available to stopping", DBInstanceStateAvailable, DBInstanceStateStopping, true},
		{"available to rebooting", DBInstanceStateAvailable, DBInstanceStateRebooting, true},
		{"available to deleting", DBInstanceStateAvailable, DBInstanceStateDeleting, true},
		{"available to backing-up", DBInstanceStateAvailable, DBInstanceStateBackingUp, true},

		// Valid transitions from modifying
		{"modifying to available", DBInstanceStateModifying, DBInstanceStateAvailable, true},
		{"modifying to failed", DBInstanceStateModifying, DBInstanceStateFailed, true},

		// Valid transitions from stopping
		{"stopping to stopped", DBInstanceStateStopping, DBInstanceStateStopped, true},
		{"stopping to failed", DBInstanceStateStopping, DBInstanceStateFailed, true},

		// Valid transitions from stopped
		{"stopped to starting", DBInstanceStateStopped, DBInstanceStateStarting, true},
		{"stopped to deleting", DBInstanceStateStopped, DBInstanceStateDeleting, true},

		// Valid transitions from starting
		{"starting to available", DBInstanceStateStarting, DBInstanceStateAvailable, true},
		{"starting to failed", DBInstanceStateStarting, DBInstanceStateFailed, true},

		// Valid transitions from rebooting
		{"rebooting to available", DBInstanceStateRebooting, DBInstanceStateAvailable, true},
		{"rebooting to failed", DBInstanceStateRebooting, DBInstanceStateFailed, true},

		// Valid transitions from backing-up
		{"backing-up to available", DBInstanceStateBackingUp, DBInstanceStateAvailable, true},

		// Valid transitions from failed
		{"failed to deleting", DBInstanceStateFailed, DBInstanceStateDeleting, true},

		// Valid transitions from storage-full
		{"storage-full to modifying", DBInstanceStateStorageFull, DBInstanceStateModifying, true},
		{"storage-full to deleting", DBInstanceStateStorageFull, DBInstanceStateDeleting, true},

		// Invalid transitions
		{"creating to modifying", DBInstanceStateCreating, DBInstanceStateModifying, false},
		{"creating to stopping", DBInstanceStateCreating, DBInstanceStateStopping, false},
		{"available to creating", DBInstanceStateAvailable, DBInstanceStateCreating, false},
		{"available to starting", DBInstanceStateAvailable, DBInstanceStateStarting, false},
		{"stopped to available", DBInstanceStateStopped, DBInstanceStateAvailable, false},
		{"stopped to modifying", DBInstanceStateStopped, DBInstanceStateModifying, false},
		{"deleting to available", DBInstanceStateDeleting, DBInstanceStateAvailable, false},
		{"deleting to stopped", DBInstanceStateDeleting, DBInstanceStateStopped, false},
		{"rebooting to stopping", DBInstanceStateRebooting, DBInstanceStateStopping, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDBInstanceTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidDBInstanceTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestDBInstanceState_String(t *testing.T) {
	tests := []struct {
		state    DBInstanceState
		expected string
	}{
		{DBInstanceStateCreating, "creating"},
		{DBInstanceStateAvailable, "available"},
		{DBInstanceStateModifying, "modifying"},
		{DBInstanceStateStopped, "stopped"},
		{DBInstanceStateDeleting, "deleting"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("DBInstanceState.String() = %s, want %s", tt.state.String(), tt.expected)
			}
		})
	}
}

func TestDBInstanceState_IsTransitionalState(t *testing.T) {
	transitionalStates := []DBInstanceState{
		DBInstanceStateCreating,
		DBInstanceStateModifying,
		DBInstanceStateStarting,
		DBInstanceStateStopping,
		DBInstanceStateRebooting,
		DBInstanceStateDeleting,
		DBInstanceStateBackingUp,
	}

	stableStates := []DBInstanceState{
		DBInstanceStateAvailable,
		DBInstanceStateStopped,
		DBInstanceStateFailed,
		DBInstanceStateStorageFull,
	}

	for _, state := range transitionalStates {
		if !state.IsTransitionalState() {
			t.Errorf("%s should be a transitional state", state)
		}
	}

	for _, state := range stableStates {
		if state.IsTransitionalState() {
			t.Errorf("%s should not be a transitional state", state)
		}
	}
}

func TestDBInstanceState_IsStableState(t *testing.T) {
	stableStates := []DBInstanceState{
		DBInstanceStateAvailable,
		DBInstanceStateStopped,
		DBInstanceStateFailed,
		DBInstanceStateStorageFull,
	}

	transitionalStates := []DBInstanceState{
		DBInstanceStateCreating,
		DBInstanceStateModifying,
		DBInstanceStateStarting,
		DBInstanceStateStopping,
		DBInstanceStateRebooting,
		DBInstanceStateDeleting,
		DBInstanceStateBackingUp,
	}

	for _, state := range stableStates {
		if !state.IsStableState() {
			t.Errorf("%s should be a stable state", state)
		}
	}

	for _, state := range transitionalStates {
		if state.IsStableState() {
			t.Errorf("%s should not be a stable state", state)
		}
	}
}

func TestResourceStateManager_GetOrCreateResourceState(t *testing.T) {
	mgr := NewResourceStateManager()

	// First call should create new state
	rs1 := mgr.GetOrCreateResourceState("db-instance:mydb")
	if rs1 == nil {
		t.Fatal("GetOrCreateResourceState returned nil")
	}

	// Second call should return same state
	rs2 := mgr.GetOrCreateResourceState("db-instance:mydb")
	if rs1 != rs2 {
		t.Error("GetOrCreateResourceState should return same state for same key")
	}

	// Different resource should get different state
	rs3 := mgr.GetOrCreateResourceState("db-instance:otherdb")
	if rs1 == rs3 {
		t.Error("GetOrCreateResourceState should return different state for different key")
	}
}

func TestResourceStateManager_SetPendingTransition(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Set a pending transition
	cancelCh := mgr.SetPendingTransition(resourceKey, "available")
	if cancelCh == nil {
		t.Fatal("SetPendingTransition returned nil channel")
	}

	// Verify the resource state has the pending transition
	rs := mgr.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	if rs.pendingTransition == nil {
		t.Error("Resource should have pending transition")
	}
	if rs.pendingTransition.targetState != "available" {
		t.Errorf("Expected target state 'available', got '%s'", rs.pendingTransition.targetState)
	}
	rs.mu.Unlock()
}

func TestResourceStateManager_SetPendingTransition_CancelsPrevious(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Set first transition
	cancelCh1 := mgr.SetPendingTransition(resourceKey, "modifying")

	// Set second transition - should cancel first
	cancelCh2 := mgr.SetPendingTransition(resourceKey, "available")

	// First channel should be closed
	select {
	case <-cancelCh1:
		// Expected - channel was closed
	default:
		t.Error("First cancel channel should have been closed")
	}

	// Second channel should still be open
	select {
	case <-cancelCh2:
		t.Error("Second cancel channel should not be closed")
	default:
		// Expected - channel is still open
	}
}

func TestResourceStateManager_CancelPendingTransition(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Set a pending transition
	cancelCh := mgr.SetPendingTransition(resourceKey, "available")

	// Cancel it
	mgr.CancelPendingTransition(resourceKey)

	// Channel should be closed
	select {
	case <-cancelCh:
		// Expected
	default:
		t.Error("Cancel channel should have been closed")
	}

	// Resource should have no pending transition
	rs := mgr.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	if rs.pendingTransition != nil {
		t.Error("Resource should not have pending transition after cancel")
	}
	rs.mu.Unlock()
}

func TestResourceStateManager_ClearPendingTransition(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Set a pending transition
	mgr.SetPendingTransition(resourceKey, "available")

	// Clear it
	mgr.ClearPendingTransition(resourceKey)

	// Resource should have no pending transition
	rs := mgr.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	if rs.pendingTransition != nil {
		t.Error("Resource should not have pending transition after clear")
	}
	rs.mu.Unlock()
}

func TestResourceStateManager_RemoveResourceState(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Create resource state and set pending transition
	mgr.SetPendingTransition(resourceKey, "available")

	// Remove it
	mgr.RemoveResourceState(resourceKey)

	// Getting it again should create a new one (without pending transition)
	rs := mgr.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	if rs.pendingTransition != nil {
		t.Error("New resource state should not have pending transition")
	}
	rs.mu.Unlock()
}

func TestResourceStateManager_ConcurrentAccess(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "db-instance:mydb"

	// Run many goroutines accessing the same resource concurrently
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine does a mix of operations
			rs := mgr.GetOrCreateResourceState(resourceKey)

			rs.mu.Lock()
			// Simulate some work
			time.Sleep(time.Microsecond)
			rs.mu.Unlock()

			if id%3 == 0 {
				mgr.SetPendingTransition(resourceKey, "available")
			} else if id%3 == 1 {
				mgr.CancelPendingTransition(resourceKey)
			} else {
				mgr.ClearPendingTransition(resourceKey)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no deadlock or panic occurred
}

func TestNewDBInstanceStateError(t *testing.T) {
	err := NewDBInstanceStateError("mydb", DBInstanceStateAvailable, DBInstanceStateCreating)

	if err.ResourceType != "db-instance" {
		t.Errorf("Expected ResourceType 'db-instance', got '%s'", err.ResourceType)
	}
	if err.ResourceID != "mydb" {
		t.Errorf("Expected ResourceID 'mydb', got '%s'", err.ResourceID)
	}
	if err.FromState != "available" {
		t.Errorf("Expected FromState 'available', got '%s'", err.FromState)
	}
	if err.ToState != "creating" {
		t.Errorf("Expected ToState 'creating', got '%s'", err.ToState)
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestStateTransitionError_Error(t *testing.T) {
	err := &StateTransitionError{
		ResourceType: "db-instance",
		ResourceID:   "mydb",
		FromState:    "available",
		ToState:      "creating",
	}

	expected := "invalid db-instance state transition from available to creating for mydb"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
