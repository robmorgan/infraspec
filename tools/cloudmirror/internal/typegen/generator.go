// Package typegen generates Go types from Smithy models with correct XML serialization tags.
package typegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/smithy"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generator generates Go types from Smithy models
type Generator struct {
	parser   *smithy.Parser
	resolver *smithy.Resolver
	config   *Config
}

// Config holds the generator configuration
type Config struct {
	ServiceName   string   // AWS service name
	PackageName   string   // Go package name
	Protocol      string   // Service protocol (ec2, query, rest-xml)
	OutputPath    string   // Output file path
	ModelPath     string   // Path to the Smithy model file
	ResponseOnly  bool     // Only generate response types (default behavior)
	IncludeInputs bool     // Also generate input types for request parsing
	Operations    []string // Specific operations to generate (empty = all)
	TypeSuffix    string   // Suffix to add to type names
}

// NewGenerator creates a new type generator
func NewGenerator(config *Config) *Generator {
	return &Generator{
		parser: smithy.NewParser(),
		config: config,
	}
}

// Generate parses the Smithy model and generates Go types
func (g *Generator) Generate() (string, error) {
	// Parse the model
	model, err := g.parser.ParseFile(g.config.ModelPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse model: %w", err)
	}

	// Get service info
	serviceInfo, err := g.parser.GetServiceInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get service info: %w", err)
	}

	// Use detected protocol if not specified
	if g.config.Protocol == "" {
		g.config.Protocol = serviceInfo.Protocol
	}

	// Create resolver with detected protocol
	g.resolver = smithy.NewResolver(g.parser, g.config.Protocol)

	// Collect types to generate
	typesToGenerate, err := g.collectTypesToGenerate()
	if err != nil {
		return "", fmt.Errorf("failed to collect types: %w", err)
	}

	// Generate the code
	code, err := g.generateCode(model, typesToGenerate)
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	return code, nil
}

// GenerateToFile generates types and writes to the output file
func (g *Generator) GenerateToFile() error {
	code, err := g.Generate()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(g.config.OutputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(g.config.OutputPath, []byte(code), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// collectTypesToGenerate determines which types need to be generated
func (g *Generator) collectTypesToGenerate() ([]string, error) {
	operations := g.parser.GetOperations()

	// Collect output shapes (response types)
	var outputShapes []string
	if len(g.config.Operations) > 0 {
		// Filter to specific operations
		for _, opName := range g.config.Operations {
			if op, ok := operations[opName]; ok && op.OutputShape != "" {
				outputShapes = append(outputShapes, op.OutputShape)
			}
		}
	} else {
		// All operations
		for _, op := range operations {
			if op.OutputShape != "" {
				outputShapes = append(outputShapes, op.OutputShape)
			}
		}
	}

	// Collect input shapes (request types) if configured
	var inputShapes []string
	if g.config.IncludeInputs {
		if len(g.config.Operations) > 0 {
			// Filter to specific operations
			for _, opName := range g.config.Operations {
				if op, ok := operations[opName]; ok && op.InputShape != "" {
					inputShapes = append(inputShapes, op.InputShape)
				}
			}
		} else {
			// All operations
			for _, op := range operations {
				if op.InputShape != "" {
					inputShapes = append(inputShapes, op.InputShape)
				}
			}
		}
	}

	// Collect all dependencies from output shapes
	allTypes := make(map[string]bool)
	for _, shapeName := range outputShapes {
		deps, err := g.resolver.CollectDependencies(shapeName)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			allTypes[dep] = true
		}
	}

	// Collect all dependencies from input shapes
	for _, shapeName := range inputShapes {
		deps, err := g.resolver.CollectDependencies(shapeName)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			allTypes[dep] = true
		}
	}

	// Convert to sorted slice
	var result []string
	for typeName := range allTypes {
		result = append(result, typeName)
	}
	sort.Strings(result)

	return result, nil
}

// TemplateData holds data for the code template
type TemplateData struct {
	Source              string
	ServiceName         string
	Protocol            string
	PackageName         string
	GeneratedAt         string
	HasTimeImport       bool
	HasXMLImport        bool // True if any response types need XMLName (EC2 protocol)
	HasRegexpImport     bool // True if any pattern validation is used
	HasFmtImport        bool // True if any validation is used (for error formatting)
	UseJSONTags         bool // True for json/rest-json protocols (use json:"" tags instead of xml:"")
	UseHTTPLocationTags bool // True for rest-json/rest-xml protocols (add header/query/uri/payload tags)
	HasUnixTimestamp    bool // True if UnixTimestamp type is needed (JSON protocols with timestamps)
	Types               []GoType
	Enums               []GoEnum // Enum type aliases to generate
}

// GoEnum represents an enum type alias
type GoEnum struct {
	Name          string
	Documentation string
}

// GoType represents a Go type to be generated
type GoType struct {
	Name                string
	Documentation       string
	IsDeprecated        bool
	Fields              []GoField
	IsResponse          bool   // True if this is a top-level response type (operation output)
	IsInput             bool   // True if this is a top-level input type (operation input)
	ResponseElementName string // XML root element name for EC2 protocol (e.g., "DescribeVpcsResponse")
	HasValidation       bool   // True if any field has validation constraints
}

// GoField represents a struct field
type GoField struct {
	Name              string
	GoType            string
	XMLTag            string // XML tag for reference (used by XML protocols)
	StructTag         string // Complete struct tag (json:"x" or xml:"x") based on protocol
	Documentation     string
	UsePointer        bool           // Whether to render as *Type (true for all except slices, maps, enums)
	Validation        ValidationInfo // Validation constraints for template rendering
	ValidationComment string         // Formatted constraint info for doc comment (e.g., "[Length: 1-256]")
	// HTTP location fields for REST protocols
	HTTPLocation     string // "header", "query", "uri", "payload", or ""
	HTTPLocationName string // Location-specific name (header name, query param name, etc.)
	IsPayload        bool   // True if this is the request/response body
}

// ValidationInfo contains validation constraint metadata for template rendering
type ValidationInfo struct {
	HasConstraints bool
	// Length constraints
	HasLength bool
	LengthMin *int64
	LengthMax *int64
	// Pattern constraint
	HasPattern bool
	Pattern    string
	// Range constraints
	HasRange bool
	RangeMin *float64
	RangeMax *float64
	// Type info for validation code generation
	IsString  bool
	IsNumeric bool
	IsSlice   bool
	IsMap     bool
	IsPointer bool
}

// generateCode generates the Go source code
func (g *Generator) generateCode(model *smithy.Model, typeNames []string) (string, error) {
	// Build template data
	data := TemplateData{
		Source:      g.config.ModelPath,
		ServiceName: g.config.ServiceName,
		Protocol:    g.config.Protocol,
		PackageName: g.config.PackageName,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Set UseJSONTags for JSON protocol services
	if g.config.Protocol == "json" || g.config.Protocol == "rest-json" {
		data.UseJSONTags = true
	}

	// Set UseHTTPLocationTags for REST protocols
	if g.config.Protocol == "rest-json" || g.config.Protocol == "rest-xml" {
		data.UseHTTPLocationTags = true
	}

	// Get output shapes for marking response types
	outputShapes := make(map[string]bool)
	for _, name := range g.parser.GetOutputShapes() {
		outputShapes[name] = true
	}

	// Get input shapes for marking input types
	inputShapes := make(map[string]bool)
	if g.config.IncludeInputs {
		for _, name := range g.parser.GetInputShapes() {
			inputShapes[name] = true
		}
	}

	// Get output shape to operation mapping for EC2 protocol response element names
	outputToOperation := g.parser.GetOutputShapeToOperationMap()

	// Track collected enum types
	enumSet := make(map[string]bool)

	// Resolve each type
	for _, typeName := range typeNames {
		resolved, _, err := g.resolver.ResolveShape(typeName)
		if err != nil {
			continue // Skip types that can't be resolved
		}

		if resolved.ShapeType != smithy.ShapeTypeStructure {
			continue // Only generate structures
		}

		goType := GoType{
			Name:          typeName + g.config.TypeSuffix,
			Documentation: cleanDocumentation(resolved.Documentation),
			IsDeprecated:  resolved.IsDeprecated,
			IsResponse:    outputShapes[typeName],
			IsInput:       inputShapes[typeName],
		}

		// Set XML response element name based on protocol
		// - EC2 protocol: BuildEC2Response adds the {Operation}Response wrapper,
		//   so generated types should NOT have XMLName (they're often reused as nested types)
		// - Query protocol (IAM, STS, SQS, etc.): Inner element is {Operation}Result,
		//   wrapped by {Operation}Response from BuildQueryResponse
		if goType.IsResponse {
			if opName, ok := outputToOperation[typeName]; ok {
				switch g.config.Protocol {
				case "ec2":
					// EC2: BuildEC2Response adds the wrapper, so no XMLName needed on generated types.
					// This prevents conflicts when types are used both as outputs and nested fields.
					// Skip setting ResponseElementName for EC2.
				case "query":
					// Query protocol: BuildQueryResponse adds outer {Operation}Response wrapper,
					// so the generated type needs {Operation}Result as its XMLName
					goType.ResponseElementName = opName + "Result"
					data.HasXMLImport = true
				}
			}
		}

		// Convert fields
		for _, field := range resolved.Fields {
			adjustedType := g.adjustGoType(field.GoType, field.TargetShape)
			isEnum := g.isEnumType(field.TargetShape)

			// Collect enum types for generating type aliases
			// Use adjustedType to match the actual type name used in fields
			if isEnum && field.TargetShape != "" {
				enumSet[adjustedType] = true
			}

			// Also collect enum types that appear as slice elements (e.g., []AcceleratorManufacturer)
			if strings.HasPrefix(adjustedType, "[]") {
				innerType := adjustedType[2:]
				if g.isEnumType(innerType) {
					enumSet[innerType] = true
				}
			}

			usePointer := shouldUsePointer(adjustedType, isEnum)
			isPayload := field.HTTP.IsPayload

			// Build struct tag based on protocol and HTTP traits
			var tagParts []string

			// Primary serialization tag (JSON or XML)
			if data.UseJSONTags {
				if isPayload {
					// Payload fields are excluded from JSON serialization
					tagParts = append(tagParts, `json:"-"`)
				} else {
					// JSON protocols: use original member name for JSON serialization
					tagParts = append(tagParts, fmt.Sprintf(`json:"%s,omitempty"`, field.MemberName))
				}
			} else {
				if isPayload {
					// Payload fields are excluded from XML serialization
					tagParts = append(tagParts, `xml:"-"`)
				} else {
					// XML protocols: use XMLTag
					tagParts = append(tagParts, fmt.Sprintf(`xml:"%s"`, field.XMLTag))
				}
			}

			// HTTP location tags (only for REST protocols)
			if data.UseHTTPLocationTags && field.HTTP.Location != "" {
				switch field.HTTP.Location {
				case "header":
					tagParts = append(tagParts, fmt.Sprintf(`header:"%s"`, field.HTTP.LocationName))
				case "query":
					tagParts = append(tagParts, fmt.Sprintf(`query:"%s"`, field.HTTP.LocationName))
				case "uri":
					// URI labels use the member name if no explicit name
					uriName := field.HTTP.LocationName
					if uriName == "" {
						uriName = field.MemberName
					}
					tagParts = append(tagParts, fmt.Sprintf(`uri:"%s"`, uriName))
				case "payload":
					tagParts = append(tagParts, `payload:"true"`)
				}
			}

			structTag := strings.Join(tagParts, " ")

			// For JSON protocols, use UnixTimestamp instead of time.Time
			// AWS JSON protocol expects timestamps as Unix epoch numbers, not RFC3339 strings
			fieldType := adjustedType
			if data.UseJSONTags && strings.Contains(adjustedType, "time.Time") {
				fieldType = strings.ReplaceAll(adjustedType, "time.Time", "UnixTimestamp")
				data.HasUnixTimestamp = true
				data.HasTimeImport = true // UnixTimestamp wraps time.Time
			} else if strings.Contains(adjustedType, "time.Time") {
				data.HasTimeImport = true
			}

			goField := GoField{
				Name:             field.Name,
				GoType:           fieldType,
				XMLTag:           field.XMLTag,
				StructTag:        structTag,
				Documentation:    cleanDocumentation(field.Documentation),
				UsePointer:       shouldUsePointer(fieldType, isEnum),
				HTTPLocation:     field.HTTP.Location,
				HTTPLocationName: field.HTTP.LocationName,
				IsPayload:        isPayload,
			}

			// Add validation info if constraints exist
			if field.Validation.HasConstraints() {
				goField.Validation = buildValidationInfo(field.Validation, fieldType, usePointer)
				goField.ValidationComment = formatValidationComment(field.Validation)
				goType.HasValidation = true

				// Track import needs
				if field.Validation.Pattern != "" {
					data.HasRegexpImport = true
				}
				data.HasFmtImport = true // For error formatting
			}

			goType.Fields = append(goType.Fields, goField)
		}

		// Sort fields by name for consistent output
		sort.Slice(goType.Fields, func(i, j int) bool {
			return goType.Fields[i].Name < goType.Fields[j].Name
		})

		data.Types = append(data.Types, goType)
	}

	// Convert enum set to sorted slice
	for enumName := range enumSet {
		data.Enums = append(data.Enums, GoEnum{Name: enumName})
	}
	sort.Slice(data.Enums, func(i, j int) bool {
		return data.Enums[i].Name < data.Enums[j].Name
	})

	// Sort types - response types first, then alphabetically
	sort.Slice(data.Types, func(i, j int) bool {
		if data.Types[i].IsResponse != data.Types[j].IsResponse {
			return data.Types[i].IsResponse
		}
		return data.Types[i].Name < data.Types[j].Name
	})

	// Execute template
	tmpl, err := template.New("smithy_types.go.tmpl").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).ParseFS(templateFS, "templates/smithy_types.go.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code with error info
		return buf.String(), fmt.Errorf("failed to format code (returning unformatted): %w", err)
	}

	return string(formatted), nil
}

// adjustGoType adjusts the Go type string for the output
func (g *Generator) adjustGoType(goType string, targetShape string) string {
	// Add suffix to custom types (but not enums - they're type aliases, not structs)
	if g.config.TypeSuffix != "" {
		// Check if it's a custom type (starts with uppercase and not a primitive)
		if len(goType) > 0 && goType[0] >= 'A' && goType[0] <= 'Z' {
			if !isPrimitiveType(goType) && !g.isEnumType(targetShape) {
				return goType + g.config.TypeSuffix
			}
		}

		// Handle slices of custom types (enums in slices also shouldn't get suffix)
		if strings.HasPrefix(goType, "[]") {
			innerType := goType[2:]
			if len(innerType) > 0 && innerType[0] >= 'A' && innerType[0] <= 'Z' {
				// Check if the inner type (not the list shape) is an enum
				if !isPrimitiveType(innerType) && !g.isEnumType(innerType) {
					return "[]" + innerType + g.config.TypeSuffix
				}
			}
		}
	}

	return goType
}

// isEnumType checks if the target shape is an enum type
func (g *Generator) isEnumType(targetShape string) bool {
	if targetShape == "" {
		return false
	}
	shape, ok := g.parser.GetShape(targetShape)
	if !ok {
		return false
	}
	return shape.Type == smithy.ShapeTypeEnum
}

// isPrimitiveType checks if a type name is a Go primitive
func isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"Time": true, // time.Time
	}
	return primitives[typeName]
}

// shouldUsePointer determines if a field should use a pointer type.
// Returns true for primitives and nested structs, false for slices, maps, and enums.
//
// Pointer rules (matching AWS SDK patterns):
//   - Primitives (string, int32, bool, time.Time): use pointer (*string, *int32)
//   - Nested structs (DBSubnetGroup): use pointer (*DBSubnetGroup)
//   - Slices ([]Tag): no pointer
//   - Maps (map[string]string): no pointer
//   - Enums (ActivityStreamMode): no pointer (type aliases to string)
func shouldUsePointer(goType string, isEnum bool) bool {
	// Slices never use pointers
	if strings.HasPrefix(goType, "[]") {
		return false
	}
	// Maps never use pointers
	if strings.HasPrefix(goType, "map[") {
		return false
	}
	// Enums are type aliases (e.g., type ActivityStreamMode string) and don't use pointers
	if isEnum {
		return false
	}
	// Everything else (primitives, nested structs) uses pointers
	return true
}

// cleanDocumentation cleans up documentation strings
func cleanDocumentation(doc string) string {
	if doc == "" {
		return ""
	}

	// Remove HTML tags
	replacements := map[string]string{
		"<p>": "", "</p>": " ",
		"<code>": "`", "</code>": "`",
		"<i>": "", "</i>": "",
		"<b>": "", "</b>": "",
		"<ul>": "", "</ul>": "",
		"<li>": "- ", "</li>": " ",
		"<a>": "", "</a>": "",
		"<br>": " ", "<br/>": " ",
		"&lt;": "<", "&gt;": ">",
		"&amp;": "&", "&quot;": "\"",
		"\n": " ",
	}

	for old, new := range replacements {
		doc = strings.ReplaceAll(doc, old, new)
	}

	// Remove remaining HTML tags
	for {
		start := strings.Index(doc, "<")
		if start == -1 {
			break
		}
		end := strings.Index(doc[start:], ">")
		if end == -1 {
			break
		}
		doc = doc[:start] + doc[start+end+1:]
	}

	// Trim and collapse whitespace
	doc = strings.TrimSpace(doc)
	for strings.Contains(doc, "  ") {
		doc = strings.ReplaceAll(doc, "  ", " ")
	}

	// Truncate long documentation
	if len(doc) > 100 {
		doc = doc[:97] + "..."
	}

	return doc
}

// formatValidationComment formats validation constraints for doc comments
func formatValidationComment(v smithy.ValidationTraits) string {
	var parts []string

	if v.LengthMin != nil || v.LengthMax != nil {
		if v.LengthMin != nil && v.LengthMax != nil {
			parts = append(parts, fmt.Sprintf("Length: %d-%d", *v.LengthMin, *v.LengthMax))
		} else if v.LengthMin != nil {
			parts = append(parts, fmt.Sprintf("Min length: %d", *v.LengthMin))
		} else {
			parts = append(parts, fmt.Sprintf("Max length: %d", *v.LengthMax))
		}
	}

	if v.Pattern != "" {
		// Truncate very long patterns
		pattern := v.Pattern
		if len(pattern) > 30 {
			pattern = pattern[:27] + "..."
		}
		parts = append(parts, fmt.Sprintf("Pattern: %s", pattern))
	}

	if v.RangeMin != nil || v.RangeMax != nil {
		if v.RangeMin != nil && v.RangeMax != nil {
			parts = append(parts, fmt.Sprintf("Range: %v-%v", *v.RangeMin, *v.RangeMax))
		} else if v.RangeMin != nil {
			parts = append(parts, fmt.Sprintf("Min: %v", *v.RangeMin))
		} else {
			parts = append(parts, fmt.Sprintf("Max: %v", *v.RangeMax))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// buildValidationInfo converts smithy.ValidationTraits to template-friendly ValidationInfo
func buildValidationInfo(v smithy.ValidationTraits, goType string, usePointer bool) ValidationInfo {
	info := ValidationInfo{
		HasConstraints: v.HasConstraints(),
		IsPointer:      usePointer,
	}

	if v.LengthMin != nil || v.LengthMax != nil {
		info.HasLength = true
		info.LengthMin = v.LengthMin
		info.LengthMax = v.LengthMax
	}

	if v.Pattern != "" {
		info.HasPattern = true
		info.Pattern = v.Pattern
	}

	if v.RangeMin != nil || v.RangeMax != nil {
		info.HasRange = true
		info.RangeMin = v.RangeMin
		info.RangeMax = v.RangeMax
	}

	// Determine type category
	info.IsString = goType == "string"
	info.IsSlice = strings.HasPrefix(goType, "[]")
	info.IsMap = strings.HasPrefix(goType, "map[")
	info.IsNumeric = isNumericType(goType)

	return info
}

// isNumericType checks if a Go type is numeric
func isNumericType(goType string) bool {
	numericTypes := map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
	}
	return numericTypes[goType]
}
