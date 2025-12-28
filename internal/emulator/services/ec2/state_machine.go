package ec2

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
	resources map[string]*ResourceState // key: "instances:i-xxx", "volumes:vol-xxx"
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

// ==================== Instance State Transitions ====================

// instanceTransitions defines valid state transitions for EC2 instances
var instanceTransitions = map[InstanceStateName][]InstanceStateName{
	InstanceStateName("pending"):       {InstanceStateName("running"), InstanceStateName("shutting-down")},
	InstanceStateName("running"):       {InstanceStateName("stopping"), InstanceStateName("shutting-down")},
	InstanceStateName("stopping"):      {InstanceStateName("stopped")},
	InstanceStateName("stopped"):       {InstanceStateName("pending"), InstanceStateName("shutting-down")},
	InstanceStateName("shutting-down"): {InstanceStateName("terminated")},
	InstanceStateName("terminated"):    {}, // Terminal state - no transitions allowed
}

// IsValidInstanceTransition checks if a state transition is valid for instances
func IsValidInstanceTransition(from, to InstanceStateName) bool {
	allowed, ok := instanceTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// InstanceStateCode returns the AWS state code for an instance state
func InstanceStateCode(state InstanceStateName) int32 {
	switch state {
	case InstanceStateName("pending"):
		return 0
	case InstanceStateName("running"):
		return 16
	case InstanceStateName("shutting-down"):
		return 32
	case InstanceStateName("terminated"):
		return 48
	case InstanceStateName("stopping"):
		return 64
	case InstanceStateName("stopped"):
		return 80
	default:
		return -1
	}
}

// ==================== Volume State Transitions ====================

// volumeTransitions defines valid state transitions for EBS volumes
var volumeTransitions = map[VolumeState][]VolumeState{
	VolumeState("creating"):  {VolumeState("available"), VolumeState("error")},
	VolumeState("available"): {VolumeState("in-use"), VolumeState("deleting")},
	VolumeState("in-use"):    {VolumeState("available"), VolumeState("deleting")},
	VolumeState("deleting"):  {VolumeState("deleted")},
	VolumeState("deleted"):   {}, // Terminal state
	VolumeState("error"):     {VolumeState("deleting")},
}

// IsValidVolumeTransition checks if a state transition is valid for volumes
func IsValidVolumeTransition(from, to VolumeState) bool {
	allowed, ok := volumeTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// ==================== VPC State Transitions ====================

// vpcTransitions defines valid state transitions for VPCs
var vpcTransitions = map[VpcState][]VpcState{
	VpcState("pending"):   {VpcState("available")},
	VpcState("available"): {}, // Deleted directly, no transition state
}

// IsValidVpcTransition checks if a state transition is valid for VPCs
func IsValidVpcTransition(from, to VpcState) bool {
	allowed, ok := vpcTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// ==================== Subnet State Transitions ====================

// subnetTransitions defines valid state transitions for subnets
var subnetTransitions = map[SubnetState][]SubnetState{
	SubnetState("pending"):   {SubnetState("available")},
	SubnetState("available"): {}, // Deleted directly
}

// IsValidSubnetTransition checks if a state transition is valid for subnets
func IsValidSubnetTransition(from, to SubnetState) bool {
	allowed, ok := subnetTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// ==================== Attachment State Transitions ====================

// attachmentTransitions defines valid state transitions for attachments (IGW, volumes)
var attachmentTransitions = map[AttachmentStatus][]AttachmentStatus{
	AttachmentStatus("attaching"): {AttachmentStatus("attached")},
	AttachmentStatus("attached"):  {AttachmentStatus("detaching")},
	AttachmentStatus("detaching"): {AttachmentStatus("detached")},
	AttachmentStatus("detached"):  {AttachmentStatus("attaching")},
}

// IsValidAttachmentTransition checks if a state transition is valid for attachments
func IsValidAttachmentTransition(from, to AttachmentStatus) bool {
	allowed, ok := attachmentTransitions[from]
	if !ok {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}

// ==================== Error Types ====================

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

// NewInstanceStateError creates an error for invalid instance state transitions
func NewInstanceStateError(instanceId string, from, to InstanceStateName) *StateTransitionError {
	return &StateTransitionError{
		ResourceType: "instance",
		ResourceID:   instanceId,
		FromState:    string(from),
		ToState:      string(to),
	}
}

// NewVolumeStateError creates an error for invalid volume state transitions
func NewVolumeStateError(volumeId string, from, to VolumeState) *StateTransitionError {
	return &StateTransitionError{
		ResourceType: "volume",
		ResourceID:   volumeId,
		FromState:    string(from),
		ToState:      string(to),
	}
}
