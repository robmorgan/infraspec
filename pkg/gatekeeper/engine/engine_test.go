package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/pkg/gatekeeper/parser"
	"github.com/robmorgan/infraspec/pkg/gatekeeper/rules"
)

func TestEngine_Evaluate_BasicOperators(t *testing.T) {
	e := New(Config{})

	resources := []parser.Resource{
		{
			Type: "aws_s3_bucket",
			Name: "test_bucket",
			Attributes: map[string]interface{}{
				"bucket":      "my-bucket",
				"acl":         "private",
				"versioning":  map[string]interface{}{"enabled": true},
				"tags":        map[string]interface{}{"env": "prod"},
				"replicas":    3,
				"enabled":     true,
				"cidr_blocks": []interface{}{"10.0.0.0/8", "192.168.0.0/16"},
			},
			Location: parser.Location{File: "main.tf", Line: 1},
		},
	}

	tests := []struct {
		name        string
		rule        rules.Rule
		expectViols int
		description string
	}{
		{
			name: "exists - pass",
			rule: rules.Rule{
				ID:           "TEST_001",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "bucket", Operator: rules.OpExists},
				Message:      "test",
			},
			expectViols: 0,
			description: "bucket exists",
		},
		{
			name: "exists - fail",
			rule: rules.Rule{
				ID:           "TEST_002",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "encryption", Operator: rules.OpExists},
				Message:      "test",
			},
			expectViols: 1,
			description: "encryption does not exist",
		},
		{
			name: "not_exists - pass",
			rule: rules.Rule{
				ID:           "TEST_003",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "encryption", Operator: rules.OpNotExists},
				Message:      "test",
			},
			expectViols: 0,
			description: "encryption does not exist",
		},
		{
			name: "equals - pass",
			rule: rules.Rule{
				ID:           "TEST_004",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "acl", Operator: rules.OpEquals, Value: "private"},
				Message:      "test",
			},
			expectViols: 0,
			description: "acl equals private",
		},
		{
			name: "equals - fail",
			rule: rules.Rule{
				ID:           "TEST_005",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "acl", Operator: rules.OpEquals, Value: "public"},
				Message:      "test",
			},
			expectViols: 1,
			description: "acl does not equal public",
		},
		{
			name: "not_equals - pass",
			rule: rules.Rule{
				ID:           "TEST_006",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "acl", Operator: rules.OpNotEquals, Value: "public"},
				Message:      "test",
			},
			expectViols: 0,
			description: "acl not equals public",
		},
		{
			name: "contains - pass (array)",
			rule: rules.Rule{
				ID:           "TEST_007",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "cidr_blocks", Operator: rules.OpContains, Value: "10.0.0.0/8"},
				Message:      "test",
			},
			expectViols: 0,
			description: "cidr_blocks contains 10.0.0.0/8",
		},
		{
			name: "contains - pass (string)",
			rule: rules.Rule{
				ID:           "TEST_008",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "bucket", Operator: rules.OpContains, Value: "bucket"},
				Message:      "test",
			},
			expectViols: 0,
			description: "bucket contains 'bucket'",
		},
		{
			name: "not_contains - pass",
			rule: rules.Rule{
				ID:           "TEST_009",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "cidr_blocks", Operator: rules.OpNotContains, Value: "0.0.0.0/0"},
				Message:      "test",
			},
			expectViols: 0,
			description: "cidr_blocks does not contain 0.0.0.0/0",
		},
		{
			name: "matches - pass",
			rule: rules.Rule{
				ID:           "TEST_010",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "bucket", Operator: rules.OpMatches, Value: "^my-.*"},
				Message:      "test",
			},
			expectViols: 0,
			description: "bucket matches ^my-.*",
		},
		{
			name: "greater_than - pass",
			rule: rules.Rule{
				ID:           "TEST_011",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "replicas", Operator: rules.OpGreaterThan, Value: 2},
				Message:      "test",
			},
			expectViols: 0,
			description: "replicas > 2",
		},
		{
			name: "less_than - pass",
			rule: rules.Rule{
				ID:           "TEST_012",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "replicas", Operator: rules.OpLessThan, Value: 5},
				Message:      "test",
			},
			expectViols: 0,
			description: "replicas < 5",
		},
		{
			name: "one_of - pass",
			rule: rules.Rule{
				ID:           "TEST_013",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "acl", Operator: rules.OpOneOf, Value: []interface{}{"private", "authenticated-read"}},
				Message:      "test",
			},
			expectViols: 0,
			description: "acl is one of [private, authenticated-read]",
		},
		{
			name: "nested attribute - pass",
			rule: rules.Rule{
				ID:           "TEST_014",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    rules.Condition{Attribute: "versioning.enabled", Operator: rules.OpEquals, Value: true},
				Message:      "test",
			},
			expectViols: 0,
			description: "versioning.enabled equals true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := e.Evaluate([]rules.Rule{tt.rule}, resources)
			assert.Len(t, violations, tt.expectViols, tt.description)
		})
	}
}

func TestEngine_Evaluate_LogicalOperators(t *testing.T) {
	e := New(Config{})

	resources := []parser.Resource{
		{
			Type: "aws_s3_bucket",
			Name: "test_bucket",
			Attributes: map[string]interface{}{
				"bucket": "my-bucket",
				"acl":    "private",
			},
			Location: parser.Location{File: "main.tf", Line: 1},
		},
	}

	tests := []struct {
		name        string
		condition   rules.Condition
		expectViols int
	}{
		{
			name: "all - pass",
			condition: rules.Condition{
				Operator: rules.OpAll,
				Conditions: []rules.Condition{
					{Attribute: "bucket", Operator: rules.OpExists},
					{Attribute: "acl", Operator: rules.OpEquals, Value: "private"},
				},
			},
			expectViols: 0,
		},
		{
			name: "all - fail (one fails)",
			condition: rules.Condition{
				Operator: rules.OpAll,
				Conditions: []rules.Condition{
					{Attribute: "bucket", Operator: rules.OpExists},
					{Attribute: "acl", Operator: rules.OpEquals, Value: "public"},
				},
			},
			expectViols: 1,
		},
		{
			name: "any - pass",
			condition: rules.Condition{
				Operator: rules.OpAny,
				Conditions: []rules.Condition{
					{Attribute: "missing", Operator: rules.OpExists},
					{Attribute: "bucket", Operator: rules.OpExists},
				},
			},
			expectViols: 0,
		},
		{
			name: "any - fail (all fail)",
			condition: rules.Condition{
				Operator: rules.OpAny,
				Conditions: []rules.Condition{
					{Attribute: "missing1", Operator: rules.OpExists},
					{Attribute: "missing2", Operator: rules.OpExists},
				},
			},
			expectViols: 1,
		},
		{
			name: "not - pass",
			condition: rules.Condition{
				Operator: rules.OpNot,
				Conditions: []rules.Condition{
					{Attribute: "acl", Operator: rules.OpEquals, Value: "public"},
				},
			},
			expectViols: 0,
		},
		{
			name: "not - fail",
			condition: rules.Condition{
				Operator: rules.OpNot,
				Conditions: []rules.Condition{
					{Attribute: "acl", Operator: rules.OpEquals, Value: "private"},
				},
			},
			expectViols: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.Rule{
				ID:           "TEST",
				Name:         "Test",
				Severity:     rules.SeverityError,
				ResourceType: "aws_s3_bucket",
				Condition:    tt.condition,
				Message:      "test",
			}
			violations := e.Evaluate([]rules.Rule{rule}, resources)
			assert.Len(t, violations, tt.expectViols)
		})
	}
}

func TestEngine_Evaluate_ResourceTypeMatching(t *testing.T) {
	e := New(Config{})

	resources := []parser.Resource{
		{Type: "aws_s3_bucket", Name: "bucket1", Attributes: map[string]interface{}{}},
		{Type: "aws_security_group", Name: "sg1", Attributes: map[string]interface{}{}},
		{Type: "aws_s3_bucket", Name: "bucket2", Attributes: map[string]interface{}{}},
	}

	rule := rules.Rule{
		ID:           "S3_001",
		Name:         "S3 Test",
		Severity:     rules.SeverityError,
		ResourceType: "aws_s3_bucket",
		Condition:    rules.Condition{Attribute: "encryption", Operator: rules.OpExists},
		Message:      "test",
	}

	violations := e.Evaluate([]rules.Rule{rule}, resources)

	// Should only match the 2 S3 buckets, not the security group
	assert.Len(t, violations, 2)
	for _, v := range violations {
		assert.Equal(t, "aws_s3_bucket", v.ResourceType)
	}
}

func TestEngine_Evaluate_UnknownValues(t *testing.T) {
	resources := []parser.Resource{
		{
			Type: "aws_s3_bucket",
			Name: "test",
			Attributes: map[string]interface{}{
				"bucket":  parser.UnknownValue{},
				"known":   "value",
				"dynamic": parser.ComputedValue{},
			},
		},
	}

	rule := rules.Rule{
		ID:           "TEST",
		Name:         "Test",
		Severity:     rules.SeverityError,
		ResourceType: "aws_s3_bucket",
		Condition:    rules.Condition{Attribute: "bucket", Operator: rules.OpEquals, Value: "expected"},
		Message:      "test",
	}

	// Non-strict mode: unknown values don't cause violations
	t.Run("non-strict mode", func(t *testing.T) {
		e := New(Config{StrictUnknowns: false})
		violations := e.Evaluate([]rules.Rule{rule}, resources)
		assert.Len(t, violations, 0, "unknown values should not cause violations in non-strict mode")
	})

	// Strict mode: unknown values cause violations
	t.Run("strict mode", func(t *testing.T) {
		e := New(Config{StrictUnknowns: true})
		violations := e.Evaluate([]rules.Rule{rule}, resources)
		assert.Len(t, violations, 1, "unknown values should cause violations in strict mode")
	})
}

func TestEngine_Evaluate_MessageTemplates(t *testing.T) {
	e := New(Config{})

	resources := []parser.Resource{
		{
			Type:       "aws_s3_bucket",
			Name:       "my_bucket",
			Attributes: map[string]interface{}{},
			Location:   parser.Location{File: "main.tf", Line: 42},
		},
	}

	rule := rules.Rule{
		ID:           "TEST",
		Name:         "Test",
		Severity:     rules.SeverityError,
		ResourceType: "aws_s3_bucket",
		Condition:    rules.Condition{Attribute: "encryption", Operator: rules.OpExists},
		Message:      "Bucket '{{.resource_name}}' at {{.file}}:{{.line}} missing encryption",
	}

	violations := e.Evaluate([]rules.Rule{rule}, resources)
	require.Len(t, violations, 1)

	assert.Equal(t, "Bucket 'my_bucket' at main.tf:42 missing encryption", violations[0].Message)
}

func TestEngine_Evaluate_ViolationDetails(t *testing.T) {
	e := New(Config{})

	resources := []parser.Resource{
		{
			Type:       "aws_s3_bucket",
			Name:       "test_bucket",
			Attributes: map[string]interface{}{},
			Location:   parser.Location{File: "/path/to/main.tf", Line: 10},
		},
	}

	rule := rules.Rule{
		ID:           "S3_001",
		Name:         "S3 Encryption Required",
		Severity:     rules.SeverityError,
		ResourceType: "aws_s3_bucket",
		Condition:    rules.Condition{Attribute: "encryption", Operator: rules.OpExists},
		Message:      "Missing encryption",
		Remediation:  "Add encryption block",
	}

	violations := e.Evaluate([]rules.Rule{rule}, resources)
	require.Len(t, violations, 1)

	v := violations[0]
	assert.Equal(t, "S3_001", v.RuleID)
	assert.Equal(t, "S3 Encryption Required", v.RuleName)
	assert.Equal(t, rules.SeverityError, v.Severity)
	assert.Equal(t, "aws_s3_bucket", v.ResourceType)
	assert.Equal(t, "test_bucket", v.ResourceName)
	assert.Equal(t, "/path/to/main.tf", v.File)
	assert.Equal(t, 10, v.Line)
	assert.Equal(t, "Missing encryption", v.Message)
	assert.Equal(t, "Add encryption block", v.Remediation)
}

func TestEngine_Equals(t *testing.T) {
	e := New(Config{})

	tests := []struct {
		a, b     interface{}
		expected bool
	}{
		// Same types
		{"hello", "hello", true},
		{"hello", "world", false},
		{42, 42, true},
		{42, 43, false},
		{3.14, 3.14, true},
		{true, true, true},
		{true, false, false},

		// Cross-type numeric
		{42, 42.0, true},
		{int64(42), 42, true},
		{42.0, int64(42), true},

		// Bool to string
		{true, "true", true},
		{false, "false", true},

		// Nil
		{nil, nil, true},
		{nil, "hello", false},

		// Arrays
		{[]interface{}{1, 2, 3}, []interface{}{1, 2, 3}, true},
		{[]interface{}{1, 2, 3}, []interface{}{1, 2}, false},
		{[]interface{}{1, 2, 3}, []interface{}{1, 3, 2}, false},

		// Maps
		{map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1}, true},
		{map[string]interface{}{"a": 1}, map[string]interface{}{"a": 2}, false},
		{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 1}, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := e.equals(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "equals(%v, %v)", tt.a, tt.b)
		})
	}
}

func TestEngine_Contains(t *testing.T) {
	e := New(Config{})

	tests := []struct {
		a, b     interface{}
		expected bool
	}{
		// String contains
		{"hello world", "world", true},
		{"hello world", "foo", false},

		// Array contains
		{[]interface{}{"a", "b", "c"}, "b", true},
		{[]interface{}{"a", "b", "c"}, "d", false},
		{[]interface{}{1, 2, 3}, 2, true},
		{[]interface{}{1, 2, 3}, 4, false},

		// Map contains key
		{map[string]interface{}{"a": 1, "b": 2}, "a", true},
		{map[string]interface{}{"a": 1, "b": 2}, "c", false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := e.contains(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "contains(%v, %v)", tt.a, tt.b)
		})
	}
}

func TestEngine_Matches(t *testing.T) {
	e := New(Config{})

	tests := []struct {
		a, b     interface{}
		expected bool
	}{
		{"hello123", "^hello\\d+$", true},
		{"hello", "^hello\\d+$", false},
		{"test@example.com", "^[a-z]+@[a-z]+\\.[a-z]+$", true},
		{123, "\\d+", true}, // Non-string converted to string
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := e.matches(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "matches(%v, %v)", tt.a, tt.b)
		})
	}
}

func TestEngine_NumericComparisons(t *testing.T) {
	e := New(Config{})

	t.Run("greaterThan", func(t *testing.T) {
		assert.True(t, e.greaterThan(5, 3))
		assert.False(t, e.greaterThan(3, 5))
		assert.False(t, e.greaterThan(3, 3))
		assert.True(t, e.greaterThan(5.5, 5.0))
		assert.True(t, e.greaterThan(int64(10), 5))
	})

	t.Run("lessThan", func(t *testing.T) {
		assert.True(t, e.lessThan(3, 5))
		assert.False(t, e.lessThan(5, 3))
		assert.False(t, e.lessThan(3, 3))
		assert.True(t, e.lessThan(4.5, 5.0))
		assert.True(t, e.lessThan(3, int64(10)))
	})
}

func TestEngine_OneOf(t *testing.T) {
	e := New(Config{})

	tests := []struct {
		a, b     interface{}
		expected bool
	}{
		{"b", []interface{}{"a", "b", "c"}, true},
		{"d", []interface{}{"a", "b", "c"}, false},
		{2, []interface{}{1, 2, 3}, true},
		{4, []interface{}{1, 2, 3}, false},
		// Single value (not array)
		{"a", "a", true},
		{"a", "b", false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := e.oneOf(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "oneOf(%v, %v)", tt.a, tt.b)
		})
	}
}
