package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromBytes_ValidRules(t *testing.T) {
	yaml := `
version: "1"
metadata:
  name: "Test Rules"
  description: "Test rule set"

rules:
  - id: TEST_001
    name: "Test rule"
    description: "A test rule"
    severity: error
    resource_type: aws_s3_bucket
    condition:
      attribute: encryption
      operator: exists
    message: "Missing encryption"
    remediation: "Add encryption"
    tags:
      - test
      - security
`

	rules, err := LoadFromBytes([]byte(yaml))
	require.NoError(t, err)
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, "TEST_001", rule.ID)
	assert.Equal(t, "Test rule", rule.Name)
	assert.Equal(t, "A test rule", rule.Description)
	assert.Equal(t, SeverityError, rule.Severity)
	assert.Equal(t, "aws_s3_bucket", rule.ResourceType)
	assert.Equal(t, "encryption", rule.Condition.Attribute)
	assert.Equal(t, OpExists, rule.Condition.Operator)
	assert.Equal(t, "Missing encryption", rule.Message)
	assert.Equal(t, "Add encryption", rule.Remediation)
	assert.Equal(t, []string{"test", "security"}, rule.Tags)
}

func TestLoadFromBytes_AllOperators(t *testing.T) {
	yaml := `
version: "1"
rules:
  - id: OP_EXISTS
    name: "Exists"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"

  - id: OP_NOT_EXISTS
    name: "Not Exists"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: not_exists
    message: "test"

  - id: OP_EQUALS
    name: "Equals"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: equals
      value: "bar"
    message: "test"

  - id: OP_NOT_EQUALS
    name: "Not Equals"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: not_equals
      value: "bar"
    message: "test"

  - id: OP_CONTAINS
    name: "Contains"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: contains
      value: "bar"
    message: "test"

  - id: OP_NOT_CONTAINS
    name: "Not Contains"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: not_contains
      value: "bar"
    message: "test"

  - id: OP_MATCHES
    name: "Matches"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: matches
      value: "^bar.*"
    message: "test"

  - id: OP_GREATER_THAN
    name: "Greater Than"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: greater_than
      value: 10
    message: "test"

  - id: OP_LESS_THAN
    name: "Less Than"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: less_than
      value: 10
    message: "test"

  - id: OP_ONE_OF
    name: "One Of"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: one_of
      value:
        - "a"
        - "b"
        - "c"
    message: "test"
`

	rules, err := LoadFromBytes([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, rules, 10)
}

func TestLoadFromBytes_LogicalOperators(t *testing.T) {
	yaml := `
version: "1"
rules:
  - id: ALL_TEST
    name: "All conditions"
    severity: error
    resource_type: test
    condition:
      operator: all
      conditions:
        - attribute: foo
          operator: exists
        - attribute: bar
          operator: equals
          value: "baz"
    message: "test"

  - id: ANY_TEST
    name: "Any condition"
    severity: error
    resource_type: test
    condition:
      operator: any
      conditions:
        - attribute: foo
          operator: exists
        - attribute: bar
          operator: exists
    message: "test"

  - id: NOT_TEST
    name: "Not condition"
    severity: error
    resource_type: test
    condition:
      operator: not
      conditions:
        - attribute: foo
          operator: equals
          value: "bad"
    message: "test"
`

	rules, err := LoadFromBytes([]byte(yaml))
	require.NoError(t, err)
	require.Len(t, rules, 3)

	// Test ALL
	assert.Equal(t, OpAll, rules[0].Condition.Operator)
	assert.Len(t, rules[0].Condition.Conditions, 2)

	// Test ANY
	assert.Equal(t, OpAny, rules[1].Condition.Operator)
	assert.Len(t, rules[1].Condition.Conditions, 2)

	// Test NOT
	assert.Equal(t, OpNot, rules[2].Condition.Operator)
	assert.Len(t, rules[2].Condition.Conditions, 1)
}

func TestLoadFromBytes_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing id",
			yaml: `
version: "1"
rules:
  - name: "Test"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"
`,
			wantErr: "id is required",
		},
		{
			name: "missing name",
			yaml: `
version: "1"
rules:
  - id: TEST
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"
`,
			wantErr: "name is required",
		},
		{
			name: "missing severity",
			yaml: `
version: "1"
rules:
  - id: TEST
    name: "Test"
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"
`,
			wantErr: "severity is required",
		},
		{
			name: "invalid severity",
			yaml: `
version: "1"
rules:
  - id: TEST
    name: "Test"
    severity: critical
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"
`,
			wantErr: "invalid severity",
		},
		{
			name: "missing resource_type",
			yaml: `
version: "1"
rules:
  - id: TEST
    name: "Test"
    severity: error
    condition:
      attribute: foo
      operator: exists
    message: "test"
`,
			wantErr: "resource_type is required",
		},
		{
			name: "missing message",
			yaml: `
version: "1"
rules:
  - id: TEST
    name: "Test"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: exists
`,
			wantErr: "message is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFromBytes([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadFromBytes_DuplicateIDs(t *testing.T) {
	yaml := `
version: "1"
rules:
  - id: DUPE_001
    name: "First"
    severity: error
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"

  - id: DUPE_001
    name: "Second"
    severity: error
    resource_type: test
    condition:
      attribute: bar
      operator: exists
    message: "test"
`

	_, err := LoadFromBytes([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate rule ID")
}

func TestLoadFromBytes_SeverityParsing(t *testing.T) {
	tests := []struct {
		severity string
		expected Severity
	}{
		{"error", SeverityError},
		{"warning", SeverityWarning},
		{"warn", SeverityWarning},
		{"info", SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			yaml := `
version: "1"
rules:
  - id: TEST
    name: "Test"
    severity: ` + tt.severity + `
    resource_type: test
    condition:
      attribute: foo
      operator: exists
    message: "test"
`
			rules, err := LoadFromBytes([]byte(yaml))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, rules[0].Severity)
		})
	}
}

func TestCondition_Validate(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		wantErr   bool
	}{
		{
			name: "valid exists",
			condition: Condition{
				Attribute: "foo",
				Operator:  OpExists,
			},
			wantErr: false,
		},
		{
			name: "valid equals",
			condition: Condition{
				Attribute: "foo",
				Operator:  OpEquals,
				Value:     "bar",
			},
			wantErr: false,
		},
		{
			name: "valid all",
			condition: Condition{
				Operator: OpAll,
				Conditions: []Condition{
					{Attribute: "foo", Operator: OpExists},
				},
			},
			wantErr: false,
		},
		{
			name: "missing attribute",
			condition: Condition{
				Operator: OpExists,
			},
			wantErr: true,
		},
		{
			name: "missing operator",
			condition: Condition{
				Attribute: "foo",
			},
			wantErr: true,
		},
		{
			name: "missing value for equals",
			condition: Condition{
				Attribute: "foo",
				Operator:  OpEquals,
			},
			wantErr: true,
		},
		{
			name: "missing conditions for all",
			condition: Condition{
				Operator: OpAll,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
