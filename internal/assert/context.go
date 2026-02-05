// Package assert provides a CEL-based assertion engine for evaluating
// expressions from test files against Terraform plan data.
package assert

import (
	"github.com/robmorgan/infraspec/internal/plan"
)

// EvalContext contains all data available for CEL expression evaluation.
type EvalContext struct {
	// Resource maps address -> attribute values (from Change.After).
	// Access: resource["aws_vpc.main"]["cidr_block"]
	Resource map[string]map[string]interface{}

	// Resources groups resources by type.
	// Access: resources["aws_vpc"][0]["cidr_block"]
	Resources map[string][]map[string]interface{}

	// Output maps output name -> value.
	// Access: output["vpc_id"]
	Output map[string]interface{}

	// Changes is the list of all resource changes with metadata.
	// Each change contains: address, type, name, actions, before, after
	Changes []map[string]interface{}

	// Var maps variable name -> value.
	// Access: tfvar["environment"] (note: "var" is a CEL reserved word)
	Var map[string]interface{}
}

// NewEvalContext creates an EvalContext from a Terraform plan.
func NewEvalContext(p *plan.Plan) *EvalContext {
	ctx := newEmptyEvalContext()
	if p == nil {
		return ctx
	}

	ctx.populateFromResourceChanges(p.ResourceChanges)
	ctx.populateOutputs(p.PlannedValues)
	ctx.populateVariables(p.Variables)

	return ctx
}

// newEmptyEvalContext creates an empty EvalContext with initialized maps.
func newEmptyEvalContext() *EvalContext {
	return &EvalContext{
		Resource:  make(map[string]map[string]interface{}),
		Resources: make(map[string][]map[string]interface{}),
		Output:    make(map[string]interface{}),
		Changes:   make([]map[string]interface{}, 0),
		Var:       make(map[string]interface{}),
	}
}

// populateFromResourceChanges builds Resource, Resources, and Changes from plan resource changes.
func (ctx *EvalContext) populateFromResourceChanges(resourceChanges []*plan.ResourceChange) {
	for _, rc := range resourceChanges {
		if rc == nil || rc.Change == nil {
			continue
		}
		ctx.addResourceChange(rc)
	}
}

// addResourceChange processes a single resource change and adds it to the context.
func (ctx *EvalContext) addResourceChange(rc *plan.ResourceChange) {
	// Store in Resource map (address -> attributes)
	after := rc.Change.After
	if after == nil {
		after = make(map[string]interface{})
	}
	ctx.Resource[rc.Address] = after

	// Group by type in Resources map
	ctx.Resources[rc.Type] = append(ctx.Resources[rc.Type], after)

	// Build change metadata
	ctx.Changes = append(ctx.Changes, map[string]interface{}{
		"address": rc.Address,
		"type":    rc.Type,
		"name":    rc.Name,
		"mode":    rc.Mode,
		"actions": actionsToStrings(rc.Change.Actions),
		"before":  rc.Change.Before,
		"after":   rc.Change.After,
	})
}

// populateOutputs builds the Output map from planned values.
func (ctx *EvalContext) populateOutputs(plannedValues *plan.StateValues) {
	if plannedValues == nil || plannedValues.Outputs == nil {
		return
	}
	for name, output := range plannedValues.Outputs {
		if output != nil {
			ctx.Output[name] = output.Value
		}
	}
}

// populateVariables builds the Var map from plan variables.
func (ctx *EvalContext) populateVariables(variables map[string]*plan.Variable) {
	if variables == nil {
		return
	}
	for name, variable := range variables {
		if variable != nil {
			ctx.Var[name] = variable.Value
		}
	}
}

// actionsToStrings converts a slice of Actions to a slice of strings.
func actionsToStrings(actions []plan.Action) []interface{} {
	result := make([]interface{}, len(actions))
	for i, a := range actions {
		result[i] = string(a)
	}
	return result
}

// ToActivation converts the EvalContext to a map suitable for CEL evaluation.
func (ctx *EvalContext) ToActivation() map[string]interface{} {
	return map[string]interface{}{
		"resource":  ctx.Resource,
		"resources": ctx.Resources,
		"output":    ctx.Output,
		"changes":   ctx.Changes,
		"tfvar":     ctx.Var, // "var" is a CEL reserved word, so we use "tfvar"
	}
}
