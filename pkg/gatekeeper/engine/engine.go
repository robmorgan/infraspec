// Package engine provides the rule evaluation engine for the gatekeeper.
package engine

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/parser"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

// Config holds the engine configuration
type Config struct {
	StrictUnknowns bool // Treat unknown values as violations
}

// Engine evaluates rules against resources
type Engine struct {
	config Config
}

// New creates a new Engine
func New(cfg Config) *Engine {
	return &Engine{config: cfg}
}

// Violation represents a rule violation
type Violation struct {
	RuleID       string
	RuleName     string
	Severity     rules.Severity
	ResourceType string
	ResourceName string
	File         string
	Line         int
	Message      string
	Remediation  string
}

// Evaluate evaluates all rules against all resources
func (e *Engine) Evaluate(ruleList []rules.Rule, resources []parser.Resource) []Violation {
	var violations []Violation

	for _, resource := range resources {
		for _, rule := range ruleList {
			// Check if rule applies to this resource type
			if rule.ResourceType != resource.Type {
				continue
			}

			// Evaluate the rule condition
			result := e.evaluateCondition(rule.Condition, resource.Attributes)

			// Determine if this is a violation
			isViolation := false

			if !result.passed {
				// Condition failed - this is a violation
				isViolation = true
			} else if result.unknown && e.config.StrictUnknowns {
				// In strict mode, unknown values are treated as violations
				isViolation = true
			}

			if isViolation {
				// Skip if result is unknown and we're not in strict mode
				if result.unknown && !e.config.StrictUnknowns {
					continue
				}

				// Render message template
				message := e.renderMessage(rule.Message, resource)

				violations = append(violations, Violation{
					RuleID:       rule.ID,
					RuleName:     rule.Name,
					Severity:     rule.Severity,
					ResourceType: resource.Type,
					ResourceName: resource.Name,
					File:         resource.Location.File,
					Line:         resource.Location.Line,
					Message:      message,
					Remediation:  rule.Remediation,
				})
			}
		}
	}

	return violations
}

// conditionResult holds the result of condition evaluation
type conditionResult struct {
	passed  bool
	unknown bool // true if the result depends on unknown values
}

// evaluateCondition evaluates a condition against resource attributes
func (e *Engine) evaluateCondition(cond rules.Condition, attrs map[string]interface{}) conditionResult {
	// Handle logical operators
	switch cond.Operator {
	case rules.OpAll:
		return e.evaluateAll(cond.Conditions, attrs)
	case rules.OpAny:
		return e.evaluateAny(cond.Conditions, attrs)
	case rules.OpNot:
		if len(cond.Conditions) > 0 {
			result := e.evaluateCondition(cond.Conditions[0], attrs)
			return conditionResult{passed: !result.passed, unknown: result.unknown}
		}
		return conditionResult{passed: true}
	}

	// Get the attribute value
	value, exists := parser.GetAttribute(attrs, cond.Attribute)

	// Handle unknown/computed values
	if parser.IsUnknown(value) || parser.IsComputed(value) {
		return conditionResult{passed: true, unknown: true}
	}

	// Evaluate the operator
	switch cond.Operator {
	case rules.OpExists:
		return conditionResult{passed: exists && value != nil}

	case rules.OpNotExists:
		return conditionResult{passed: !exists || value == nil}

	case rules.OpEquals:
		return conditionResult{passed: e.equals(value, cond.Value)}

	case rules.OpNotEquals:
		return conditionResult{passed: !e.equals(value, cond.Value)}

	case rules.OpContains:
		return conditionResult{passed: e.contains(value, cond.Value)}

	case rules.OpNotContains:
		return conditionResult{passed: !e.contains(value, cond.Value)}

	case rules.OpMatches:
		return conditionResult{passed: e.matches(value, cond.Value)}

	case rules.OpGreaterThan:
		return conditionResult{passed: e.greaterThan(value, cond.Value)}

	case rules.OpLessThan:
		return conditionResult{passed: e.lessThan(value, cond.Value)}

	case rules.OpOneOf:
		return conditionResult{passed: e.oneOf(value, cond.Value)}

	default:
		return conditionResult{passed: true}
	}
}

// evaluateAll evaluates an "all" condition (AND)
func (e *Engine) evaluateAll(conditions []rules.Condition, attrs map[string]interface{}) conditionResult {
	anyUnknown := false
	for _, cond := range conditions {
		result := e.evaluateCondition(cond, attrs)
		if !result.passed {
			return result
		}
		if result.unknown {
			anyUnknown = true
		}
	}
	return conditionResult{passed: true, unknown: anyUnknown}
}

// evaluateAny evaluates an "any" condition (OR)
func (e *Engine) evaluateAny(conditions []rules.Condition, attrs map[string]interface{}) conditionResult {
	anyUnknown := false
	for _, cond := range conditions {
		result := e.evaluateCondition(cond, attrs)
		if result.passed && !result.unknown {
			return conditionResult{passed: true}
		}
		if result.unknown {
			anyUnknown = true
		}
	}
	return conditionResult{passed: false, unknown: anyUnknown}
}

// equals compares two values for equality
func (e *Engine) equals(a, b interface{}) bool {
	// Handle nil
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle type conversions for comparison
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
		// Handle string "true"/"false"
		if bv, ok := b.(string); ok {
			return (av && bv == "true") || (!av && bv == "false")
		}
	case int:
		return e.numbersEqual(float64(av), b)
	case int64:
		return e.numbersEqual(float64(av), b)
	case float64:
		return e.numbersEqual(av, b)
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !e.equals(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !e.equals(v, bv[k]) {
				return false
			}
		}
		return true
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// numbersEqual compares numbers with type flexibility
func (e *Engine) numbersEqual(a float64, b interface{}) bool {
	switch bv := b.(type) {
	case int:
		return a == float64(bv)
	case int64:
		return a == float64(bv)
	case float64:
		return a == bv
	default:
		return false
	}
}

// contains checks if a contains b
func (e *Engine) contains(a, b interface{}) bool {
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return strings.Contains(av, bv)
		}
		return strings.Contains(av, fmt.Sprintf("%v", b))
	case []interface{}:
		for _, item := range av {
			if e.equals(item, b) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		// Check if map contains key
		if bv, ok := b.(string); ok {
			_, exists := av[bv]
			return exists
		}
	}
	return false
}

// matches checks if a matches regex b
func (e *Engine) matches(a, b interface{}) bool {
	aStr, ok := a.(string)
	if !ok {
		aStr = fmt.Sprintf("%v", a)
	}
	bStr, ok := b.(string)
	if !ok {
		return false
	}

	re, err := regexp.Compile(bStr)
	if err != nil {
		return false
	}
	return re.MatchString(aStr)
}

// greaterThan checks if a > b (numeric comparison)
func (e *Engine) greaterThan(a, b interface{}) bool {
	aNum := e.toFloat(a)
	bNum := e.toFloat(b)
	if aNum == nil || bNum == nil {
		return false
	}
	return *aNum > *bNum
}

// lessThan checks if a < b (numeric comparison)
func (e *Engine) lessThan(a, b interface{}) bool {
	aNum := e.toFloat(a)
	bNum := e.toFloat(b)
	if aNum == nil || bNum == nil {
		return false
	}
	return *aNum < *bNum
}

// toFloat converts a value to float64
func (e *Engine) toFloat(v interface{}) *float64 {
	switch n := v.(type) {
	case int:
		f := float64(n)
		return &f
	case int64:
		f := float64(n)
		return &f
	case float64:
		return &n
	default:
		return nil
	}
}

// oneOf checks if a is one of the values in b
func (e *Engine) oneOf(a, b interface{}) bool {
	bList, ok := b.([]interface{})
	if !ok {
		return e.equals(a, b)
	}
	for _, item := range bList {
		if e.equals(a, item) {
			return true
		}
	}
	return false
}

// renderMessage renders a message template with resource data
func (e *Engine) renderMessage(msg string, resource parser.Resource) string {
	tmpl, err := template.New("message").Parse(msg)
	if err != nil {
		return msg
	}

	data := map[string]interface{}{
		"resource_name": resource.Name,
		"resource_type": resource.Type,
		"file":          resource.Location.File,
		"line":          resource.Location.Line,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return msg
	}

	return buf.String()
}
