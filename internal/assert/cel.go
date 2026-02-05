package assert

import (
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

// newCELEnvironment creates a CEL environment with plan variables and custom functions.
func newCELEnvironment() (*cel.Env, error) {
	return cel.NewEnv(
		// Declare variables available in expressions
		cel.Variable("resource", cel.MapType(cel.StringType, cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("resources", cel.MapType(cel.StringType, cel.ListType(cel.MapType(cel.StringType, cel.DynType)))),
		cel.Variable("output", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("changes", cel.ListType(cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("tfvar", cel.MapType(cel.StringType, cel.DynType)),

		// Custom functions
		cel.Function("contains",
			// contains(list, item) -> bool
			cel.Overload("contains_list_any",
				[]*cel.Type{cel.ListType(cel.DynType), cel.DynType},
				cel.BoolType,
				cel.BinaryBinding(containsListFunc),
			),
			// contains(string, substr) -> bool
			cel.Overload("contains_string_string",
				[]*cel.Type{cel.StringType, cel.StringType},
				cel.BoolType,
				cel.BinaryBinding(containsStringFunc),
			),
		),

		cel.Function("anytrue",
			// anytrue(list<bool>) -> bool
			cel.Overload("anytrue_list",
				[]*cel.Type{cel.ListType(cel.BoolType)},
				cel.BoolType,
				cel.UnaryBinding(anytrueFunc),
			),
		),

		cel.Function("alltrue",
			// alltrue(list<bool>) -> bool
			cel.Overload("alltrue_list",
				[]*cel.Type{cel.ListType(cel.BoolType)},
				cel.BoolType,
				cel.UnaryBinding(alltrueFunc),
			),
		),

		cel.Function("length",
			// length(list) -> int
			cel.Overload("length_list",
				[]*cel.Type{cel.ListType(cel.DynType)},
				cel.IntType,
				cel.UnaryBinding(lengthListFunc),
			),
			// length(map) -> int
			cel.Overload("length_map",
				[]*cel.Type{cel.MapType(cel.StringType, cel.DynType)},
				cel.IntType,
				cel.UnaryBinding(lengthMapFunc),
			),
			// length(string) -> int
			cel.Overload("length_string",
				[]*cel.Type{cel.StringType},
				cel.IntType,
				cel.UnaryBinding(lengthStringFunc),
			),
		),
	)
}

// containsListFunc checks if a list contains a value.
func containsListFunc(lhs, rhs ref.Val) ref.Val {
	list, ok := lhs.(traits.Lister)
	if !ok {
		return types.NewErr("contains: first argument must be a list")
	}

	iter := list.Iterator()
	for iter.HasNext() == types.True {
		elem := iter.Next()
		if elem.Equal(rhs) == types.True {
			return types.True
		}
	}
	return types.False
}

// containsStringFunc checks if a string contains a substring.
func containsStringFunc(lhs, rhs ref.Val) ref.Val {
	str, ok := lhs.Value().(string)
	if !ok {
		return types.NewErr("contains: first argument must be a string")
	}
	substr, ok := rhs.Value().(string)
	if !ok {
		return types.NewErr("contains: second argument must be a string")
	}
	return types.Bool(strings.Contains(str, substr))
}

// anytrueFunc returns true if any element in the list is true.
func anytrueFunc(val ref.Val) ref.Val {
	list, ok := val.(traits.Lister)
	if !ok {
		return types.NewErr("anytrue: argument must be a list")
	}

	iter := list.Iterator()
	for iter.HasNext() == types.True {
		elem := iter.Next()
		if b, ok := elem.Value().(bool); ok && b {
			return types.True
		}
	}
	return types.False
}

// alltrueFunc returns true if all elements in the list are true.
func alltrueFunc(val ref.Val) ref.Val {
	list, ok := val.(traits.Lister)
	if !ok {
		return types.NewErr("alltrue: argument must be a list")
	}

	iter := list.Iterator()
	for iter.HasNext() == types.True {
		elem := iter.Next()
		if b, ok := elem.Value().(bool); ok && !b {
			return types.False
		}
	}
	return types.True
}

// lengthListFunc returns the length of a list.
func lengthListFunc(val ref.Val) ref.Val {
	list, ok := val.(traits.Lister)
	if !ok {
		return types.NewErr("length: argument must be a list")
	}
	return list.Size()
}

// lengthMapFunc returns the size of a map.
func lengthMapFunc(val ref.Val) ref.Val {
	m, ok := val.(traits.Mapper)
	if !ok {
		return types.NewErr("length: argument must be a map")
	}
	return m.Size()
}

// lengthStringFunc returns the length of a string.
func lengthStringFunc(val ref.Val) ref.Val {
	str, ok := val.Value().(string)
	if !ok {
		return types.NewErr("length: argument must be a string")
	}
	return types.Int(len(str))
}

// Regex patterns for dot notation conversion
var (
	// output.name -> output["name"]
	outputPattern = regexp.MustCompile(`\boutput\.(\w+)`)
	// var.name -> tfvar["name"] (CEL reserves "var" as a keyword)
	varDotPattern = regexp.MustCompile(`\bvar\.(\w+)`)
	// var["name"] -> tfvar["name"] (for bracket notation)
	varBracketPattern = regexp.MustCompile(`\bvar\[`)
	// resource.type.name.attr... -> resource["type.name"]["attr"]...
	// Matches: resource. followed by type.name (e.g., aws_vpc.main) followed by .attr chains
	resourcePattern = regexp.MustCompile(`\bresource\.(\w+)\.(\w+)(\.\w+)+`)
	// resources.type -> resources["type"]
	resourcesPattern = regexp.MustCompile(`\bresources\.(\w+)`)
)

// convertDotNotation converts HCL-style dot notation to CEL bracket notation.
//
// Conversion rules:
//   - output.name -> output["name"]
//   - var.name -> tfvar["name"] (CEL reserves "var")
//   - resource.type.name.attr -> resource["type.name"]["attr"]
//   - resource.type.name.attr.nested -> resource["type.name"]["attr"]["nested"]
//   - resources.type -> resources["type"]
func convertDotNotation(expr string) string {
	result := expr

	// Convert output.name -> output["name"]
	result = outputPattern.ReplaceAllString(result, `output["$1"]`)

	// Convert var.name -> tfvar["name"] (CEL reserves "var" as a keyword)
	result = varDotPattern.ReplaceAllString(result, `tfvar["$1"]`)

	// Convert var["name"] -> tfvar["name"] (for bracket notation)
	result = varBracketPattern.ReplaceAllString(result, `tfvar[`)

	// Convert resource.type.name.attr... -> resource["type.name"]["attr"]...
	// This is more complex - need a custom replacement function
	result = resourcePattern.ReplaceAllStringFunc(result, convertResourceAccess)

	// Convert resources.type -> resources["type"]
	result = resourcesPattern.ReplaceAllString(result, `resources["$1"]`)

	return result
}

// minResourceParts is the minimum number of parts needed for a valid resource path
// (type.name.attr, e.g., "aws_vpc.main.cidr_block").
const minResourceParts = 3

// convertResourceAccess converts a resource access pattern to bracket notation.
// Input:  resource.aws_vpc.main.cidr_block
// Output: resource["aws_vpc.main"]["cidr_block"]
// Input:  resource.aws_vpc.main.tags.Name
// Output: resource["aws_vpc.main"]["tags"]["Name"]
func convertResourceAccess(match string) string {
	// Remove "resource." prefix
	rest := strings.TrimPrefix(match, "resource.")

	// Split into parts
	parts := strings.Split(rest, ".")

	if len(parts) < minResourceParts {
		// Not enough parts for type.name.attr
		return match
	}

	// First two parts are type.name (e.g., aws_vpc.main)
	typeName := parts[0] + "." + parts[1]

	// Remaining parts are attribute access
	attrs := parts[2:]

	// Build the bracket notation
	var sb strings.Builder
	sb.WriteString(`resource["`)
	sb.WriteString(typeName)
	sb.WriteString(`"]`)

	for _, attr := range attrs {
		sb.WriteString(`["`)
		sb.WriteString(attr)
		sb.WriteString(`"]`)
	}

	return sb.String()
}
