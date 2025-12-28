package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createVolume(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	az, ok := params["AvailabilityZone"].(string)
	if !ok || az == "" {
		return s.errorResponse(400, "MissingParameter", "AvailabilityZone is required"), nil
	}

	volumeId := fmt.Sprintf("vol-%s", uuid.New().String()[:8])
	size := int32(getIntParam(params, "Size", 8))
	volumeType := getStringParamValue(params, "VolumeType", "gp2")

	volume := Volume{
		VolumeId:         &volumeId,
		AvailabilityZone: &az,
		Size:             &size,
		VolumeType:       VolumeType(volumeType),
		State:            VolumeState("creating"),
		CreateTime:       helpers.TimePtr(time.Now()),
		Encrypted:        helpers.BoolPtr(false),
	}

	if iops, ok := params["Iops"].(string); ok {
		var iopsVal int32
		fmt.Sscanf(iops, "%d", &iopsVal)
		volume.Iops = &iopsVal
	}

	if snapshotId, ok := params["SnapshotId"].(string); ok && snapshotId != "" {
		volume.SnapshotId = &snapshotId
	}

	if err := s.state.Set(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store volume"), nil
	}

	// Schedule transition to available
	s.scheduleVolumeTransition(volumeId, VolumeState("available"), 2*time.Second)

	return s.createVolumeResponse(volume)
}

func (s *EC2Service) describeVolumes(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	volumeIds := s.parseVolumeIds(params)

	var volumes []Volume

	if len(volumeIds) > 0 {
		for _, volumeId := range volumeIds {
			var volume Volume
			if err := s.state.Get(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
				return s.errorResponse(400, "InvalidVolume.NotFound", fmt.Sprintf("The volume '%s' does not exist", volumeId)), nil
			}
			volumes = append(volumes, volume)
		}
	} else {
		keys, err := s.state.List("ec2:volumes:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list volumes"), nil
		}

		for _, key := range keys {
			var volume Volume
			if err := s.state.Get(key, &volume); err == nil {
				volumes = append(volumes, volume)
			}
		}
	}

	return s.describeVolumesResponse(volumes)
}

func (s *EC2Service) attachVolume(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	volumeId, ok := params["VolumeId"].(string)
	if !ok || volumeId == "" {
		return s.errorResponse(400, "MissingParameter", "VolumeId is required"), nil
	}

	instanceId, ok := params["InstanceId"].(string)
	if !ok || instanceId == "" {
		return s.errorResponse(400, "MissingParameter", "InstanceId is required"), nil
	}

	device, ok := params["Device"].(string)
	if !ok || device == "" {
		return s.errorResponse(400, "MissingParameter", "Device is required"), nil
	}

	// Lock volume resource
	volumeResourceKey := "volumes:" + volumeId
	volumeRS := s.stateMachine.GetOrCreateResourceState(volumeResourceKey)
	volumeRS.mu.Lock()
	defer volumeRS.mu.Unlock()

	var volume Volume
	if err := s.state.Get(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(400, "InvalidVolume.NotFound", fmt.Sprintf("The volume '%s' does not exist", volumeId)), nil
	}

	// Validate volume is available
	if volume.State != VolumeState("available") {
		return s.errorResponse(400, "IncorrectState", fmt.Sprintf("Volume '%s' is not in 'available' state. Current state: %s", volumeId, volume.State)), nil
	}

	var instance Instance
	if err := s.state.Get(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
		return s.errorResponse(400, "InvalidInstanceID.NotFound", fmt.Sprintf("The instance ID '%s' does not exist", instanceId)), nil
	}

	// Validate instance is running
	if instance.State == nil || instance.State.Name != InstanceStateName("running") {
		currentState := "unknown"
		if instance.State != nil {
			currentState = string(instance.State.Name)
		}
		return s.errorResponse(400, "IncorrectInstanceState", fmt.Sprintf("Instance '%s' is not in 'running' state. Current state: %s", instanceId, currentState)), nil
	}

	volume.State = VolumeState("in-use")
	volume.Attachments = []VolumeAttachment{
		{
			VolumeId:            &volumeId,
			InstanceId:          &instanceId,
			Device:              &device,
			State:               VolumeAttachmentState("attached"),
			AttachTime:          helpers.TimePtr(time.Now()),
			DeleteOnTermination: helpers.BoolPtr(false),
		},
	}

	if err := s.state.Set(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update volume"), nil
	}

	return s.attachVolumeResponse(volume.Attachments[0])
}

func (s *EC2Service) detachVolume(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	volumeId, ok := params["VolumeId"].(string)
	if !ok || volumeId == "" {
		return s.errorResponse(400, "MissingParameter", "VolumeId is required"), nil
	}

	// Lock volume resource
	resourceKey := "volumes:" + volumeId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var volume Volume
	if err := s.state.Get(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(400, "InvalidVolume.NotFound", fmt.Sprintf("The volume '%s' does not exist", volumeId)), nil
	}

	// Validate volume is in-use
	if volume.State != VolumeState("in-use") {
		return s.errorResponse(400, "IncorrectState", fmt.Sprintf("Volume '%s' is not attached. Current state: %s", volumeId, volume.State)), nil
	}

	var detachedAttachment VolumeAttachment
	if len(volume.Attachments) > 0 {
		detachedAttachment = volume.Attachments[0]
		detachedAttachment.State = VolumeAttachmentState("detaching")
	}

	volume.State = VolumeState("available")
	volume.Attachments = []VolumeAttachment{}

	if err := s.state.Set(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update volume"), nil
	}

	return s.detachVolumeResponse(detachedAttachment)
}

func (s *EC2Service) deleteVolume(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	volumeId, ok := params["VolumeId"].(string)
	if !ok || volumeId == "" {
		return s.errorResponse(400, "MissingParameter", "VolumeId is required"), nil
	}

	// Lock volume resource
	resourceKey := "volumes:" + volumeId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var volume Volume
	if err := s.state.Get(fmt.Sprintf("ec2:volumes:%s", volumeId), &volume); err != nil {
		return s.errorResponse(400, "InvalidVolume.NotFound", fmt.Sprintf("The volume '%s' does not exist", volumeId)), nil
	}

	if volume.State == VolumeState("in-use") {
		return s.errorResponse(400, "VolumeInUse", "Volume is currently attached to an instance"), nil
	}

	s.state.Delete(fmt.Sprintf("ec2:volumes:%s", volumeId))
	s.stateMachine.RemoveResourceState(resourceKey)

	return s.deleteVolumeResponse()
}

// transitionVolumeState atomically transitions a volume to a new state with validation
func (s *EC2Service) transitionVolumeState(volumeId string, newState VolumeState) error {
	resourceKey := "volumes:" + volumeId
	rs := s.stateMachine.GetOrCreateResourceState(resourceKey)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	key := fmt.Sprintf("ec2:volumes:%s", volumeId)
	var volume Volume

	// Use atomic Update to prevent race conditions between Get and Set
	return s.state.Update(key, &volume, func() error {
		currentState := volume.State

		// Validate the transition
		if !IsValidVolumeTransition(currentState, newState) {
			return NewVolumeStateError(volumeId, currentState, newState)
		}

		volume.State = newState
		return nil
	})
}

// scheduleVolumeTransition schedules an async volume state transition
func (s *EC2Service) scheduleVolumeTransition(volumeId string, targetState VolumeState, delay time.Duration) {
	resourceKey := "volumes:" + volumeId
	cancelCh := s.stateMachine.SetPendingTransition(resourceKey, string(targetState))

	go func() {
		select {
		case <-s.shutdownCtx.Done():
			return
		case <-cancelCh:
			return
		case <-time.After(delay):
			s.transitionVolumeState(volumeId, targetState)
			s.stateMachine.ClearPendingTransition(resourceKey)
		}
	}()
}
