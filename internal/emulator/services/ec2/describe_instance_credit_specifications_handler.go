package ec2

import (
	"context"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// describeInstanceCreditSpecifications returns CPU credit option for burstable performance instances.
// For T2/T3 instances, this returns the credit specification (standard or unlimited).
func (s *EC2Service) describeInstanceCreditSpecifications(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Parse instance IDs from the request
	instanceIds := s.parseInstanceIds(params)

	// If no instance IDs provided, return empty result
	if len(instanceIds) == 0 {
		return s.describeInstanceCreditSpecificationsResponse([]InstanceCreditSpecification{})
	}

	creditSpecs := []InstanceCreditSpecification{}

	for _, instanceId := range instanceIds {
		// Check if instance exists
		var instance Instance
		if err := s.state.Get(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
			// Instance not found - skip it (AWS doesn't error for non-existent instances)
			continue
		}

		// Only burstable instances (T2, T3, T3a, T4g) have credit specifications
		instanceType := string(instance.InstanceType)
		if isBurstableInstance(instanceType) {
			// Default to "standard" for T2 and "unlimited" for T3/T3a/T4g
			cpuCredits := "standard"
			if strings.HasPrefix(instanceType, "t3") || strings.HasPrefix(instanceType, "t4g") {
				cpuCredits = "unlimited"
			}

			creditSpecs = append(creditSpecs, InstanceCreditSpecification{
				InstanceId: helpers.StringPtr(instanceId),
				CpuCredits: helpers.StringPtr(cpuCredits),
			})
		}
	}

	return s.describeInstanceCreditSpecificationsResponse(creditSpecs)
}

// isBurstableInstance returns true if the instance type is a burstable performance instance
func isBurstableInstance(instanceType string) bool {
	burstablePrefixes := []string{"t2.", "t3.", "t3a.", "t4g."}
	for _, prefix := range burstablePrefixes {
		if strings.HasPrefix(instanceType, prefix) {
			return true
		}
	}
	return false
}
