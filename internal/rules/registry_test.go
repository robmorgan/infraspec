package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/plan"
)

// mockRule implements Rule interface for testing.
type mockRule struct {
	id           string
	description  string
	severity     Severity
	provider     string
	resourceType string
}

func (m *mockRule) ID() string           { return m.id }
func (m *mockRule) Description() string  { return m.description }
func (m *mockRule) Severity() Severity   { return m.severity }
func (m *mockRule) Provider() string     { return m.provider }
func (m *mockRule) ResourceType() string { return m.resourceType }

func (m *mockRule) Check(_ *plan.ResourceChange) (*Result, error) {
	return &Result{Passed: true, RuleID: m.id}, nil
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"info", Info, "info"},
		{"warning", Warning, "warning"},
		{"critical", Critical, "critical"},
		{"unknown", Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.severity.String())
		})
	}
}

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()

	require.NotNil(t, reg)
	assert.Empty(t, reg.AllRules())
}

func TestRegister(t *testing.T) {
	reg := NewRegistry()
	rule := &mockRule{id: "test-rule-1", provider: "aws", resourceType: "aws_s3_bucket"}

	reg.Register(rule)

	rules := reg.AllRules()
	require.Len(t, rules, 1)
	assert.Equal(t, "test-rule-1", rules[0].ID())
}

func TestRegisterAll(t *testing.T) {
	reg := NewRegistry()
	rule1 := &mockRule{id: "rule-1", provider: "aws", resourceType: "aws_s3_bucket"}
	rule2 := &mockRule{id: "rule-2", provider: "aws", resourceType: "aws_security_group"}
	rule3 := &mockRule{id: "rule-3", provider: "gcp", resourceType: "google_storage_bucket"}

	reg.RegisterAll(rule1, rule2, rule3)

	rules := reg.AllRules()
	require.Len(t, rules, 3)
	assert.Equal(t, "rule-1", rules[0].ID())
	assert.Equal(t, "rule-2", rules[1].ID())
	assert.Equal(t, "rule-3", rules[2].ID())
}

func TestRulesForResource(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockRule{id: "s3-rule-1", resourceType: "aws_s3_bucket"})
	reg.Register(&mockRule{id: "s3-rule-2", resourceType: "aws_s3_bucket"})
	reg.Register(&mockRule{id: "sg-rule", resourceType: "aws_security_group"})

	matched := reg.RulesForResource("aws_s3_bucket")
	require.Len(t, matched, 2)
	assert.Equal(t, "s3-rule-1", matched[0].ID())
	assert.Equal(t, "s3-rule-2", matched[1].ID())

	noMatch := reg.RulesForResource("aws_lambda_function")
	assert.Empty(t, noMatch)

	emptyReg := NewRegistry()
	assert.Empty(t, emptyReg.RulesForResource("aws_s3_bucket"))
}

func TestRulesForProvider(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockRule{id: "aws-rule-1", provider: "aws", resourceType: "aws_s3_bucket"})
	reg.Register(&mockRule{id: "aws-rule-2", provider: "aws", resourceType: "aws_ec2_instance"})
	reg.Register(&mockRule{id: "gcp-rule", provider: "gcp", resourceType: "google_storage_bucket"})

	matched := reg.RulesForProvider("aws")
	require.Len(t, matched, 2)
	assert.Equal(t, "aws-rule-1", matched[0].ID())
	assert.Equal(t, "aws-rule-2", matched[1].ID())

	noMatch := reg.RulesForProvider("azure")
	assert.Empty(t, noMatch)

	emptyReg := NewRegistry()
	assert.Empty(t, emptyReg.RulesForProvider("aws"))
}

func TestAllRules(t *testing.T) {
	emptyReg := NewRegistry()
	assert.Empty(t, emptyReg.AllRules())

	reg := NewRegistry()
	reg.Register(&mockRule{id: "rule-1"})
	reg.Register(&mockRule{id: "rule-2"})
	reg.Register(&mockRule{id: "rule-3"})
	require.Len(t, reg.AllRules(), 3)
}

func TestRuleByID(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockRule{id: "rule-1"})
	reg.Register(&mockRule{id: "rule-2"})
	reg.Register(&mockRule{id: "rule-3"})

	rule, found := reg.RuleByID("rule-2")
	require.True(t, found)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-2", rule.ID())

	notFound, ok := reg.RuleByID("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, notFound)

	emptyReg := NewRegistry()
	empty, emptyOk := emptyReg.RuleByID("rule-1")
	assert.False(t, emptyOk)
	assert.Nil(t, empty)
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()

	require.NotNil(t, reg)
	assert.Empty(t, reg.AllRules())
}
