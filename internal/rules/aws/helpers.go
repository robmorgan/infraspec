package aws

import (
	"encoding/json"

	"github.com/robmorgan/infraspec/internal/plan"
)

// isPublicCIDR returns true if the CIDR represents unrestricted access.
func isPublicCIDR(cidr string) bool {
	return cidr == "0.0.0.0/0" || cidr == "::/0"
}

// containsPublicCIDR checks if any CIDR in the slice is public.
func containsPublicCIDR(cidrs []string) bool {
	for _, cidr := range cidrs {
		if isPublicCIDR(cidr) {
			return true
		}
	}
	return false
}

// IngressRule represents a parsed ingress rule from a security group.
type IngressRule struct {
	FromPort       int
	ToPort         int
	Protocol       string
	CIDRBlocks     []string
	IPv6CIDRBlocks []string
}

// extractIngressRules extracts ingress rules from an aws_security_group resource.
func extractIngressRules(resource *plan.ResourceChange) []IngressRule {
	ingress, ok := resource.GetAfter("ingress")
	if !ok {
		return nil
	}

	ingressSlice, ok := ingress.([]interface{})
	if !ok {
		return nil
	}

	var rules []IngressRule
	for _, rule := range ingressSlice {
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			continue
		}

		ir := parseIngressRuleMap(ruleMap)
		rules = append(rules, ir)
	}

	return rules
}

// parseIngressRuleMap parses a single ingress rule map into an IngressRule.
func parseIngressRuleMap(ruleMap map[string]interface{}) IngressRule {
	ir := IngressRule{
		FromPort: toInt(ruleMap["from_port"]),
		ToPort:   toInt(ruleMap["to_port"]),
		Protocol: toString(ruleMap["protocol"]),
	}

	ir.CIDRBlocks = extractStringSlice(ruleMap, "cidr_blocks")
	ir.IPv6CIDRBlocks = extractStringSlice(ruleMap, "ipv6_cidr_blocks")

	return ir
}

// extractStringSlice extracts a string slice from a map value.
func extractStringSlice(m map[string]interface{}, key string) []string {
	cidrs, ok := m[key].([]interface{})
	if !ok {
		return nil
	}

	var result []string
	for _, c := range cidrs {
		if s, ok := c.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// extractSecurityGroupRuleIngress extracts ingress data from an aws_security_group_rule resource.
// Returns nil if this is not an ingress rule.
func extractSecurityGroupRuleIngress(resource *plan.ResourceChange) *IngressRule {
	ruleType := resource.GetAfterString("type")
	if ruleType != "ingress" {
		return nil
	}

	ir := &IngressRule{
		FromPort:       resource.GetAfterInt("from_port"),
		ToPort:         resource.GetAfterInt("to_port"),
		Protocol:       resource.GetAfterString("protocol"),
		CIDRBlocks:     resource.GetAfterStringSlice("cidr_blocks"),
		IPv6CIDRBlocks: resource.GetAfterStringSlice("ipv6_cidr_blocks"),
	}

	return ir
}

// IAMPolicy represents a parsed IAM policy document.
type IAMPolicy struct {
	Version   string         `json:"Version"`
	Statement []IAMStatement `json:"Statement"`
}

// IAMStatement represents a statement in an IAM policy.
type IAMStatement struct {
	Effect   string      `json:"Effect"`
	Action   interface{} `json:"Action"`   // Can be string or []string
	Resource interface{} `json:"Resource"` // Can be string or []string
}

// parseIAMPolicy parses an IAM policy JSON string.
func parseIAMPolicy(policyJSON string) (*IAMPolicy, error) {
	var policy IAMPolicy
	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// hasWildcardActionAndResource checks if any statement has Action:"*" with Resource:"*".
func hasWildcardActionAndResource(policy *IAMPolicy) bool {
	for _, stmt := range policy.Statement {
		if stmt.Effect != "Allow" {
			continue
		}
		if hasWildcard(stmt.Action) && hasWildcard(stmt.Resource) {
			return true
		}
	}
	return false
}

// hasWildcard checks if the value is "*" or contains "*".
func hasWildcard(val interface{}) bool {
	switch v := val.(type) {
	case string:
		return v == "*"
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s == "*" {
				return true
			}
		}
	}
	return false
}

// toInt converts an interface{} to int.
func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// toString converts an interface{} to string.
func toString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// portInRange checks if a target port falls within the from/to range.
func portInRange(fromPort, toPort, targetPort int) bool {
	// Protocol -1 means all ports
	if fromPort == 0 && toPort == 0 {
		return true
	}
	return fromPort <= targetPort && targetPort <= toPort
}

// isAllTrafficProtocol checks if the protocol allows all traffic.
func isAllTrafficProtocol(protocol string) bool {
	return protocol == ProtocolAll || protocol == ProtocolAny
}
