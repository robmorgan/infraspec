package ec2

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) describeInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	instanceIds := s.parseInstanceIds(params)

	var reservations []Reservation

	if len(instanceIds) > 0 {
		// Get specific instances
		instances := make([]Instance, 0)
		for _, instanceId := range instanceIds {
			var instance Instance
			if err := s.state.Get(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
				return s.errorResponse(400, "InvalidInstanceID.NotFound", fmt.Sprintf("The instance ID '%s' does not exist", instanceId)), nil
			}
			// Merge tags from separate tag storage
			s.mergeResourceTags(&instance.Tags, instanceId)
			instances = append(instances, instance)
		}
		if len(instances) > 0 {
			reservations = append(reservations, Reservation{
				ReservationId: helpers.StringPtr("r-synthetic"),
				OwnerId:       helpers.StringPtr("123456789012"),
				Instances:     instances,
			})
		}
	} else {
		// List all instances
		keys, err := s.state.List("ec2:instances:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list instances"), nil
		}

		instances := make([]Instance, 0)
		for _, key := range keys {
			var instance Instance
			if err := s.state.Get(key, &instance); err == nil {
				// Merge tags from separate tag storage
				if instance.InstanceId != nil {
					s.mergeResourceTags(&instance.Tags, *instance.InstanceId)
				}
				instances = append(instances, instance)
			}
		}
		if len(instances) > 0 {
			reservations = append(reservations, Reservation{
				ReservationId: helpers.StringPtr("r-synthetic"),
				OwnerId:       helpers.StringPtr("123456789012"),
				Instances:     instances,
			})
		}
	}

	return s.describeInstancesResponse(reservations)
}
