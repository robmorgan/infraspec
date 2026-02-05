package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Config holds parser configuration
type Config struct {
	VarFile string // Path to tfvars file for variable resolution
}

// Parser parses Terraform HCL files
type Parser struct {
	config    Config
	variables map[string]interface{}
	locals    map[string]interface{}
	hclParser *hclparse.Parser
}

// New creates a new Parser with the given configuration
func New(cfg Config) *Parser {
	return &Parser{
		config:    cfg,
		variables: make(map[string]interface{}),
		locals:    make(map[string]interface{}),
		hclParser: hclparse.NewParser(),
	}
}

// ParseFile parses a single Terraform file and returns the resources
func (p *Parser) ParseFile(path string) ([]Resource, error) {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.parseHCL(path, content)
}

// ParseDirectory parses all .tf files in a directory
func (p *Parser) ParseDirectory(dir string) ([]Resource, error) {
	var allResources []Resource

	// First pass: collect variables and locals
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", path, err)
			}
			if err := p.collectVariables(path, content); err != nil {
				return fmt.Errorf("failed to collect variables from %s: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Load tfvars if specified
	if p.config.VarFile != "" {
		if err := p.loadTfvars(p.config.VarFile); err != nil {
			return nil, fmt.Errorf("failed to load tfvars: %w", err)
		}
	}

	// Second pass: parse resources
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			resources, err := p.ParseFile(path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			allResources = append(allResources, resources...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allResources, nil
}

// collectVariables extracts variable and local definitions from a file
func (p *Parser) collectVariables(path string, content []byte) error {
	file, diags := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("parse error: %s", diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "variable":
			if len(block.Labels) > 0 {
				varName := block.Labels[0]
				// Extract default value if present
				if attr, exists := block.Body.Attributes["default"]; exists {
					val := p.evaluateExpression(attr.Expr)
					p.variables[varName] = val
				}
			}
		case "locals":
			for name, attr := range block.Body.Attributes {
				val := p.evaluateExpression(attr.Expr)
				p.locals[name] = val
			}
		}
	}

	return nil
}

// loadTfvars loads variable values from a tfvars file
func (p *Parser) loadTfvars(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read tfvars: %w", err)
	}

	file, diags := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("parse error: %s", diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	for name, attr := range body.Attributes {
		val := p.evaluateExpression(attr.Expr)
		p.variables[name] = val
	}

	return nil
}

// parseHCL parses HCL content and extracts resources
func (p *Parser) parseHCL(path string, content []byte) ([]Resource, error) {
	file, diags := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse error: %s", diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("unexpected body type")
	}

	var resources []Resource

	for _, block := range body.Blocks {
		if block.Type == "resource" && len(block.Labels) >= 2 {
			resource := Resource{
				Type: block.Labels[0],
				Name: block.Labels[1],
				Location: Location{
					File:   path,
					Line:   block.DefRange().Start.Line,
					Column: block.DefRange().Start.Column,
				},
				Attributes: make(map[string]interface{}),
			}

			// Parse attributes
			p.parseBlockBody(block.Body, resource.Attributes)

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// parseBlockBody parses a block body into a map
func (p *Parser) parseBlockBody(body *hclsyntax.Body, attrs map[string]interface{}) {
	// Parse attributes
	for name, attr := range body.Attributes {
		attrs[name] = p.evaluateExpression(attr.Expr)
	}

	// Parse nested blocks
	for _, block := range body.Blocks {
		blockAttrs := make(map[string]interface{})
		p.parseBlockBody(block.Body, blockAttrs)

		// Handle block as nested object or array of objects
		if existing, ok := attrs[block.Type]; ok {
			// Already have this block type, make it an array
			if arr, ok := existing.([]interface{}); ok {
				attrs[block.Type] = append(arr, blockAttrs)
			} else {
				attrs[block.Type] = []interface{}{existing, blockAttrs}
			}
		} else {
			attrs[block.Type] = blockAttrs
		}
	}
}

// evaluateExpression evaluates an HCL expression to a Go value
func (p *Parser) evaluateExpression(expr hcl.Expression) interface{} {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyToGo(e.Val)

	case *hclsyntax.TemplateExpr:
		// Template string - check if it's just a variable reference
		if len(e.Parts) == 1 {
			return p.evaluateExpression(e.Parts[0])
		}
		// Complex template - mark as computed
		return ComputedValue{}

	case *hclsyntax.TemplateWrapExpr:
		return p.evaluateExpression(e.Wrapped)

	case *hclsyntax.ScopeTraversalExpr:
		return p.evaluateTraversal(e.Traversal)

	case *hclsyntax.TupleConsExpr:
		var result []interface{}
		for _, elem := range e.Exprs {
			result = append(result, p.evaluateExpression(elem))
		}
		return result

	case *hclsyntax.ObjectConsExpr:
		result := make(map[string]interface{})
		for _, item := range e.Items {
			// Try to get the key as a string
			var keyStr string

			// Handle ObjectConsKeyExpr specially - it wraps the actual key expression
			if keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr); ok {
				// ForceNonLiteral means it's a variable reference, otherwise it's a literal key
				if keyExpr.ForceNonLiteral {
					keyVal := p.evaluateExpression(keyExpr.Wrapped)
					if s, ok := keyVal.(string); ok {
						keyStr = s
					} else {
						continue
					}
				} else {
					// Try as scope traversal (simple identifier)
					if trav, ok := keyExpr.Wrapped.(*hclsyntax.ScopeTraversalExpr); ok {
						if len(trav.Traversal) > 0 {
							if root, ok := trav.Traversal[0].(hcl.TraverseRoot); ok {
								keyStr = root.Name
							}
						}
					} else {
						keyVal := p.evaluateExpression(keyExpr.Wrapped)
						if s, ok := keyVal.(string); ok {
							keyStr = s
						} else {
							continue
						}
					}
				}
			} else {
				keyVal := p.evaluateExpression(item.KeyExpr)
				if s, ok := keyVal.(string); ok {
					keyStr = s
				} else {
					continue
				}
			}

			if keyStr != "" {
				result[keyStr] = p.evaluateExpression(item.ValueExpr)
			}
		}
		return result

	case *hclsyntax.FunctionCallExpr:
		// Function calls are computed
		return ComputedValue{}

	case *hclsyntax.ConditionalExpr:
		// Conditionals are computed
		return ComputedValue{}

	case *hclsyntax.BinaryOpExpr:
		// Binary operations are computed
		return ComputedValue{}

	case *hclsyntax.UnaryOpExpr:
		// Unary operations are computed
		return ComputedValue{}

	case *hclsyntax.IndexExpr:
		// Index expressions are computed
		return ComputedValue{}

	case *hclsyntax.RelativeTraversalExpr:
		// Relative traversal
		return ComputedValue{}

	case *hclsyntax.SplatExpr:
		// Splat expressions are computed
		return ComputedValue{}

	default:
		// Unknown expression type
		return ComputedValue{}
	}
}

// evaluateTraversal evaluates a scope traversal (variable reference)
func (p *Parser) evaluateTraversal(traversal hcl.Traversal) interface{} {
	if len(traversal) == 0 {
		return UnknownValue{}
	}

	// Get the root name
	root, ok := traversal[0].(hcl.TraverseRoot)
	if !ok {
		return UnknownValue{}
	}

	switch root.Name {
	case "var":
		if len(traversal) > 1 {
			if attr, ok := traversal[1].(hcl.TraverseAttr); ok {
				if val, exists := p.variables[attr.Name]; exists {
					return p.traverseValue(val, traversal[2:])
				}
			}
		}
		return UnknownValue{}

	case "local":
		if len(traversal) > 1 {
			if attr, ok := traversal[1].(hcl.TraverseAttr); ok {
				if val, exists := p.locals[attr.Name]; exists {
					return p.traverseValue(val, traversal[2:])
				}
			}
		}
		return UnknownValue{}

	case "true":
		return true

	case "false":
		return false

	case "null":
		return nil

	default:
		// Other references (data, resource, etc.) are computed
		return ComputedValue{}
	}
}

// traverseValue navigates through a value using the remaining traversal
func (p *Parser) traverseValue(val interface{}, traversal hcl.Traversal) interface{} {
	current := val
	for _, step := range traversal {
		switch s := step.(type) {
		case hcl.TraverseAttr:
			if m, ok := current.(map[string]interface{}); ok {
				if v, exists := m[s.Name]; exists {
					current = v
				} else {
					return UnknownValue{}
				}
			} else {
				return UnknownValue{}
			}
		case hcl.TraverseIndex:
			key := ctyToGo(s.Key)
			switch c := current.(type) {
			case []interface{}:
				if idx, ok := key.(int); ok && idx >= 0 && idx < len(c) {
					current = c[idx]
				} else {
					return UnknownValue{}
				}
			case map[string]interface{}:
				if k, ok := key.(string); ok {
					if v, exists := c[k]; exists {
						current = v
					} else {
						return UnknownValue{}
					}
				} else {
					return UnknownValue{}
				}
			default:
				return UnknownValue{}
			}
		}
	}
	return current
}

// ctyToGo converts a cty.Value to a Go value
func ctyToGo(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	if !val.IsKnown() {
		return UnknownValue{}
	}

	return ctyTypeToGo(val)
}

// ctyTypeToGo handles the type conversion for known, non-null values
func ctyTypeToGo(val cty.Value) interface{} {
	t := val.Type()
	switch {
	case t == cty.String:
		return val.AsString()
	case t == cty.Number:
		return ctyNumberToGo(val)
	case t == cty.Bool:
		return val.True()
	case t.IsListType() || t.IsTupleType() || t.IsSetType():
		return ctyCollectionToSlice(val)
	case t.IsMapType() || t.IsObjectType():
		return ctyMapToGo(val)
	default:
		return ComputedValue{}
	}
}

// ctyNumberToGo converts a cty number to int or float64
func ctyNumberToGo(val cty.Value) interface{} {
	bf := val.AsBigFloat()
	if bf.IsInt() {
		i, _ := bf.Int64()
		return int(i)
	}
	f, _ := bf.Float64()
	return f
}

// ctyCollectionToSlice converts a cty list/tuple/set to a Go slice
func ctyCollectionToSlice(val cty.Value) []interface{} {
	var result []interface{}
	for it := val.ElementIterator(); it.Next(); {
		_, v := it.Element()
		result = append(result, ctyToGo(v))
	}
	return result
}

// ctyMapToGo converts a cty map/object to a Go map
func ctyMapToGo(val cty.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for it := val.ElementIterator(); it.Next(); {
		k, v := it.Element()
		result[k.AsString()] = ctyToGo(v)
	}
	return result
}

// GetAttribute retrieves a nested attribute from a resource using dot notation
// Supports array wildcards like "ingress[*].from_port"
func GetAttribute(attrs map[string]interface{}, path string) (interface{}, bool) {
	parts := splitPath(path)
	return getAttributeRecursive(attrs, parts)
}

// splitPath splits an attribute path into parts, handling array notation
func splitPath(path string) []string {
	// First split by dots
	var parts []string
	current := ""
	inBracket := false

	for _, ch := range path {
		switch ch {
		case '.':
			if !inBracket && current != "" {
				parts = append(parts, current)
				current = ""
			} else if inBracket {
				current += string(ch)
			}
		case '[':
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			inBracket = true
			current = "["
		case ']':
			current += "]"
			parts = append(parts, current)
			current = ""
			inBracket = false
		default:
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// getAttributeRecursive navigates through nested attributes
func getAttributeRecursive(val interface{}, parts []string) (interface{}, bool) {
	if len(parts) == 0 {
		return val, true
	}

	part := parts[0]
	remaining := parts[1:]

	// Handle array wildcard [*]
	if part == "[*]" {
		arr, ok := val.([]interface{})
		if !ok {
			// Also check if it's a single value that should be treated as an array
			if val != nil {
				return getAttributeRecursive(val, remaining)
			}
			return nil, false
		}
		var results []interface{}
		for _, item := range arr {
			if result, ok := getAttributeRecursive(item, remaining); ok {
				results = append(results, result)
			}
		}
		if len(results) == 0 {
			return nil, false
		}
		return results, true
	}

	// Handle array index [N]
	if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
		indexStr := part[1 : len(part)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, false
		}
		arr, ok := val.([]interface{})
		if !ok || index < 0 || index >= len(arr) {
			return nil, false
		}
		return getAttributeRecursive(arr[index], remaining)
	}

	// Handle object attribute
	obj, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}

	attrVal, exists := obj[part]
	if !exists {
		return nil, false
	}

	return getAttributeRecursive(attrVal, remaining)
}
