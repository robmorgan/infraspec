package testfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// testFileSchema defines the top-level blocks allowed in a test file.
var testFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "variables"},
		{Type: "run", LabelNames: []string{"name"}},
	},
}

// runBlockSchema defines the structure of a run block.
var runBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "module", Required: true},
		{Name: "state", Required: false},
		{Name: "command", Required: false},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "assert"},
	},
}

// assertBlockSchema defines the structure of an assert block.
var assertBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "condition", Required: true},
		{Name: "error_message", Required: true},
	},
}

// ParseFile reads and parses a single .infraspec.hcl file.
func ParseFile(path string) (*TestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}
	return ParseBytes(data, path)
}

// ParseBytes parses HCL content from bytes.
func ParseBytes(data []byte, filename string) (*TestFile, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}
	return decodeTestFile(file, filename)
}

// ParseDir discovers and parses all .infraspec.hcl files in a directory.
func ParseDir(dir string) ([]*TestFile, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	var testFiles []*TestFile
	err = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".infraspec.hcl") {
			tf, parseErr := ParseFile(p)
			if parseErr != nil {
				return fmt.Errorf("failed to parse %s: %w", p, parseErr)
			}
			testFiles = append(testFiles, tf)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover test files: %w", err)
	}

	return testFiles, nil
}

// decodeTestFile decodes the parsed HCL file into a TestFile struct.
func decodeTestFile(file *hcl.File, filename string) (*TestFile, error) {
	tf := &TestFile{
		SourceFile: filename,
		Variables:  make(map[string]interface{}),
	}

	content, diags := file.Body.Content(testFileSchema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode test file structure: %s", diags.Error())
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "variables":
			if err := decodeVariables(block.Body, tf); err != nil {
				return nil, err
			}
		case "run":
			run, err := decodeRun(block)
			if err != nil {
				return nil, err
			}
			tf.Runs = append(tf.Runs, run)
		}
	}

	return tf, nil
}

// decodeVariables decodes the variables block into the TestFile.
func decodeVariables(body hcl.Body, tf *TestFile) error {
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return fmt.Errorf("failed to decode variables: %s", diags.Error())
	}

	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("failed to evaluate variable %q: %s", name, diags.Error())
		}
		tf.Variables[name] = ctyValueToGo(val)
	}

	return nil
}

// decodeRun decodes a run block into a Run struct.
func decodeRun(block *hcl.Block) (*Run, error) {
	run := &Run{
		Name:    block.Labels[0],
		Command: "plan", // default
		Range:   block.DefRange,
	}

	content, diags := block.Body.Content(runBlockSchema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode run block %q: %s", run.Name, diags.Error())
	}

	// Decode required module attribute
	if attr, ok := content.Attributes["module"]; ok {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate module in run %q: %s", run.Name, diags.Error())
		}
		run.Module = val.AsString()
	}

	// Decode optional state attribute
	if attr, ok := content.Attributes["state"]; ok {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate state in run %q: %s", run.Name, diags.Error())
		}
		run.State = val.AsString()
	}

	// Decode optional command attribute
	if attr, ok := content.Attributes["command"]; ok {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate command in run %q: %s", run.Name, diags.Error())
		}
		run.Command = val.AsString()
	}

	// Decode assert blocks
	for _, assertBlock := range content.Blocks {
		assert, err := decodeAssert(assertBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to decode assert in run %q: %w", run.Name, err)
		}
		run.Asserts = append(run.Asserts, assert)
	}

	return run, nil
}

// decodeAssert decodes an assert block into an Assert struct.
func decodeAssert(block *hcl.Block) (*Assert, error) {
	assert := &Assert{
		Range: block.DefRange,
	}

	content, diags := block.Body.Content(assertBlockSchema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode assert block: %s", diags.Error())
	}

	// Store condition as both expression and raw text
	if attr, ok := content.Attributes["condition"]; ok {
		assert.Condition = attr.Expr
		// Extract raw source text for display
		assert.ConditionRaw = extractExprSource(attr.Expr)
	}

	// Decode error_message
	if attr, ok := content.Attributes["error_message"]; ok {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate error_message: %s", diags.Error())
		}
		assert.ErrorMessage = val.AsString()
	}

	return assert, nil
}

// extractExprSource extracts the source code range of an expression.
func extractExprSource(expr hcl.Expression) string {
	r := expr.Range()
	// Return a representation based on the range
	return fmt.Sprintf("%s:%d,%d-%d,%d", r.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}

// ctyValueToGo converts a cty.Value to a native Go value.
func ctyValueToGo(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	ty := val.Type()
	switch {
	case ty == cty.String:
		return val.AsString()
	case ty == cty.Number:
		return ctyNumberToGo(val)
	case ty == cty.Bool:
		return val.True()
	case ty.IsListType(), ty.IsTupleType(), ty.IsSetType():
		return ctyCollectionToSlice(val)
	case ty.IsMapType(), ty.IsObjectType():
		return ctyObjectToMap(val)
	default:
		return val.GoString()
	}
}

// ctyNumberToGo converts a cty.Number to int64 or float64.
func ctyNumberToGo(val cty.Value) interface{} {
	bf := val.AsBigFloat()
	if bf.IsInt() {
		i, _ := bf.Int64()
		return i
	}
	f, _ := bf.Float64()
	return f
}

// ctyCollectionToSlice converts a cty list/tuple/set to a Go slice.
func ctyCollectionToSlice(val cty.Value) []interface{} {
	result := make([]interface{}, 0)
	for it := val.ElementIterator(); it.Next(); {
		_, v := it.Element()
		result = append(result, ctyValueToGo(v))
	}
	return result
}

// ctyObjectToMap converts a cty map/object to a Go map.
func ctyObjectToMap(val cty.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for it := val.ElementIterator(); it.Next(); {
		k, v := it.Element()
		result[k.AsString()] = ctyValueToGo(v)
	}
	return result
}
