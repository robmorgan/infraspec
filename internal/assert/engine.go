package assert

import (
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/robmorgan/infraspec/internal/testfile"
)

// newEvalErrorResult creates a failed Result with an evaluation error.
func newEvalErrorResult(expr string, err error) *Result {
	return &Result{
		Passed:          false,
		Expression:      expr,
		EvaluationError: err,
	}
}

// Result represents the outcome of evaluating an assertion.
type Result struct {
	// Passed indicates whether the condition evaluated to true.
	Passed bool

	// Expression is the CEL expression that was evaluated.
	Expression string

	// Value is the actual result value from evaluation.
	Value interface{}

	// ErrorMessage is the user-provided error message for failures.
	ErrorMessage string

	// EvaluationError contains any error that occurred during evaluation.
	EvaluationError error
}

// Engine evaluates CEL expressions against Terraform plan data.
type Engine struct {
	env *cel.Env
}

// NewEngine creates a new assertion engine.
func NewEngine() (*Engine, error) {
	env, err := newCELEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}
	return &Engine{env: env}, nil
}

// Evaluate evaluates an assertion against the given context.
func (e *Engine) Evaluate(assertion *testfile.Assert, ctx *EvalContext) (*Result, error) {
	if assertion == nil {
		return nil, fmt.Errorf("assertion is nil")
	}

	// Get expression string from assertion
	expr := assertion.ConditionRaw
	if expr == "" {
		return nil, fmt.Errorf("assertion condition is empty")
	}

	result, err := e.EvaluateExpression(expr, ctx)
	if err != nil {
		return nil, err
	}

	// Add the user-provided error message
	result.ErrorMessage = assertion.ErrorMessage

	return result, nil
}

// EvaluateExpression evaluates a raw CEL expression string.
func (e *Engine) EvaluateExpression(expr string, ctx *EvalContext) (*Result, error) {
	if expr == "" {
		return nil, fmt.Errorf("expression is empty")
	}

	if ctx == nil {
		ctx = &EvalContext{
			Resource:  make(map[string]map[string]interface{}),
			Resources: make(map[string][]map[string]interface{}),
			Output:    make(map[string]interface{}),
			Changes:   make([]map[string]interface{}, 0),
			Var:       make(map[string]interface{}),
		}
	}

	// Convert HCL dot notation to CEL bracket notation
	celExpr := convertDotNotation(expr)

	// Compile the expression
	ast, issues := e.env.Compile(celExpr)
	if issues != nil && issues.Err() != nil {
		return newEvalErrorResult(expr, fmt.Errorf("failed to compile expression: %w", issues.Err())), nil //nolint:nilerr // intentional design: return Result with error, not error
	}

	// Create program
	prg, err := e.env.Program(ast)
	if err != nil {
		return newEvalErrorResult(expr, fmt.Errorf("failed to create program: %w", err)), nil
	}

	// Build activation map from context
	input := ctx.ToActivation()

	// Evaluate
	out, _, err := prg.Eval(input)
	if err != nil {
		return newEvalErrorResult(expr, fmt.Errorf("failed to evaluate expression: %w", err)), nil
	}

	// Convert result to bool
	result := &Result{
		Expression: expr,
		Value:      out.Value(),
	}

	switch v := out.Value().(type) {
	case bool:
		result.Passed = v
	default:
		// Non-boolean results are considered failures
		result.Passed = false
		result.EvaluationError = fmt.Errorf("expression did not evaluate to boolean, got %T: %v", v, v)
	}

	return result, nil
}

// EvaluateAll evaluates multiple assertions and returns all results.
func (e *Engine) EvaluateAll(assertions []*testfile.Assert, ctx *EvalContext) ([]*Result, error) {
	results := make([]*Result, 0, len(assertions))

	for _, assertion := range assertions {
		result, err := e.Evaluate(assertion, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate assertion: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}
