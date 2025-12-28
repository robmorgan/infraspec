package generator

import (
	"bytes"
	"embed"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/*.tmpl
var templates embed.FS

// StubGenerator generates stub implementations for missing operations
type StubGenerator struct {
	templates *template.Template
}

// NewStubGenerator creates a new stub generator
func NewStubGenerator() (*StubGenerator, error) {
	funcMap := template.FuncMap{
		"toLowerCamel": toLowerCamel,
		"toTitle":      toTitle,
		"toGoType":     toGoType,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &StubGenerator{templates: tmpl}, nil
}

// stubData is the data passed to stub templates
type stubData struct {
	ServiceName  string
	ServiceTitle string
	ServiceType  string
	PackageName  string
	Protocol     string
	MissingOps   []models.OperationStatus
	GeneratedAt  string
}

// GenerateStubs generates stub implementations for missing operations
func (g *StubGenerator) GenerateStubs(report *models.CoverageReport) (string, error) {
	data := stubData{
		ServiceName:  report.ServiceName,
		ServiceTitle: toTitle(report.ServiceName),
		ServiceType:  toTitle(report.ServiceName) + "Service",
		PackageName:  strings.ToLower(report.ServiceName),
		Protocol:     report.Protocol,
		MissingOps:   report.Missing,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "service_stub.go.tmpl", data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateStubsForPriority generates stubs only for operations of a specific priority
func (g *StubGenerator) GenerateStubsForPriority(report *models.CoverageReport, priority models.Priority) (string, error) {
	// Filter missing ops by priority
	var filteredOps []models.OperationStatus
	for _, op := range report.Missing {
		if op.Priority == priority {
			filteredOps = append(filteredOps, op)
		}
	}

	data := stubData{
		ServiceName:  report.ServiceName,
		ServiceTitle: toTitle(report.ServiceName),
		ServiceType:  toTitle(report.ServiceName) + "Service",
		PackageName:  strings.ToLower(report.ServiceName),
		Protocol:     report.Protocol,
		MissingOps:   filteredOps,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "service_stub.go.tmpl", data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// toLowerCamel converts a string to lowerCamelCase
func toLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	// First character lowercase, rest as-is
	return strings.ToLower(s[:1]) + s[1:]
}

// ToSnakeCase converts a PascalCase or camelCase string to snake_case
// e.g., "DeleteScheduledAction" -> "delete_scheduled_action"
// Handles consecutive uppercase letters (acronyms) properly:
// e.g., "CreateDBInstance" -> "create_db_instance"
func ToSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	var result strings.Builder

	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if this is part of an acronym (consecutive uppercase)
			prevUpper := runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'

			// Add underscore before:
			// 1. Start of a new word (previous is lowercase)
			// 2. End of an acronym followed by a new word (e.g., "DB" in "DBInstance")
			if !prevUpper || nextLower {
				result.WriteByte('_')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// toTitle converts a string to Title Case
func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	caser := cases.Title(language.English)
	return caser.String(s)
}

// GenerateSwitchCases generates just the switch case statements
func (g *StubGenerator) GenerateSwitchCases(report *models.CoverageReport) string {
	var sb strings.Builder
	for _, op := range report.Missing {
		sb.WriteString("case \"")
		sb.WriteString(op.Name)
		sb.WriteString("\":\n\treturn s.")
		sb.WriteString(toLowerCamel(op.Name))
		sb.WriteString("(ctx, params)\n")
	}
	return sb.String()
}

// GenerateOperationList generates a simple list of missing operations
func (g *StubGenerator) GenerateOperationList(report *models.CoverageReport) string {
	var sb strings.Builder
	sb.WriteString("// Missing operations for ")
	sb.WriteString(report.ServiceName)
	sb.WriteString(":\n")

	for _, op := range report.Missing {
		sb.WriteString("// - ")
		sb.WriteString(op.Name)
		sb.WriteString(" (")
		sb.WriteString(string(op.Priority))
		sb.WriteString(" priority)\n")
	}

	return sb.String()
}

// ============================================================================
// Scaffold Generation
// ============================================================================

// scaffoldData is the data passed to scaffold templates
type scaffoldData struct {
	ServiceName      string
	ServiceNameLower string
	ServiceType      string
	PackageName      string
	Protocol         string
	APIVersion       string
	Operations       []operationData
	GeneratedAt      string
}

// operationData represents an operation for scaffold templates
type operationData struct {
	Name           string
	MethodName     string // lowerCamelCase
	Documentation  string
	Deprecated     bool
	Priority       string
	RequiredParams []paramData
	OptionalParams []paramData
}

// paramData represents a parameter for scaffold templates
type paramData struct {
	Name   string
	GoType string
}

// ScaffoldResult contains the generated scaffold files
type ScaffoldResult struct {
	ServiceCode string
	TypesCode   string
}

// GenerateScaffold generates complete service scaffold for a new service
func (g *StubGenerator) GenerateScaffold(awsService *models.AWSService, priority models.Priority) (*ScaffoldResult, error) {
	data := g.buildScaffoldData(awsService, priority)

	// Generate service.go
	var serviceBuf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&serviceBuf, "service_scaffold.go.tmpl", data); err != nil {
		return nil, err
	}

	// Generate types.go
	var typesBuf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&typesBuf, "types_scaffold.go.tmpl", data); err != nil {
		return nil, err
	}

	return &ScaffoldResult{
		ServiceCode: serviceBuf.String(),
		TypesCode:   typesBuf.String(),
	}, nil
}

// buildScaffoldData converts AWS service model to scaffold template data
func (g *StubGenerator) buildScaffoldData(awsService *models.AWSService, priority models.Priority) scaffoldData {
	serviceName := awsService.Name
	serviceTitle := toTitle(serviceName)

	// Build operations list
	var operations []operationData

	// Get sorted operation names for consistent output
	var opNames []string
	for name := range awsService.Operations {
		opNames = append(opNames, name)
	}
	sort.Strings(opNames)

	for _, opName := range opNames {
		op := awsService.Operations[opName]

		// Skip deprecated operations
		if op.Deprecated {
			continue
		}

		// Calculate priority
		opPriority := calculatePriority(opName)

		// Filter by priority if specified
		if priority != "" && opPriority != priority {
			continue
		}

		// Build parameter lists
		var requiredParams, optionalParams []paramData
		for _, param := range op.Parameters {
			p := paramData{
				Name:   param.Name,
				GoType: toGoType(param.Type),
			}
			if param.Required {
				requiredParams = append(requiredParams, p)
			} else {
				optionalParams = append(optionalParams, p)
			}
		}

		operations = append(operations, operationData{
			Name:           opName,
			MethodName:     toLowerCamel(opName),
			Documentation:  cleanDocumentation(op.Documentation),
			Deprecated:     op.Deprecated,
			Priority:       string(opPriority),
			RequiredParams: requiredParams,
			OptionalParams: optionalParams,
		})
	}

	return scaffoldData{
		ServiceName:      serviceName,
		ServiceNameLower: strings.ToLower(serviceName),
		ServiceType:      serviceTitle + "Service",
		PackageName:      strings.ToLower(serviceName),
		Protocol:         awsService.Protocol,
		APIVersion:       awsService.APIVersion,
		Operations:       operations,
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}
}

// calculatePriority determines operation priority based on naming conventions
func calculatePriority(opName string) models.Priority {
	highPriority := []string{"Create", "Describe", "Delete", "List", "Get", "Put"}
	mediumPriority := []string{"Modify", "Update", "Add", "Remove", "Attach", "Detach", "Enable", "Disable"}

	for _, prefix := range highPriority {
		if strings.HasPrefix(opName, prefix) {
			return models.PriorityHigh
		}
	}
	for _, prefix := range mediumPriority {
		if strings.HasPrefix(opName, prefix) {
			return models.PriorityMedium
		}
	}
	return models.PriorityLow
}

// toGoType converts AWS/Smithy types to Go types
func toGoType(smithyType string) string {
	switch strings.ToLower(smithyType) {
	case "string":
		return "string"
	case "integer", "int":
		return "int32"
	case "long":
		return "int64"
	case "boolean", "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "timestamp":
		return "time.Time"
	case "blob":
		return "[]byte"
	case "list":
		return "[]interface{}"
	case "map":
		return "map[string]interface{}"
	case "structure":
		return "interface{}"
	default:
		return "interface{}"
	}
}

// cleanDocumentation removes HTML tags and excessive whitespace from documentation
func cleanDocumentation(doc string) string {
	// Remove HTML tags
	doc = strings.ReplaceAll(doc, "<p>", "")
	doc = strings.ReplaceAll(doc, "</p>", "")
	doc = strings.ReplaceAll(doc, "<code>", "")
	doc = strings.ReplaceAll(doc, "</code>", "")
	doc = strings.ReplaceAll(doc, "<a>", "")
	doc = strings.ReplaceAll(doc, "</a>", "")

	// Replace newlines with spaces
	doc = strings.ReplaceAll(doc, "\n", " ")

	// Collapse multiple spaces
	for strings.Contains(doc, "  ") {
		doc = strings.ReplaceAll(doc, "  ", " ")
	}

	// Trim and truncate if too long
	doc = strings.TrimSpace(doc)
	if len(doc) > 100 {
		doc = doc[:97] + "..."
	}

	return doc
}
