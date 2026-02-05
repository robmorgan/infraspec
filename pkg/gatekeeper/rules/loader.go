package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LoadFromYAMLBytes loads rules from YAML bytes
func LoadFromYAMLBytes(data []byte) ([]Rule, error) {
	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate and process rules
	rules := make([]Rule, 0, len(ruleSet.Rules))
	seenIDs := make(map[string]bool)

	for i, rule := range ruleSet.Rules {
		// Validate required fields
		if err := validateRule(&rule, i); err != nil {
			return nil, err
		}

		// Check for duplicate IDs
		if seenIDs[rule.ID] {
			return nil, fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		seenIDs[rule.ID] = true

		// Parse severity string to enum
		rule.Severity = ParseSeverity(rule.SeverityStr)

		rules = append(rules, rule)
	}

	return rules, nil
}

// LoadFromBytes loads rules from YAML bytes (for backwards compatibility)
// Deprecated: Use LoadFromYAMLBytes or LoadFromHCLBytes instead
func LoadFromBytes(data []byte) ([]Rule, error) {
	return LoadFromYAMLBytes(data)
}

// validateRule validates a single rule
func validateRule(rule *Rule, index int) error {
	if rule.ID == "" {
		return fmt.Errorf("rule at index %d: id is required", index)
	}

	if rule.Name == "" {
		return fmt.Errorf("rule %s: name is required", rule.ID)
	}

	if rule.SeverityStr == "" {
		return fmt.Errorf("rule %s: severity is required", rule.ID)
	}

	switch rule.SeverityStr {
	case "error", "warning", "warn", "info":
		// Valid
	default:
		return fmt.Errorf("rule %s: invalid severity '%s' (must be error, warning, or info)", rule.ID, rule.SeverityStr)
	}

	if rule.ResourceType == "" {
		return fmt.Errorf("rule %s: resource_type is required", rule.ID)
	}

	if err := rule.Condition.Validate(); err != nil {
		return fmt.Errorf("rule %s: invalid condition: %w", rule.ID, err)
	}

	if rule.Message == "" {
		return fmt.Errorf("rule %s: message is required", rule.ID)
	}

	return nil
}
