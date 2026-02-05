// Package rules provides rule definitions and loading for the gatekeeper.
package rules

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// HCLRuleFile represents the top-level structure of an HCL rules file
type HCLRuleFile struct {
	Rules  []HCLRule `hcl:"rule,block"`
	Remain hcl.Body  `hcl:",remain"`
}

// HCLRule represents a rule block in HCL format
type HCLRule struct {
	ID           string        `hcl:",label"`
	Name         string        `hcl:"name,attr"`
	Description  string        `hcl:"description,optional"`
	SeverityStr  string        `hcl:"severity,attr"`
	ResourceType string        `hcl:"resource_type,attr"`
	Condition    *HCLCondition `hcl:"condition,block"`
	Message      string        `hcl:"message,attr"`
	Remediation  string        `hcl:"remediation,optional"`
	Tags         []string      `hcl:"tags,optional"`
}

// HCLCondition represents a condition block in HCL format
type HCLCondition struct {
	Check *HCLCheck      `hcl:"check,block"`
	All   *HCLConditions `hcl:"all,block"`
	Any   *HCLConditions `hcl:"any,block"`
	Not   *HCLCondition  `hcl:"not,block"`
}

// HCLConditions represents a logical grouping of conditions (all/any)
type HCLConditions struct {
	Checks []*HCLCheck    `hcl:"check,block"`
	All    *HCLConditions `hcl:"all,block"`
	Any    *HCLConditions `hcl:"any,block"`
	Not    *HCLCondition  `hcl:"not,block"`
}

// HCLCheck represents a single check block
type HCLCheck struct {
	Attribute string    `hcl:"attribute,attr"`
	Operator  string    `hcl:"operator,attr"`
	Value     cty.Value `hcl:"value,optional"`
}

// LoadFromHCLFile loads rules from an HCL file
func LoadFromHCLFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromHCLBytes(data, path)
}

// LoadFromHCLBytes loads rules from HCL bytes
func LoadFromHCLBytes(data []byte, filename string) ([]Rule, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	var hclFile HCLRuleFile
	diags = gohcl.DecodeBody(file.Body, nil, &hclFile)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode HCL: %s", diags.Error())
	}

	// Convert HCL rules to internal Rule format
	rules := make([]Rule, 0, len(hclFile.Rules))
	seenIDs := make(map[string]bool)

	for i, hclRule := range hclFile.Rules {
		rule, err := convertHCLRule(&hclRule, i)
		if err != nil {
			return nil, err
		}

		// Check for duplicate IDs
		if seenIDs[rule.ID] {
			return nil, fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		seenIDs[rule.ID] = true

		rules = append(rules, *rule)
	}

	return rules, nil
}

// convertHCLRule converts an HCL rule to the internal Rule format
func convertHCLRule(hclRule *HCLRule, index int) (*Rule, error) {
	// Validate required fields
	if hclRule.ID == "" {
		return nil, fmt.Errorf("rule at index %d: id is required", index)
	}

	if hclRule.Name == "" {
		return nil, fmt.Errorf("rule %s: name is required", hclRule.ID)
	}

	if hclRule.SeverityStr == "" {
		return nil, fmt.Errorf("rule %s: severity is required", hclRule.ID)
	}

	// Validate severity value
	switch hclRule.SeverityStr {
	case "error", "warning", "warn", "info":
		// Valid
	default:
		return nil, fmt.Errorf("rule %s: invalid severity '%s' (must be error, warning, or info)", hclRule.ID, hclRule.SeverityStr)
	}

	if hclRule.ResourceType == "" {
		return nil, fmt.Errorf("rule %s: resource_type is required", hclRule.ID)
	}

	if hclRule.Condition == nil {
		return nil, fmt.Errorf("rule %s: condition block is required", hclRule.ID)
	}

	if hclRule.Message == "" {
		return nil, fmt.Errorf("rule %s: message is required", hclRule.ID)
	}

	// Convert condition
	condition, err := convertHCLCondition(hclRule.Condition)
	if err != nil {
		return nil, fmt.Errorf("rule %s: invalid condition: %w", hclRule.ID, err)
	}

	// Validate condition
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("rule %s: invalid condition: %w", hclRule.ID, err)
	}

	return &Rule{
		ID:           hclRule.ID,
		Name:         hclRule.Name,
		Description:  hclRule.Description,
		Severity:     ParseSeverity(hclRule.SeverityStr),
		SeverityStr:  hclRule.SeverityStr,
		ResourceType: hclRule.ResourceType,
		Condition:    *condition,
		Message:      hclRule.Message,
		Remediation:  hclRule.Remediation,
		Tags:         hclRule.Tags,
	}, nil
}

// convertHCLCondition converts an HCL condition to the internal Condition format
func convertHCLCondition(hc *HCLCondition) (*Condition, error) {
	if hc == nil {
		return nil, fmt.Errorf("condition is nil")
	}

	// Check which type of condition we have
	count := 0
	if hc.Check != nil {
		count++
	}
	if hc.All != nil {
		count++
	}
	if hc.Any != nil {
		count++
	}
	if hc.Not != nil {
		count++
	}

	if count == 0 {
		return nil, fmt.Errorf("condition must have exactly one of: check, all, any, or not block")
	}
	if count > 1 {
		return nil, fmt.Errorf("condition must have exactly one of: check, all, any, or not block (found %d)", count)
	}

	// Handle check block
	if hc.Check != nil {
		return convertHCLCheck(hc.Check)
	}

	// Handle all block
	if hc.All != nil {
		conditions, err := convertHCLConditions(hc.All)
		if err != nil {
			return nil, err
		}
		return &Condition{
			Operator:   OpAll,
			Conditions: conditions,
		}, nil
	}

	// Handle any block
	if hc.Any != nil {
		conditions, err := convertHCLConditions(hc.Any)
		if err != nil {
			return nil, err
		}
		return &Condition{
			Operator:   OpAny,
			Conditions: conditions,
		}, nil
	}

	// Handle not block
	if hc.Not != nil {
		nestedCondition, err := convertHCLCondition(hc.Not)
		if err != nil {
			return nil, err
		}
		return &Condition{
			Operator:   OpNot,
			Conditions: []Condition{*nestedCondition},
		}, nil
	}

	return nil, fmt.Errorf("condition must have exactly one of: check, all, any, or not block")
}

// convertHCLConditions converts an HCL conditions block to a slice of Conditions
func convertHCLConditions(hc *HCLConditions) ([]Condition, error) {
	if hc == nil {
		return nil, fmt.Errorf("conditions is nil")
	}

	var conditions []Condition

	// Convert check blocks
	for _, check := range hc.Checks {
		cond, err := convertHCLCheck(check)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, *cond)
	}

	// Convert nested all block
	if hc.All != nil {
		nestedConditions, err := convertHCLConditions(hc.All)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, Condition{
			Operator:   OpAll,
			Conditions: nestedConditions,
		})
	}

	// Convert nested any block
	if hc.Any != nil {
		nestedConditions, err := convertHCLConditions(hc.Any)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, Condition{
			Operator:   OpAny,
			Conditions: nestedConditions,
		})
	}

	// Convert nested not block
	if hc.Not != nil {
		nestedCondition, err := convertHCLCondition(hc.Not)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, Condition{
			Operator:   OpNot,
			Conditions: []Condition{*nestedCondition},
		})
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("conditions block must contain at least one check, all, any, or not block")
	}

	return conditions, nil
}

// convertHCLCheck converts an HCL check block to a Condition
func convertHCLCheck(check *HCLCheck) (*Condition, error) {
	if check == nil {
		return nil, fmt.Errorf("check is nil")
	}

	if check.Attribute == "" {
		return nil, fmt.Errorf("check: attribute is required")
	}

	if check.Operator == "" {
		return nil, fmt.Errorf("check: operator is required")
	}

	// Validate operator
	op := Operator(check.Operator)
	switch op {
	case OpExists, OpNotExists, OpEquals, OpNotEquals, OpContains, OpNotContains, OpMatches, OpGreaterThan, OpLessThan, OpOneOf:
		// Valid
	default:
		return nil, fmt.Errorf("check: unknown operator: %s", check.Operator)
	}

	// Convert cty.Value to interface{}
	var value interface{}
	if !check.Value.IsNull() {
		value = ctyValueToInterface(check.Value)
	}

	return &Condition{
		Attribute: check.Attribute,
		Operator:  op,
		Value:     value,
	}, nil
}

// ctyValueToInterface converts a cty.Value to a Go interface{}
func ctyValueToInterface(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	switch {
	case val.Type() == cty.String:
		return val.AsString()
	case val.Type() == cty.Bool:
		return val.True()
	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		// Try to return as int if it's a whole number
		if bf.IsInt() {
			i64, _ := bf.Int64()
			return int(i64)
		}
		f64, _ := bf.Float64()
		return f64
	case val.Type().IsListType() || val.Type().IsTupleType():
		var result []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			result = append(result, ctyValueToInterface(v))
		}
		return result
	case val.Type().IsMapType() || val.Type().IsObjectType():
		result := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			result[k.AsString()] = ctyValueToInterface(v)
		}
		return result
	default:
		// For unknown types, return the string representation
		return val.GoString()
	}
}

// LoadFromFile loads rules from either YAML or HCL file based on extension
func LoadFromFile(path string) ([]Rule, error) {
	if strings.HasSuffix(path, ".hcl") || strings.HasSuffix(path, ".spec.hcl") {
		return LoadFromHCLFile(path)
	}
	// Fall back to YAML for backwards compatibility
	return LoadFromYAMLFile(path)
}

// LoadFromYAMLFile loads rules from a YAML file (for backwards compatibility)
func LoadFromYAMLFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromYAMLBytes(data)
}
