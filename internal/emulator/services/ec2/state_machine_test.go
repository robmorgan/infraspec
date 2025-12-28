package ec2

import (
	"sync"
	"testing"
	"time"
)

func TestIsValidInstanceTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     InstanceStateName
		to       InstanceStateName
		expected bool
	}{
		// Valid transitions
		{"pending to running", InstanceStateName("pending"), InstanceStateName("running"), true},
		{"pending to shutting-down", InstanceStateName("pending"), InstanceStateName("shutting-down"), true},
		{"running to stopping", InstanceStateName("running"), InstanceStateName("stopping"), true},
		{"running to shutting-down", InstanceStateName("running"), InstanceStateName("shutting-down"), true},
		{"stopping to stopped", InstanceStateName("stopping"), InstanceStateName("stopped"), true},
		{"stopped to pending", InstanceStateName("stopped"), InstanceStateName("pending"), true},
		{"stopped to shutting-down", InstanceStateName("stopped"), InstanceStateName("shutting-down"), true},
		{"shutting-down to terminated", InstanceStateName("shutting-down"), InstanceStateName("terminated"), true},

		// Invalid transitions
		{"pending to stopped", InstanceStateName("pending"), InstanceStateName("stopped"), false},
		{"running to pending", InstanceStateName("running"), InstanceStateName("pending"), false},
		{"stopped to running", InstanceStateName("stopped"), InstanceStateName("running"), false},
		{"terminated to running", InstanceStateName("terminated"), InstanceStateName("running"), false},
		{"terminated to pending", InstanceStateName("terminated"), InstanceStateName("pending"), false},
		{"terminated to stopped", InstanceStateName("terminated"), InstanceStateName("stopped"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidInstanceTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidInstanceTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestIsValidVolumeTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     VolumeState
		to       VolumeState
		expected bool
	}{
		// Valid transitions
		{"creating to available", VolumeState("creating"), VolumeState("available"), true},
		{"creating to error", VolumeState("creating"), VolumeState("error"), true},
		{"available to in-use", VolumeState("available"), VolumeState("in-use"), true},
		{"available to deleting", VolumeState("available"), VolumeState("deleting"), true},
		{"in-use to available", VolumeState("in-use"), VolumeState("available"), true},
		{"in-use to deleting", VolumeState("in-use"), VolumeState("deleting"), true},
		{"deleting to deleted", VolumeState("deleting"), VolumeState("deleted"), true},
		{"error to deleting", VolumeState("error"), VolumeState("deleting"), true},

		// Invalid transitions
		{"creating to in-use", VolumeState("creating"), VolumeState("in-use"), false},
		{"available to creating", VolumeState("available"), VolumeState("creating"), false},
		{"deleted to available", VolumeState("deleted"), VolumeState("available"), false},
		{"deleted to creating", VolumeState("deleted"), VolumeState("creating"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVolumeTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidVolumeTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestIsValidVpcTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     VpcState
		to       VpcState
		expected bool
	}{
		{"pending to available", VpcState("pending"), VpcState("available"), true},
		{"available to pending", VpcState("available"), VpcState("pending"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVpcTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidVpcTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestIsValidSubnetTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     SubnetState
		to       SubnetState
		expected bool
	}{
		{"pending to available", SubnetState("pending"), SubnetState("available"), true},
		{"available to pending", SubnetState("available"), SubnetState("pending"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSubnetTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidSubnetTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestIsValidAttachmentTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     AttachmentStatus
		to       AttachmentStatus
		expected bool
	}{
		{"attaching to attached", AttachmentStatus("attaching"), AttachmentStatus("attached"), true},
		{"attached to detaching", AttachmentStatus("attached"), AttachmentStatus("detaching"), true},
		{"detaching to detached", AttachmentStatus("detaching"), AttachmentStatus("detached"), true},
		{"detached to attaching", AttachmentStatus("detached"), AttachmentStatus("attaching"), true},

		{"attached to attaching", AttachmentStatus("attached"), AttachmentStatus("attaching"), false},
		{"detached to attached", AttachmentStatus("detached"), AttachmentStatus("attached"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAttachmentTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidAttachmentTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestInstanceStateCode(t *testing.T) {
	tests := []struct {
		state InstanceStateName
		code  int32
	}{
		{InstanceStateName("pending"), 0},
		{InstanceStateName("running"), 16},
		{InstanceStateName("shutting-down"), 32},
		{InstanceStateName("terminated"), 48},
		{InstanceStateName("stopping"), 64},
		{InstanceStateName("stopped"), 80},
		{InstanceStateName("unknown"), -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			result := InstanceStateCode(tt.state)
			if result != tt.code {
				t.Errorf("InstanceStateCode(%s) = %d, want %d", tt.state, result, tt.code)
			}
		})
	}
}

func TestResourceStateManager_GetOrCreateResourceState(t *testing.T) {
	mgr := NewResourceStateManager()

	// First call should create new state
	rs1 := mgr.GetOrCreateResourceState("instances:i-123")
	if rs1 == nil {
		t.Fatal("GetOrCreateResourceState returned nil")
	}

	// Second call should return same state
	rs2 := mgr.GetOrCreateResourceState("instances:i-123")
	if rs1 != rs2 {
		t.Error("GetOrCreateResourceState should return same state for same key")
	}

	// Different resource should get different state
	rs3 := mgr.GetOrCreateResourceState("instances:i-456")
	if rs1 == rs3 {
		t.Error("GetOrCreateResourceState should return different state for different key")
	}
}

func TestResourceStateManager_SetPendingTransition(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "instances:i-123"

	// Set a pending transition
	cancelCh := mgr.SetPendingTransition(resourceKey, "running")
	if cancelCh == nil {
		t.Fatal("SetPendingTransition returned nil channel")
	}

	// Verify the resource state has the pending transition
	rs := mgr.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	if rs.pendingTransition == nil {
		t.Error("Resource should have pending transition")
	}
	if rs.pendingTransition.targetState != "running" {
		t.Errorf("Expected target state 'running', got '%s'", rs.pendingTransition.targetState)
	}
	rs.mu.Unlock()
}

func TestResourceStateManager_SetPendingTransition_CancelsPrevious(t *testing.T) {
	mgr := NewResourceStateManager()
	resourceKey := "instances:i-123"

	// Set first transition
	cancelCh1 := mgr.SetPendingTransition(resourceKey, "running")

	// Set second transition - should cancel first
	cancelCh2 := mgr.SetPendingTransition(resourceKey, "stopped")

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
	resourceKey := "instances:i-123"

	// Set a pending transition
	cancelCh := mgr.SetPendingTransition(resourceKey, "running")

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
	resourceKey := "instances:i-123"

	// Set a pending transition
	mgr.SetPendingTransition(resourceKey, "running")

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
	resourceKey := "instances:i-123"

	// Create resource state and set pending transition
	mgr.SetPendingTransition(resourceKey, "running")

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
	resourceKey := "instances:i-123"

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
				mgr.SetPendingTransition(resourceKey, "running")
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

func TestNewInstanceStateError(t *testing.T) {
	err := NewInstanceStateError("i-123", InstanceStateName("running"), InstanceStateName("pending"))

	if err.ResourceType != "instance" {
		t.Errorf("Expected ResourceType 'instance', got '%s'", err.ResourceType)
	}
	if err.ResourceID != "i-123" {
		t.Errorf("Expected ResourceID 'i-123', got '%s'", err.ResourceID)
	}
	if err.FromState != "running" {
		t.Errorf("Expected FromState 'running', got '%s'", err.FromState)
	}
	if err.ToState != "pending" {
		t.Errorf("Expected ToState 'pending', got '%s'", err.ToState)
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestNewVolumeStateError(t *testing.T) {
	err := NewVolumeStateError("vol-123", VolumeState("available"), VolumeState("creating"))

	if err.ResourceType != "volume" {
		t.Errorf("Expected ResourceType 'volume', got '%s'", err.ResourceType)
	}
	if err.ResourceID != "vol-123" {
		t.Errorf("Expected ResourceID 'vol-123', got '%s'", err.ResourceID)
	}
	if err.FromState != "available" {
		t.Errorf("Expected FromState 'available', got '%s'", err.FromState)
	}
	if err.ToState != "creating" {
		t.Errorf("Expected ToState 'creating', got '%s'", err.ToState)
	}
}
