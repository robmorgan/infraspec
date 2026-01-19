package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// describeInstanceTypes returns information about available EC2 instance types.
// This implementation returns a curated list of common instance types for testing.
func (s *EC2Service) describeInstanceTypes(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// Parse optional InstanceTypes filter
	requestedTypes := s.parseInstanceTypes(params)

	// Get all available instance types
	instanceTypes := s.getAvailableInstanceTypes()

	// Filter by requested types if specified
	if len(requestedTypes) > 0 {
		filtered := []InstanceTypeInfo{}
		requestedSet := make(map[string]bool)
		for _, t := range requestedTypes {
			requestedSet[t] = true
		}
		for _, it := range instanceTypes {
			if requestedSet[string(it.InstanceType)] {
				filtered = append(filtered, it)
			}
		}
		instanceTypes = filtered
	}

	return s.describeInstanceTypesResponse(instanceTypes)
}

// parseInstanceTypes extracts instance type names from the request parameters
func (s *EC2Service) parseInstanceTypes(params map[string]interface{}) []string {
	var types []string

	// Check for InstanceTypes.N parameters (numbered format)
	for i := 1; ; i++ {
		key := fmt.Sprintf("InstanceType.%d", i)
		if val, ok := params[key].(string); ok && val != "" {
			types = append(types, val)
		} else {
			break
		}
	}

	// Also check for InstanceTypes array format
	if arr, ok := params["InstanceTypes"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				types = append(types, s)
			}
		}
	}

	return types
}

// getAvailableInstanceTypes returns the list of instance types supported by the emulator
func (s *EC2Service) getAvailableInstanceTypes() []InstanceTypeInfo {
	return []InstanceTypeInfo{
		// T3 instances - burstable
		{
			InstanceType:                  InstanceType("t3.micro"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(1024),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		{
			InstanceType:                  InstanceType("t3.small"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(2048),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		{
			InstanceType:                  InstanceType("t3.medium"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(4096),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		{
			InstanceType:                  InstanceType("t3.large"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(8192),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		// T2 instances - burstable (older generation)
		{
			InstanceType:                  InstanceType("t2.micro"),
			CurrentGeneration:             helpers.BoolPtr(false),
			FreeTierEligible:              helpers.BoolPtr(true),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("xen"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(1),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(1024),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		{
			InstanceType:                  InstanceType("t2.small"),
			CurrentGeneration:             helpers.BoolPtr(false),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(true),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("xen"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(1),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(2048),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		// M5 instances - general purpose
		{
			InstanceType:                  InstanceType("m5.large"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(false),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(8192),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		{
			InstanceType:                  InstanceType("m5.xlarge"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(false),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(4),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(16384),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		// C5 instances - compute optimized
		{
			InstanceType:                  InstanceType("c5.large"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(false),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(4096),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
		// R5 instances - memory optimized
		{
			InstanceType:                  InstanceType("r5.large"),
			CurrentGeneration:             helpers.BoolPtr(true),
			FreeTierEligible:              helpers.BoolPtr(false),
			BurstablePerformanceSupported: helpers.BoolPtr(false),
			BareMetal:                     helpers.BoolPtr(false),
			Hypervisor:                    InstanceTypeHypervisor("nitro"),
			VCpuInfo: &VCpuInfo{
				DefaultVCpus: helpers.Int32Ptr(2),
			},
			MemoryInfo: &MemoryInfo{
				SizeInMiB: helpers.Int64Ptr(16384),
			},
			ProcessorInfo: &ProcessorInfo{
				SupportedArchitectures: []ArchitectureType{ArchitectureType("x86_64")},
			},
			SupportedRootDeviceTypes:     []RootDeviceType{RootDeviceType("ebs")},
			SupportedVirtualizationTypes: []VirtualizationType{VirtualizationType("hvm")},
		},
	}
}
