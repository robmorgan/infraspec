package ec2

import (
	"fmt"
	"net"
	"strings"
)

// ==================== Parameter Helpers ====================

// getStringParamValue extracts a string parameter with a default fallback
func getStringParamValue(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

// getIntParam extracts an integer parameter with a default fallback
func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(string); ok {
		var result int
		if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
			return result
		}
	}
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	if val, ok := params[key].(int); ok {
		return val
	}
	return defaultValue
}

// ==================== CIDR Validation ====================

// isValidCIDR validates that a string is a well-formed CIDR block
func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// ==================== Pointer Comparison Helpers ====================

// stringPtrEqual compares two string pointers for equality
func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// int32PtrEqual compares two int32 pointers for equality
func int32PtrEqual(a, b *int32) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// ==================== Resource Type Detection ====================

// getResourceType determines the EC2 resource type from a resource ID prefix
func getResourceType(resourceId string) ResourceType {
	switch {
	case strings.HasPrefix(resourceId, "i-"):
		return ResourceType("instance")
	case strings.HasPrefix(resourceId, "vol-"):
		return ResourceType("volume")
	case strings.HasPrefix(resourceId, "vpc-"):
		return ResourceType("vpc")
	case strings.HasPrefix(resourceId, "subnet-"):
		return ResourceType("subnet")
	case strings.HasPrefix(resourceId, "sg-"):
		return ResourceType("security-group")
	case strings.HasPrefix(resourceId, "igw-"):
		return ResourceType("internet-gateway")
	case strings.HasPrefix(resourceId, "lt-"):
		return ResourceType("launch-template")
	default:
		return ResourceType("instance")
	}
}

// ==================== Parameter Parsers ====================

// parseInstanceIds extracts instance IDs from request parameters
func (s *EC2Service) parseInstanceIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "InstanceId")
}

// parseVpcIds extracts VPC IDs from request parameters
func (s *EC2Service) parseVpcIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "VpcId")
}

// parseSubnetIds extracts subnet IDs from request parameters
func (s *EC2Service) parseSubnetIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "SubnetId")
}

// parseSecurityGroupIds extracts security group IDs from request parameters
func (s *EC2Service) parseSecurityGroupIds(params map[string]interface{}) []string {
	ids := s.parseIndexedParams(params, "SecurityGroupId")
	if len(ids) == 0 {
		ids = s.parseIndexedParams(params, "GroupId")
	}
	return ids
}

// parseInternetGatewayIds extracts internet gateway IDs from request parameters
func (s *EC2Service) parseInternetGatewayIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "InternetGatewayId")
}

// parseImageIds extracts AMI IDs from request parameters
func (s *EC2Service) parseImageIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "ImageId")
}

// parseVolumeIds extracts volume IDs from request parameters
func (s *EC2Service) parseVolumeIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "VolumeId")
}

// parseKeyNames extracts key pair names from request parameters
func (s *EC2Service) parseKeyNames(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "KeyName")
}

// parseLaunchTemplateIds extracts launch template IDs from request parameters
func (s *EC2Service) parseLaunchTemplateIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "LaunchTemplateId")
}

// parseResourceIds extracts resource IDs from request parameters
func (s *EC2Service) parseResourceIds(params map[string]interface{}) []string {
	return s.parseIndexedParams(params, "ResourceId")
}

// parseIndexedParams extracts a list of values from indexed parameters (e.g., InstanceId.1, InstanceId.2)
func (s *EC2Service) parseIndexedParams(params map[string]interface{}, prefix string) []string {
	var values []string

	// Check for direct value
	if val, ok := params[prefix].(string); ok && val != "" {
		values = append(values, val)
	}

	// Check for indexed values (e.g., InstanceId.1, InstanceId.2)
	for i := 1; i <= 100; i++ {
		key := fmt.Sprintf("%s.%d", prefix, i)
		if val, ok := params[key].(string); ok && val != "" {
			values = append(values, val)
		} else {
			break
		}
	}

	return values
}

// parseTags extracts tags from request parameters
func (s *EC2Service) parseTags(params map[string]interface{}) []Tag {
	var tags []Tag

	for i := 1; i <= 50; i++ {
		keyParam := fmt.Sprintf("Tag.%d.Key", i)
		valueParam := fmt.Sprintf("Tag.%d.Value", i)

		key, hasKey := params[keyParam].(string)
		value, hasValue := params[valueParam].(string)

		if !hasKey {
			break
		}

		tag := Tag{Key: &key}
		if hasValue {
			tag.Value = &value
		}
		tags = append(tags, tag)
	}

	return tags
}

// parseTagSpecifications extracts tags from TagSpecification parameters used in resource creation APIs.
// Returns tags for a specific resource type, or all tags if resourceType is empty.
// Format: TagSpecification.N.ResourceType and TagSpecification.N.Tag.M.Key/Value
func (s *EC2Service) parseTagSpecifications(params map[string]interface{}, resourceType string) []Tag {
	var tags []Tag

	for i := 1; i <= 10; i++ {
		rtKey := fmt.Sprintf("TagSpecification.%d.ResourceType", i)
		rt, hasRT := params[rtKey].(string)

		if !hasRT {
			break
		}

		// If resourceType specified, only return tags for that type
		if resourceType != "" && rt != resourceType {
			continue
		}

		// Parse tags for this TagSpecification
		for j := 1; j <= 50; j++ {
			keyParam := fmt.Sprintf("TagSpecification.%d.Tag.%d.Key", i, j)
			valueParam := fmt.Sprintf("TagSpecification.%d.Tag.%d.Value", i, j)

			key, hasKey := params[keyParam].(string)
			value, hasValue := params[valueParam].(string)

			if !hasKey {
				break
			}

			tag := Tag{Key: &key}
			if hasValue {
				tag.Value = &value
			}
			tags = append(tags, tag)
		}
	}

	return tags
}

// parseIpPermissions extracts IP permissions from request parameters
func (s *EC2Service) parseIpPermissions(params map[string]interface{}) []IpPermission {
	var permissions []IpPermission

	for i := 1; i <= 50; i++ {
		protocolKey := fmt.Sprintf("IpPermissions.%d.IpProtocol", i)
		protocol, hasProtocol := params[protocolKey].(string)

		if !hasProtocol {
			break
		}

		perm := IpPermission{
			IpProtocol: &protocol,
		}

		// Parse FromPort
		fromPortKey := fmt.Sprintf("IpPermissions.%d.FromPort", i)
		if fromPort, ok := params[fromPortKey].(string); ok {
			var port int32
			fmt.Sscanf(fromPort, "%d", &port)
			perm.FromPort = &port
		}

		// Parse ToPort
		toPortKey := fmt.Sprintf("IpPermissions.%d.ToPort", i)
		if toPort, ok := params[toPortKey].(string); ok {
			var port int32
			fmt.Sscanf(toPort, "%d", &port)
			perm.ToPort = &port
		}

		// Parse CIDR blocks
		for j := 1; j <= 50; j++ {
			cidrKey := fmt.Sprintf("IpPermissions.%d.IpRanges.%d.CidrIp", i, j)
			if cidr, ok := params[cidrKey].(string); ok {
				perm.IpRanges = append(perm.IpRanges, IpRange{CidrIp: &cidr})
			} else {
				break
			}
		}

		permissions = append(permissions, perm)
	}

	return permissions
}

// extractFilterValue extracts a filter value from the params (handles Filter.N.Name/Value format)
func (s *EC2Service) extractFilterValue(params map[string]interface{}, filterName string) string {
	// Check for Filter.N.Name / Filter.N.Value format
	for i := 1; i <= 10; i++ {
		nameKey := fmt.Sprintf("Filter.%d.Name", i)
		valueKey := fmt.Sprintf("Filter.%d.Value.1", i)

		name, _ := params[nameKey].(string)
		if name == filterName {
			value, _ := params[valueKey].(string)
			return value
		}
	}
	return ""
}

// ==================== Security Group Rule Helpers ====================

// removeMatchingRules removes rules from existingRules that match any of the rulesToRevoke.
// A rule matches if protocol, port range, and CIDR blocks match.
func (s *EC2Service) removeMatchingRules(existingRules []IpPermission, rulesToRevoke []IpPermission) []IpPermission {
	var remainingRules []IpPermission

	for _, existing := range existingRules {
		shouldRemove := false
		for _, toRevoke := range rulesToRevoke {
			if s.rulesMatch(existing, toRevoke) {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			remainingRules = append(remainingRules, existing)
		}
	}

	return remainingRules
}

// rulesMatch checks if two IpPermission rules match based on protocol, ports, and CIDR blocks.
func (s *EC2Service) rulesMatch(a, b IpPermission) bool {
	// Compare protocol
	if !stringPtrEqual(a.IpProtocol, b.IpProtocol) {
		return false
	}

	// Compare ports (only if not protocol -1 which means all traffic)
	if a.IpProtocol != nil && *a.IpProtocol != "-1" {
		if !int32PtrEqual(a.FromPort, b.FromPort) || !int32PtrEqual(a.ToPort, b.ToPort) {
			return false
		}
	}

	// Compare CIDR blocks - for a match, all CIDR blocks in b must be present in a
	if len(b.IpRanges) > 0 {
		for _, bRange := range b.IpRanges {
			found := false
			for _, aRange := range a.IpRanges {
				if stringPtrEqual(aRange.CidrIp, bRange.CidrIp) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}
