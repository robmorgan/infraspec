package assert

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/internal/plan"
	"github.com/robmorgan/infraspec/internal/testfile"
)

func loadTestPlan(t *testing.T) *plan.Plan {
	t.Helper()
	data, err := os.ReadFile("../plan/testdata/plans/vpc_basic.json")
	require.NoError(t, err, "failed to read test plan")

	p, err := plan.ParsePlanBytes(data)
	require.NoError(t, err, "failed to parse test plan")

	return p
}

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)
	assert.NotNil(t, engine)
}

func TestNewEvalContext(t *testing.T) {
	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	// Verify Resource map is populated
	assert.Contains(t, ctx.Resource, "aws_vpc.main")
	assert.Contains(t, ctx.Resource, "aws_subnet.public")
	assert.Contains(t, ctx.Resource, "aws_subnet.private")

	// Verify Resources map groups by type
	assert.Contains(t, ctx.Resources, "aws_vpc")
	assert.Contains(t, ctx.Resources, "aws_subnet")
	assert.Len(t, ctx.Resources["aws_subnet"], 2)

	// Verify Output map
	assert.Contains(t, ctx.Output, "vpc_id")
	assert.Equal(t, "vpc-12345678", ctx.Output["vpc_id"])

	// Verify Var map
	assert.Contains(t, ctx.Var, "environment")
	assert.Equal(t, "production", ctx.Var["environment"])
	assert.Contains(t, ctx.Var, "vpc_cidr")
	assert.Equal(t, "10.0.0.0/16", ctx.Var["vpc_cidr"])

	// Verify Changes list
	assert.NotEmpty(t, ctx.Changes)
}

func TestNewEvalContext_NilPlan(t *testing.T) {
	ctx := NewEvalContext(nil)
	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Resource)
	assert.NotNil(t, ctx.Resources)
	assert.NotNil(t, ctx.Output)
	assert.NotNil(t, ctx.Changes)
	assert.NotNil(t, ctx.Var)
}

func TestEngine_SimpleConditions(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "output not empty",
			expression: `output["vpc_id"] != ""`,
			expected:   true,
		},
		{
			name:       "var equals value",
			expression: `tfvar["environment"] == "production"`,
			expected:   true,
		},
		{
			name:       "var not equals",
			expression: `tfvar["environment"] == "staging"`,
			expected:   false,
		},
		{
			name:       "changes not empty",
			expression: `length(changes) > 0`,
			expected:   true,
		},
		{
			name:       "true literal",
			expression: `true`,
			expected:   true,
		},
		{
			name:       "false literal",
			expression: `false`,
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed)
		})
	}
}

func TestEngine_ResourceAccess(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "resource dns support",
			expression: `resource["aws_vpc.main"]["enable_dns_support"] == true`,
			expected:   true,
		},
		{
			name:       "resource cidr block",
			expression: `resource["aws_vpc.main"]["cidr_block"] == "10.0.0.0/16"`,
			expected:   true,
		},
		{
			name:       "resource nested tags",
			expression: `resource["aws_vpc.main"]["tags"]["Name"] == "main-vpc"`,
			expected:   true,
		},
		{
			name:       "resource dns hostnames",
			expression: `resource["aws_vpc.main"]["enable_dns_hostnames"] == true`,
			expected:   true,
		},
		{
			name:       "subnet public ip on launch",
			expression: `resource["aws_subnet.public"]["map_public_ip_on_launch"] == true`,
			expected:   true,
		},
		{
			name:       "private subnet no public ip",
			expression: `resource["aws_subnet.private"]["map_public_ip_on_launch"] == false`,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed, "expression: %s", tc.expression)
		})
	}
}

func TestEngine_CollectionOperations(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "resources count subnets",
			expression: `length(resources["aws_subnet"]) == 2`,
			expected:   true,
		},
		{
			name:       "resources count vpc",
			expression: `length(resources["aws_vpc"]) == 1`,
			expected:   true,
		},
		{
			name:       "contains in list literal",
			expression: `contains(["a", "b", "c"], "b")`,
			expected:   true,
		},
		{
			name:       "contains not in list",
			expression: `contains(["a", "b", "c"], "d")`,
			expected:   false,
		},
		{
			name:       "anytrue with true",
			expression: `anytrue([false, true, false])`,
			expected:   true,
		},
		{
			name:       "anytrue all false",
			expression: `anytrue([false, false, false])`,
			expected:   false,
		},
		{
			name:       "alltrue all true",
			expression: `alltrue([true, true, true])`,
			expected:   true,
		},
		{
			name:       "alltrue with false",
			expression: `alltrue([true, false, true])`,
			expected:   false,
		},
		{
			name:       "length string",
			expression: `length("hello") == 5`,
			expected:   true,
		},
		{
			name:       "contains string",
			expression: `contains("hello world", "world")`,
			expected:   true,
		},
		{
			name:       "contains string not found",
			expression: `contains("hello world", "foo")`,
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed, "expression: %s", tc.expression)
		})
	}
}

func TestEngine_ComplexExpressions(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "check all vpcs have dns support using macros",
			expression: `resources["aws_vpc"].all(v, v["enable_dns_support"] == true)`,
			expected:   true,
		},
		{
			name:       "check any subnet is public",
			expression: `resources["aws_subnet"].exists(s, s["map_public_ip_on_launch"] == true)`,
			expected:   true,
		},
		{
			name:       "combined conditions",
			expression: `tfvar["environment"] == "production" && output["vpc_id"] != ""`,
			expected:   true,
		},
		{
			name:       "or conditions",
			expression: `tfvar["environment"] == "staging" || tfvar["environment"] == "production"`,
			expected:   true,
		},
		{
			name:       "negation",
			expression: `!(tfvar["environment"] == "staging")`,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed, "expression: %s", tc.expression)
		})
	}
}

func TestEngine_ErrorCases(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	ctx := NewEvalContext(nil)

	tests := []struct {
		name           string
		expression     string
		expectEvalErr  bool
		expectParseErr bool
	}{
		{
			name:           "invalid syntax",
			expression:     `this is not valid`,
			expectParseErr: true,
		},
		{
			name:          "missing resource key",
			expression:    `resource["nonexistent"]["attr"] == "value"`,
			expectEvalErr: true,
		},
		{
			name:           "unclosed bracket",
			expression:     `resource["aws_vpc.main"`,
			expectParseErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err) // The function itself shouldn't error

			if tc.expectParseErr || tc.expectEvalErr {
				assert.NotNil(t, result.EvaluationError)
				assert.False(t, result.Passed)
			}
		})
	}
}

func TestEngine_EmptyExpression(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	ctx := NewEvalContext(nil)

	_, err = engine.EvaluateExpression("", ctx)
	assert.Error(t, err)
}

func TestEngine_NilContext(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	result, err := engine.EvaluateExpression("true", nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)
}

func TestEngine_Evaluate(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	assertion := &testfile.Assert{
		ConditionRaw: `output["vpc_id"] != ""`,
		ErrorMessage: "VPC ID should not be empty",
	}

	result, err := engine.Evaluate(assertion, ctx)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, "VPC ID should not be empty", result.ErrorMessage)
	assert.Equal(t, `output["vpc_id"] != ""`, result.Expression)
}

func TestEngine_EvaluateAll(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	assertions := []*testfile.Assert{
		{
			ConditionRaw: `output["vpc_id"] != ""`,
			ErrorMessage: "VPC ID should not be empty",
		},
		{
			ConditionRaw: `tfvar["environment"] == "production"`,
			ErrorMessage: "Environment should be production",
		},
		{
			ConditionRaw: `length(changes) > 0`,
			ErrorMessage: "There should be changes",
		},
	}

	results, err := engine.EvaluateAll(assertions, ctx)
	require.NoError(t, err)
	require.Len(t, results, 3)

	for _, result := range results {
		assert.True(t, result.Passed)
		assert.Nil(t, result.EvaluationError)
	}
}

func TestConvertDotNotation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "output dot notation",
			input:    `output.vpc_id != ""`,
			expected: `output["vpc_id"] != ""`,
		},
		{
			name:     "var dot notation",
			input:    `var.environment == "production"`,
			expected: `tfvar["environment"] == "production"`,
		},
		{
			name:     "resource dot notation simple",
			input:    `resource.aws_vpc.main.cidr_block`,
			expected: `resource["aws_vpc.main"]["cidr_block"]`,
		},
		{
			name:     "resource dot notation nested",
			input:    `resource.aws_vpc.main.tags.Name`,
			expected: `resource["aws_vpc.main"]["tags"]["Name"]`,
		},
		{
			name:     "resources dot notation",
			input:    `resources.aws_vpc`,
			expected: `resources["aws_vpc"]`,
		},
		{
			name:     "mixed notation",
			input:    `output.vpc_id != "" && var.environment == "prod"`,
			expected: `output["vpc_id"] != "" && tfvar["environment"] == "prod"`,
		},
		{
			name:     "bracket notation unchanged",
			input:    `resource["aws_vpc.main"]["cidr_block"]`,
			expected: `resource["aws_vpc.main"]["cidr_block"]`,
		},
		{
			name:     "no conversion needed",
			input:    `length(changes) == 0`,
			expected: `length(changes) == 0`,
		},
		{
			name:     "complex resource expression",
			input:    `resource.aws_vpc.main.enable_dns_support == true`,
			expected: `resource["aws_vpc.main"]["enable_dns_support"] == true`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertDotNotation(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEngine_DotNotationExpressions(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "output dot notation",
			expression: `output.vpc_id != ""`,
			expected:   true,
		},
		{
			name:       "var dot notation",
			expression: `var.environment == "production"`,
			expected:   true,
		},
		{
			name:       "resource dot notation",
			expression: `resource.aws_vpc.main.enable_dns_support == true`,
			expected:   true,
		},
		{
			name:       "resource nested dot notation",
			expression: `resource.aws_vpc.main.tags.Name == "main-vpc"`,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed, "expression: %s", tc.expression)
		})
	}
}

func TestEngine_ChangesMetadata(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	p := loadTestPlan(t)
	ctx := NewEvalContext(p)

	// Verify changes have expected metadata
	require.NotEmpty(t, ctx.Changes)

	// Check first change (should be aws_vpc.main based on test data)
	found := false
	for _, change := range ctx.Changes {
		if change["address"] != "aws_vpc.main" {
			continue
		}
		found = true
		assert.Equal(t, "aws_vpc", change["type"])
		assert.Equal(t, "main", change["name"])
		assert.NotNil(t, change["actions"])
		assert.NotNil(t, change["after"])
		break
	}
	assert.True(t, found, "should find aws_vpc.main in changes")

	// Test expressions using changes
	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "changes not empty",
			expression: `length(changes) > 0`,
			expected:   true,
		},
		{
			name:       "filter changes by type",
			expression: `changes.filter(c, c["type"] == "aws_vpc").size() >= 1`,
			expected:   true,
		},
		{
			name:       "check for create action",
			expression: `changes.exists(c, c["actions"].exists(a, a == "create"))`,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateExpression(tc.expression, ctx)
			require.NoError(t, err)
			assert.Nil(t, result.EvaluationError, "evaluation error: %v", result.EvaluationError)
			assert.Equal(t, tc.expected, result.Passed, "expression: %s", tc.expression)
		})
	}
}
