package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromHCLBytes_ValidRules(t *testing.T) {
	hcl := `
rule "TEST_001" {
  name          = "Test rule"
  description   = "A test rule"
  severity      = "error"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "encryption"
      operator  = "exists"
    }
  }

  message     = "Missing encryption"
  remediation = "Add encryption"
  tags        = ["test", "security"]
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
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

func TestLoadFromHCLBytes_AllOperators(t *testing.T) {
	hcl := `
rule "OP_EXISTS" {
  name          = "Exists"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}

rule "OP_NOT_EXISTS" {
  name          = "Not Exists"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "not_exists"
    }
  }
  message = "test"
}

rule "OP_EQUALS" {
  name          = "Equals"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "equals"
      value     = "bar"
    }
  }
  message = "test"
}

rule "OP_NOT_EQUALS" {
  name          = "Not Equals"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "not_equals"
      value     = "bar"
    }
  }
  message = "test"
}

rule "OP_CONTAINS" {
  name          = "Contains"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "contains"
      value     = "bar"
    }
  }
  message = "test"
}

rule "OP_NOT_CONTAINS" {
  name          = "Not Contains"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "not_contains"
      value     = "bar"
    }
  }
  message = "test"
}

rule "OP_MATCHES" {
  name          = "Matches"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "matches"
      value     = "^bar.*"
    }
  }
  message = "test"
}

rule "OP_GREATER_THAN" {
  name          = "Greater Than"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "greater_than"
      value     = 10
    }
  }
  message = "test"
}

rule "OP_LESS_THAN" {
  name          = "Less Than"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "less_than"
      value     = 10
    }
  }
  message = "test"
}

rule "OP_ONE_OF" {
  name          = "One Of"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "one_of"
      value     = ["a", "b", "c"]
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	assert.Len(t, rules, 10)
}

func TestLoadFromHCLBytes_LogicalOperators(t *testing.T) {
	hcl := `
rule "ALL_TEST" {
  name          = "All conditions"
  severity      = "error"
  resource_type = "test"
  condition {
    all {
      check {
        attribute = "foo"
        operator  = "exists"
      }
      check {
        attribute = "bar"
        operator  = "equals"
        value     = "baz"
      }
    }
  }
  message = "test"
}

rule "ANY_TEST" {
  name          = "Any condition"
  severity      = "error"
  resource_type = "test"
  condition {
    any {
      check {
        attribute = "foo"
        operator  = "exists"
      }
      check {
        attribute = "bar"
        operator  = "exists"
      }
    }
  }
  message = "test"
}

rule "NOT_TEST" {
  name          = "Not condition"
  severity      = "error"
  resource_type = "test"
  condition {
    not {
      check {
        attribute = "foo"
        operator  = "equals"
        value     = "bad"
      }
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
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

func TestLoadFromHCLBytes_NestedLogic(t *testing.T) {
	hcl := `
rule "NESTED_TEST" {
  name          = "Nested logic"
  severity      = "error"
  resource_type = "test"
  condition {
    any {
      check {
        attribute = "encryption"
        operator  = "exists"
      }
      all {
        check {
          attribute = "kms_key_id"
          operator  = "exists"
        }
        check {
          attribute = "sse_algorithm"
          operator  = "equals"
          value     = "aws:kms"
        }
      }
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	// Top level is ANY
	assert.Equal(t, OpAny, rules[0].Condition.Operator)
	assert.Len(t, rules[0].Condition.Conditions, 2)

	// First condition is a simple check
	assert.Equal(t, "encryption", rules[0].Condition.Conditions[0].Attribute)
	assert.Equal(t, OpExists, rules[0].Condition.Conditions[0].Operator)

	// Second condition is ALL with two checks
	assert.Equal(t, OpAll, rules[0].Condition.Conditions[1].Operator)
	assert.Len(t, rules[0].Condition.Conditions[1].Conditions, 2)
}

func TestLoadFromHCLBytes_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		hcl     string
		wantErr string
	}{
		{
			name: "missing name",
			hcl: `
rule "TEST" {
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}
`,
			wantErr: "\"name\" is required",
		},
		{
			name: "missing severity",
			hcl: `
rule "TEST" {
  name          = "Test"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}
`,
			wantErr: "\"severity\" is required",
		},
		{
			name: "invalid severity",
			hcl: `
rule "TEST" {
  name          = "Test"
  severity      = "critical"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}
`,
			wantErr: "invalid severity",
		},
		{
			name: "missing resource_type",
			hcl: `
rule "TEST" {
  name     = "Test"
  severity = "error"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}
`,
			wantErr: "\"resource_type\" is required",
		},
		{
			name: "missing message",
			hcl: `
rule "TEST" {
  name          = "Test"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
}
`,
			wantErr: "\"message\" is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFromHCLBytes([]byte(tt.hcl), "test.hcl")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadFromHCLBytes_DuplicateIDs(t *testing.T) {
	hcl := `
rule "DUPE_001" {
  name          = "First"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}

rule "DUPE_001" {
  name          = "Second"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "bar"
      operator  = "exists"
    }
  }
  message = "test"
}
`

	_, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate rule ID")
}

func TestLoadFromHCLBytes_SeverityParsing(t *testing.T) {
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
			hcl := `
rule "TEST" {
  name          = "Test"
  severity      = "` + tt.severity + `"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
}
`
			rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, rules[0].Severity)
		})
	}
}

func TestLoadFromHCLBytes_BooleanValue(t *testing.T) {
	hcl := `
rule "BOOL_TEST" {
  name          = "Boolean test"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "enabled"
      operator  = "equals"
      value     = true
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	// Value should be a boolean
	assert.Equal(t, true, rules[0].Condition.Value)
}

func TestLoadFromHCLBytes_NumericValue(t *testing.T) {
	hcl := `
rule "NUM_TEST" {
  name          = "Numeric test"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "count"
      operator  = "greater_than"
      value     = 100
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	// Value should be an integer
	assert.Equal(t, 100, rules[0].Condition.Value)
}

func TestLoadFromHCLBytes_ListValue(t *testing.T) {
	hcl := `
rule "LIST_TEST" {
  name          = "List test"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "type"
      operator  = "one_of"
      value     = ["t3.micro", "t3.small", "t3.medium"]
    }
  }
  message = "test"
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	// Value should be a list
	expected := []interface{}{"t3.micro", "t3.small", "t3.medium"}
	assert.Equal(t, expected, rules[0].Condition.Value)
}

func TestLoadFromHCLBytes_HeredocRemediation(t *testing.T) {
	hcl := `
rule "HEREDOC_TEST" {
  name          = "Heredoc test"
  severity      = "error"
  resource_type = "test"
  condition {
    check {
      attribute = "foo"
      operator  = "exists"
    }
  }
  message = "test"
  remediation = <<-EOT
    Multi-line remediation text.

    resource "example" {
      setting = "value"
    }
  EOT
}
`

	rules, err := LoadFromHCLBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	// Remediation should contain the heredoc content
	assert.Contains(t, rules[0].Remediation, "Multi-line remediation text")
	assert.Contains(t, rules[0].Remediation, "resource \"example\"")
}
