package ec2

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

// describeInstanceAttribute returns information about a specific attribute of an EC2 instance.
func (s *EC2Service) describeInstanceAttribute(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	instanceId, ok := params["InstanceId"].(string)
	if !ok || instanceId == "" {
		return s.errorResponse(400, "MissingParameter", "InstanceId is required"), nil
	}

	attribute, ok := params["Attribute"].(string)
	if !ok || attribute == "" {
		return s.errorResponse(400, "MissingParameter", "Attribute is required"), nil
	}

	// Verify instance exists
	var instance Instance
	if err := s.state.Get(fmt.Sprintf("ec2:instances:%s", instanceId), &instance); err != nil {
		return s.errorResponse(400, "InvalidInstanceID.NotFound",
			fmt.Sprintf("The instance ID '%s' does not exist", instanceId)), nil
	}

	// Build response based on requested attribute
	response := InstanceAttribute{
		InstanceId: &instanceId,
	}

	switch attribute {
	case "disableApiTermination":
		response.DisableApiTermination = &AttributeBooleanValue{Value: helpers.BoolPtr(false)}
	case "disableApiStop":
		response.DisableApiStop = &AttributeBooleanValue{Value: helpers.BoolPtr(false)}
	case "ebsOptimized":
		response.EbsOptimized = &AttributeBooleanValue{Value: helpers.BoolPtr(false)}
	case "enaSupport":
		response.EnaSupport = &AttributeBooleanValue{Value: helpers.BoolPtr(true)}
	case "instanceType":
		instType := string(instance.InstanceType)
		response.InstanceType = &AttributeValue{Value: &instType}
	case "kernel":
		response.KernelId = &AttributeValue{}
	case "ramdisk":
		response.RamdiskId = &AttributeValue{}
	case "userData":
		response.UserData = &AttributeValue{}
	case "sourceDestCheck":
		response.SourceDestCheck = &AttributeBooleanValue{Value: helpers.BoolPtr(true)}
	case "sriovNetSupport":
		response.SriovNetSupport = &AttributeValue{Value: helpers.StringPtr("simple")}
	case "blockDeviceMapping":
		response.BlockDeviceMappings = []InstanceBlockDeviceMapping{}
	case "productCodes":
		response.ProductCodes = []ProductCode{}
	case "groupSet":
		// Convert security groups to GroupIdentifier for response
		groups := make([]GroupIdentifier, 0, len(instance.SecurityGroups))
		for _, sg := range instance.SecurityGroups {
			groups = append(groups, GroupIdentifier{
				GroupId:   sg.GroupId,
				GroupName: sg.GroupName,
			})
		}
		response.Groups = groups
	case "instanceInitiatedShutdownBehavior":
		response.InstanceInitiatedShutdownBehavior = &AttributeValue{Value: helpers.StringPtr("stop")}
	case "rootDeviceName":
		response.RootDeviceName = &AttributeValue{Value: instance.RootDeviceName}
	default:
		return s.errorResponse(400, "InvalidParameterValue",
			fmt.Sprintf("Unsupported attribute: %s", attribute)), nil
	}

	return s.describeInstanceAttributeResponse(response)
}
