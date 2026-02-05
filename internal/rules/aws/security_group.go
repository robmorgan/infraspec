package aws

import (
	"fmt"

	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/rules"
)

// Common port numbers for security rules.
const (
	PortSSH      = 22
	PortRDP      = 3389
	PortMySQL    = 3306
	PortPostgres = 5432
)

// Protocol constants.
const (
	ProtocolTCP = "tcp"
	ProtocolAll = "-1"
	ProtocolAny = "all"
)

// checkIngressPortRule is a helper that checks if any ingress rule allows
// public access to the specified port.
func checkIngressPortRule(resource *plan.ResourceChange, port int, portName string) (passed bool, message string) {
	ingressRules := extractIngressRules(resource)
	for _, ir := range ingressRules {
		if !portInRange(ir.FromPort, ir.ToPort, port) {
			continue
		}
		if !isTCPOrAllProtocol(ir.Protocol) {
			continue
		}
		if containsPublicCIDR(ir.CIDRBlocks) || containsPublicCIDR(ir.IPv6CIDRBlocks) {
			return false, fmt.Sprintf("Security group allows %s (port %d) from public internet (0.0.0.0/0 or ::/0)", portName, port)
		}
	}
	return true, fmt.Sprintf("Security group does not allow public %s access", portName)
}

// checkSGRuleIngressPort is a helper that checks aws_security_group_rule resources
// for public access to the specified port.
func checkSGRuleIngressPort(resource *plan.ResourceChange, port int, portName string) (passed bool, message string) {
	ir := extractSecurityGroupRuleIngress(resource)
	if ir == nil {
		return true, "Not an ingress rule"
	}

	if !portInRange(ir.FromPort, ir.ToPort, port) {
		return true, fmt.Sprintf("Rule does not include port %d", port)
	}
	if !isTCPOrAllProtocol(ir.Protocol) {
		return true, "Rule does not use TCP protocol"
	}
	if containsPublicCIDR(ir.CIDRBlocks) || containsPublicCIDR(ir.IPv6CIDRBlocks) {
		return false, fmt.Sprintf("Security group rule allows %s (port %d) from public internet (0.0.0.0/0 or ::/0)", portName, port)
	}
	return true, fmt.Sprintf("Security group rule does not allow public %s access", portName)
}

// isTCPOrAllProtocol checks if the protocol is TCP or allows all traffic.
func isTCPOrAllProtocol(protocol string) bool {
	return protocol == ProtocolTCP || protocol == ProtocolAll || protocol == ProtocolAny
}

// SGNoPublicSSHRule checks that security groups don't allow SSH (port 22) from 0.0.0.0/0.
type SGNoPublicSSHRule struct {
	BaseRule
}

// NewSGNoPublicSSHRule creates a new rule that checks for public SSH access.
func NewSGNoPublicSSHRule() *SGNoPublicSSHRule {
	return &SGNoPublicSSHRule{
		BaseRule: BaseRule{
			id:           "aws-sg-no-public-ssh",
			description:  "Security groups should not allow SSH (port 22) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group",
		},
	}
}

// Check evaluates the rule against an aws_security_group resource.
func (r *SGNoPublicSSHRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkIngressPortRule(resource, PortSSH, "SSH")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGRuleNoPublicSSHRule checks aws_security_group_rule resources for public SSH.
type SGRuleNoPublicSSHRule struct {
	BaseRule
}

// NewSGRuleNoPublicSSHRule creates a new rule for security_group_rule resources.
func NewSGRuleNoPublicSSHRule() *SGRuleNoPublicSSHRule {
	return &SGRuleNoPublicSSHRule{
		BaseRule: BaseRule{
			id:           "aws-sg-rule-no-public-ssh",
			description:  "Security group rules should not allow SSH (port 22) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group_rule",
		},
	}
}

// Check evaluates the rule against an aws_security_group_rule resource.
func (r *SGRuleNoPublicSSHRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkSGRuleIngressPort(resource, PortSSH, "SSH")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGNoPublicRDPRule checks that security groups don't allow RDP (port 3389) from 0.0.0.0/0.
type SGNoPublicRDPRule struct {
	BaseRule
}

// NewSGNoPublicRDPRule creates a new rule that checks for public RDP access.
func NewSGNoPublicRDPRule() *SGNoPublicRDPRule {
	return &SGNoPublicRDPRule{
		BaseRule: BaseRule{
			id:           "aws-sg-no-public-rdp",
			description:  "Security groups should not allow RDP (port 3389) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group",
		},
	}
}

// Check evaluates the rule against an aws_security_group resource.
func (r *SGNoPublicRDPRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkIngressPortRule(resource, PortRDP, "RDP")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGRuleNoPublicRDPRule checks aws_security_group_rule resources for public RDP.
type SGRuleNoPublicRDPRule struct {
	BaseRule
}

// NewSGRuleNoPublicRDPRule creates a new rule for security_group_rule resources.
func NewSGRuleNoPublicRDPRule() *SGRuleNoPublicRDPRule {
	return &SGRuleNoPublicRDPRule{
		BaseRule: BaseRule{
			id:           "aws-sg-rule-no-public-rdp",
			description:  "Security group rules should not allow RDP (port 3389) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group_rule",
		},
	}
}

// Check evaluates the rule against an aws_security_group_rule resource.
func (r *SGRuleNoPublicRDPRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkSGRuleIngressPort(resource, PortRDP, "RDP")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGNoPublicMySQLRule checks that security groups don't allow MySQL (port 3306) from 0.0.0.0/0.
type SGNoPublicMySQLRule struct {
	BaseRule
}

// NewSGNoPublicMySQLRule creates a new rule that checks for public MySQL access.
func NewSGNoPublicMySQLRule() *SGNoPublicMySQLRule {
	return &SGNoPublicMySQLRule{
		BaseRule: BaseRule{
			id:           "aws-sg-no-public-mysql",
			description:  "Security groups should not allow MySQL (port 3306) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group",
		},
	}
}

// Check evaluates the rule against an aws_security_group resource.
func (r *SGNoPublicMySQLRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkIngressPortRule(resource, PortMySQL, "MySQL")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGRuleNoPublicMySQLRule checks aws_security_group_rule resources for public MySQL.
type SGRuleNoPublicMySQLRule struct {
	BaseRule
}

// NewSGRuleNoPublicMySQLRule creates a new rule for security_group_rule resources.
func NewSGRuleNoPublicMySQLRule() *SGRuleNoPublicMySQLRule {
	return &SGRuleNoPublicMySQLRule{
		BaseRule: BaseRule{
			id:           "aws-sg-rule-no-public-mysql",
			description:  "Security group rules should not allow MySQL (port 3306) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group_rule",
		},
	}
}

// Check evaluates the rule against an aws_security_group_rule resource.
func (r *SGRuleNoPublicMySQLRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkSGRuleIngressPort(resource, PortMySQL, "MySQL")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGNoPublicPostgresRule checks that security groups don't allow PostgreSQL (port 5432) from 0.0.0.0/0.
type SGNoPublicPostgresRule struct {
	BaseRule
}

// NewSGNoPublicPostgresRule creates a new rule that checks for public PostgreSQL access.
func NewSGNoPublicPostgresRule() *SGNoPublicPostgresRule {
	return &SGNoPublicPostgresRule{
		BaseRule: BaseRule{
			id:           "aws-sg-no-public-postgres",
			description:  "Security groups should not allow PostgreSQL (port 5432) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group",
		},
	}
}

// Check evaluates the rule against an aws_security_group resource.
func (r *SGNoPublicPostgresRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkIngressPortRule(resource, PortPostgres, "PostgreSQL")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGRuleNoPublicPostgresRule checks aws_security_group_rule resources for public PostgreSQL.
type SGRuleNoPublicPostgresRule struct {
	BaseRule
}

// NewSGRuleNoPublicPostgresRule creates a new rule for security_group_rule resources.
func NewSGRuleNoPublicPostgresRule() *SGRuleNoPublicPostgresRule {
	return &SGRuleNoPublicPostgresRule{
		BaseRule: BaseRule{
			id:           "aws-sg-rule-no-public-postgres",
			description:  "Security group rules should not allow PostgreSQL (port 5432) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group_rule",
		},
	}
}

// Check evaluates the rule against an aws_security_group_rule resource.
func (r *SGRuleNoPublicPostgresRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	passed, msg := checkSGRuleIngressPort(resource, PortPostgres, "PostgreSQL")
	if passed {
		return r.passResult(resource, msg), nil
	}
	return r.failResult(resource, msg), nil
}

// SGNoUnrestrictedIngressRule checks that security groups don't allow all traffic (protocol -1) from 0.0.0.0/0.
type SGNoUnrestrictedIngressRule struct {
	BaseRule
}

// NewSGNoUnrestrictedIngressRule creates a new rule that checks for unrestricted ingress.
func NewSGNoUnrestrictedIngressRule() *SGNoUnrestrictedIngressRule {
	return &SGNoUnrestrictedIngressRule{
		BaseRule: BaseRule{
			id:           "aws-sg-no-unrestricted-ingress",
			description:  "Security groups should not allow all traffic (protocol -1) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group",
		},
	}
}

// Check evaluates the rule against an aws_security_group resource.
func (r *SGNoUnrestrictedIngressRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	ingressRules := extractIngressRules(resource)
	for _, ir := range ingressRules {
		if !isAllTrafficProtocol(ir.Protocol) {
			continue
		}
		if containsPublicCIDR(ir.CIDRBlocks) || containsPublicCIDR(ir.IPv6CIDRBlocks) {
			return r.failResult(resource, fmt.Sprintf("Security group allows all traffic (protocol %s) from public internet (0.0.0.0/0 or ::/0)", ir.Protocol)), nil
		}
	}
	return r.passResult(resource, "Security group does not allow unrestricted public ingress"), nil
}

// SGRuleNoUnrestrictedIngressRule checks aws_security_group_rule resources for unrestricted ingress.
type SGRuleNoUnrestrictedIngressRule struct {
	BaseRule
}

// NewSGRuleNoUnrestrictedIngressRule creates a new rule for security_group_rule resources.
func NewSGRuleNoUnrestrictedIngressRule() *SGRuleNoUnrestrictedIngressRule {
	return &SGRuleNoUnrestrictedIngressRule{
		BaseRule: BaseRule{
			id:           "aws-sg-rule-no-unrestricted-ingress",
			description:  "Security group rules should not allow all traffic (protocol -1) from 0.0.0.0/0 or ::/0",
			severity:     rules.Critical,
			resourceType: "aws_security_group_rule",
		},
	}
}

// Check evaluates the rule against an aws_security_group_rule resource.
func (r *SGRuleNoUnrestrictedIngressRule) Check(resource *plan.ResourceChange) (*rules.Result, error) {
	ir := extractSecurityGroupRuleIngress(resource)
	if ir == nil {
		return r.passResult(resource, "Not an ingress rule"), nil
	}

	if !isAllTrafficProtocol(ir.Protocol) {
		return r.passResult(resource, "Rule does not allow all protocols"), nil
	}
	if containsPublicCIDR(ir.CIDRBlocks) || containsPublicCIDR(ir.IPv6CIDRBlocks) {
		return r.failResult(resource, fmt.Sprintf("Security group rule allows all traffic (protocol %s) from public internet (0.0.0.0/0 or ::/0)", ir.Protocol)), nil
	}
	return r.passResult(resource, "Security group rule does not allow unrestricted public ingress"), nil
}
