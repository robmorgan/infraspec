// Package config provides configuration loading for InfraSpec Gatekeeper.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

// InfraspecConfig represents the complete configuration from .infraspec.hcl
type InfraspecConfig struct {
	Config *ConfigBlock    `hcl:"config,block"`
	Rules  []HCLRuleConfig `hcl:"rule,block"`
	Remain hcl.Body        `hcl:",remain"`
}

// ConfigBlock represents the config block in .infraspec.hcl
type ConfigBlock struct {
	MinSeverity string `hcl:"min_severity,optional"`
	Format      string `hcl:"format,optional"`
	Strict      bool   `hcl:"strict,optional"`
	NoBuiltin   bool   `hcl:"no_builtin,optional"`
}

// HCLRuleConfig represents a rule block in the config file
type HCLRuleConfig struct {
	ID           string         `hcl:",label"`
	Name         string         `hcl:"name,attr"`
	Description  string         `hcl:"description,optional"`
	SeverityStr  string         `hcl:"severity,attr"`
	ResourceType string         `hcl:"resource_type,attr"`
	Condition    *HCLCondConfig `hcl:"condition,block"`
	Message      string         `hcl:"message,attr"`
	Remediation  string         `hcl:"remediation,optional"`
	Tags         []string       `hcl:"tags,optional"`
}

// HCLCondConfig represents a condition block in the config file
type HCLCondConfig struct {
	Check *HCLCheckConfig `hcl:"check,block"`
	All   *HCLCondsConfig `hcl:"all,block"`
	Any   *HCLCondsConfig `hcl:"any,block"`
	Not   *HCLCondConfig  `hcl:"not,block"`
}

// HCLCondsConfig represents a logical grouping of conditions
type HCLCondsConfig struct {
	Checks []*HCLCheckConfig `hcl:"check,block"`
	All    *HCLCondsConfig   `hcl:"all,block"`
	Any    *HCLCondsConfig   `hcl:"any,block"`
	Not    *HCLCondConfig    `hcl:"not,block"`
}

// HCLCheckConfig represents a check block in the config file
type HCLCheckConfig struct {
	Attribute string    `hcl:"attribute,attr"`
	Operator  string    `hcl:"operator,attr"`
	Value     cty.Value `hcl:"value,optional"`
}

// LoadedConfig represents the processed configuration with parsed rules
type LoadedConfig struct {
	// Configuration settings
	MinSeverity string
	Format      string
	Strict      bool
	NoBuiltin   bool

	// Rules defined in the config file
	Rules []rules.Rule

	// Path to the config file (empty if not found)
	FilePath string
}

// DefaultConfig returns a LoadedConfig with default values
func DefaultConfig() *LoadedConfig {
	return &LoadedConfig{
		MinSeverity: "error",
		Format:      "text",
		Strict:      false,
		NoBuiltin:   false,
		Rules:       nil,
		FilePath:    "",
	}
}

// FindConfigFile searches for .infraspec.hcl starting from the given path
// and walking up the directory tree until it finds one or reaches the root.
func FindConfigFile(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// If startPath is a file, start from its directory
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}
	if !info.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	// Walk up the directory tree
	for {
		configPath := filepath.Join(absPath, ".infraspec.hcl")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(absPath)
		if parent == absPath {
			// Reached root, no config found
			return "", nil
		}
		absPath = parent
	}
}

// LoadConfigFile loads and parses an .infraspec.hcl file
func LoadConfigFile(path string) (*LoadedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return LoadConfigBytes(data, path)
}

// LoadConfigBytes parses .infraspec.hcl configuration from bytes
func LoadConfigBytes(data []byte, filename string) (*LoadedConfig, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse config HCL: %s", diags.Error())
	}

	var hclConfig InfraspecConfig
	diags = gohcl.DecodeBody(file.Body, nil, &hclConfig)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode config: %s", diags.Error())
	}

	// Start with defaults
	config := DefaultConfig()
	config.FilePath = filename

	// Apply config block settings if present
	if hclConfig.Config != nil {
		if hclConfig.Config.MinSeverity != "" {
			config.MinSeverity = hclConfig.Config.MinSeverity
		}
		if hclConfig.Config.Format != "" {
			config.Format = hclConfig.Config.Format
		}
		config.Strict = hclConfig.Config.Strict
		config.NoBuiltin = hclConfig.Config.NoBuiltin
	}

	// Convert rules from HCL format to internal format
	seenIDs := make(map[string]bool)
	for i, hclRule := range hclConfig.Rules {
		rule, err := convertHCLRuleToRule(&hclRule, i)
		if err != nil {
			return nil, fmt.Errorf("config file %s: %w", filename, err)
		}

		// Check for duplicate IDs
		if seenIDs[rule.ID] {
			return nil, fmt.Errorf("config file %s: duplicate rule ID: %s", filename, rule.ID)
		}
		seenIDs[rule.ID] = true

		config.Rules = append(config.Rules, *rule)
	}

	return config, nil
}

// convertHCLRuleToRule converts an HCL rule config to internal Rule format
func convertHCLRuleToRule(hclRule *HCLRuleConfig, index int) (*rules.Rule, error) {
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
	condition, err := convertCondition(hclRule.Condition)
	if err != nil {
		return nil, fmt.Errorf("rule %s: invalid condition: %w", hclRule.ID, err)
	}

	// Validate condition
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("rule %s: invalid condition: %w", hclRule.ID, err)
	}

	return &rules.Rule{
		ID:           hclRule.ID,
		Name:         hclRule.Name,
		Description:  hclRule.Description,
		Severity:     rules.ParseSeverity(hclRule.SeverityStr),
		SeverityStr:  hclRule.SeverityStr,
		ResourceType: hclRule.ResourceType,
		Condition:    *condition,
		Message:      hclRule.Message,
		Remediation:  hclRule.Remediation,
		Tags:         hclRule.Tags,
	}, nil
}

// convertCondition converts an HCL condition to the internal format
func convertCondition(hc *HCLCondConfig) (*rules.Condition, error) {
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
		return convertCheck(hc.Check)
	}

	// Handle all block
	if hc.All != nil {
		conditions, err := convertConditions(hc.All)
		if err != nil {
			return nil, err
		}
		return &rules.Condition{
			Operator:   rules.OpAll,
			Conditions: conditions,
		}, nil
	}

	// Handle any block
	if hc.Any != nil {
		conditions, err := convertConditions(hc.Any)
		if err != nil {
			return nil, err
		}
		return &rules.Condition{
			Operator:   rules.OpAny,
			Conditions: conditions,
		}, nil
	}

	// Handle not block
	if hc.Not != nil {
		nestedCondition, err := convertCondition(hc.Not)
		if err != nil {
			return nil, err
		}
		return &rules.Condition{
			Operator:   rules.OpNot,
			Conditions: []rules.Condition{*nestedCondition},
		}, nil
	}

	return nil, fmt.Errorf("condition must have exactly one of: check, all, any, or not block")
}

// convertConditions converts an HCL conditions block to a slice of Conditions
func convertConditions(hc *HCLCondsConfig) ([]rules.Condition, error) {
	if hc == nil {
		return nil, fmt.Errorf("conditions is nil")
	}

	var conditions []rules.Condition

	// Convert check blocks
	for _, check := range hc.Checks {
		cond, err := convertCheck(check)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, *cond)
	}

	// Convert nested all block
	if hc.All != nil {
		nestedConditions, err := convertConditions(hc.All)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, rules.Condition{
			Operator:   rules.OpAll,
			Conditions: nestedConditions,
		})
	}

	// Convert nested any block
	if hc.Any != nil {
		nestedConditions, err := convertConditions(hc.Any)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, rules.Condition{
			Operator:   rules.OpAny,
			Conditions: nestedConditions,
		})
	}

	// Convert nested not block
	if hc.Not != nil {
		nestedCondition, err := convertCondition(hc.Not)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, rules.Condition{
			Operator:   rules.OpNot,
			Conditions: []rules.Condition{*nestedCondition},
		})
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("conditions block must contain at least one check, all, any, or not block")
	}

	return conditions, nil
}

// convertCheck converts an HCL check block to a Condition
func convertCheck(check *HCLCheckConfig) (*rules.Condition, error) {
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
	op := rules.Operator(check.Operator)
	switch op {
	case rules.OpExists, rules.OpNotExists, rules.OpEquals, rules.OpNotEquals,
		rules.OpContains, rules.OpNotContains, rules.OpMatches,
		rules.OpGreaterThan, rules.OpLessThan, rules.OpOneOf:
		// Valid
	default:
		return nil, fmt.Errorf("check: unknown operator: %s", check.Operator)
	}

	// Convert cty.Value to interface{}
	var value interface{}
	if !check.Value.IsNull() {
		value = ctyValueToInterface(check.Value)
	}

	return &rules.Condition{
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
