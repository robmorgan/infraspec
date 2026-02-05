package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// AWSModelParser parses AWS SDK service models (Smithy 2.0 format)
type AWSModelParser struct {
	sdkPath string
}

// NewAWSModelParser creates a new AWS model parser
func NewAWSModelParser(sdkPath string) *AWSModelParser {
	return &AWSModelParser{sdkPath: sdkPath}
}

// smithyModel represents the Smithy 2.0 JSON AST format used by AWS SDK Go V2
type smithyModel struct {
	Smithy   string                 `json:"smithy"`
	Metadata map[string]interface{} `json:"metadata"`
	Shapes   map[string]smithyShape `json:"shapes"`
}

type smithyShape struct {
	Type    string                  `json:"type"`
	Members map[string]smithyMember `json:"members"`
	Traits  map[string]interface{}  `json:"traits"`
	Input   *smithyShapeRef         `json:"input"`
	Output  *smithyShapeRef         `json:"output"`
	Errors  []smithyShapeRef        `json:"errors"`
	Target  string                  `json:"target"` // For list/map/enum
}

type smithyMember struct {
	Target string                 `json:"target"`
	Traits map[string]interface{} `json:"traits"`
}

type smithyShapeRef struct {
	Target string `json:"target"`
}

// ParseService parses an AWS service model from the SDK
func (p *AWSModelParser) ParseService(serviceName string) (*models.AWSService, error) {
	modelPath := p.findModelPath(serviceName)
	if modelPath == "" {
		return nil, fmt.Errorf("model not found for service: %s", serviceName)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read model: %w", err)
	}

	var model smithyModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}

	return p.convertSmithyToService(serviceName, &model)
}

// ListServices returns a list of available services in the SDK
func (p *AWSModelParser) ListServices() ([]string, error) {
	var services []string

	modelsPath := filepath.Join(p.sdkPath, "codegen", "sdk-codegen", "aws-models")
	entries, err := os.ReadDir(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			services = append(services, name)
		}
	}

	return services, nil
}

func (p *AWSModelParser) findModelPath(serviceName string) string {
	// Use centralized service name mappings
	modelName := models.GetAWSModelName(strings.ToLower(serviceName))

	// Try the models directory
	candidates := []string{
		filepath.Join(p.sdkPath, "codegen", "sdk-codegen", "aws-models", modelName+".json"),
		filepath.Join(p.sdkPath, "codegen", "sdk-codegen", "aws-models", strings.ToLower(modelName)+".json"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func (p *AWSModelParser) convertSmithyToService(name string, model *smithyModel) (*models.AWSService, error) {
	service := &models.AWSService{
		Name:       name,
		Operations: make(map[string]*models.Operation),
		Shapes:     make(map[string]*models.Shape),
	}

	// Find the service shape to get metadata
	var serviceNamespace string
	for shapeName, shape := range model.Shapes {
		if shape.Type == "service" {
			serviceNamespace = extractNamespace(shapeName)
			service.FullName = p.extractServiceTitle(shape.Traits)
			service.APIVersion = p.extractAPIVersion(shape.Traits)
			service.Protocol = p.extractProtocol(shape.Traits)
			break
		}
	}

	if serviceNamespace == "" {
		// Try to infer namespace from shape names
		for shapeName := range model.Shapes {
			serviceNamespace = extractNamespace(shapeName)
			if serviceNamespace != "" {
				break
			}
		}
	}

	// Extract operations
	for shapeName, shape := range model.Shapes {
		if shape.Type == "operation" {
			opName := extractLocalName(shapeName)
			op := &models.Operation{
				Name:          opName,
				Documentation: p.extractDocumentation(shape.Traits),
				Deprecated:    p.isDeprecated(shape.Traits),
				DeprecatedMsg: p.extractDeprecatedMessage(shape.Traits),
				HTTPMethod:    p.extractHTTPMethod(shape.Traits),
				HTTPPath:      p.extractHTTPPath(shape.Traits),
			}

			if shape.Input != nil {
				op.InputShape = extractLocalName(shape.Input.Target)
				op.Parameters = p.extractParametersFromShape(shape.Input.Target, model.Shapes)
			}

			if shape.Output != nil {
				op.OutputShape = extractLocalName(shape.Output.Target)
			}

			for _, errRef := range shape.Errors {
				op.Errors = append(op.Errors, extractLocalName(errRef.Target))
			}

			service.Operations[opName] = op
		}
	}

	// Extract shapes (for reference)
	for shapeName, shape := range model.Shapes {
		if shape.Type == "structure" || shape.Type == "string" || shape.Type == "integer" ||
			shape.Type == "boolean" || shape.Type == "list" || shape.Type == "map" ||
			shape.Type == "enum" || shape.Type == "timestamp" || shape.Type == "blob" ||
			shape.Type == "long" || shape.Type == "double" || shape.Type == "float" {

			localName := extractLocalName(shapeName)
			service.Shapes[localName] = &models.Shape{
				Name:       localName,
				Type:       shape.Type,
				Deprecated: p.isDeprecated(shape.Traits),
			}
		}
	}

	// Set defaults if not found
	if service.FullName == "" {
		service.FullName = name
	}
	if service.Protocol == "" {
		service.Protocol = "unknown"
	}

	return service, nil
}

func (p *AWSModelParser) extractParametersFromShape(shapeRef string, shapes map[string]smithyShape) []models.Parameter {
	shape, ok := shapes[shapeRef]
	if !ok {
		return nil
	}

	var params []models.Parameter
	requiredSet := p.extractRequired(shape.Traits)

	for memberName, member := range shape.Members {
		param := models.Parameter{
			Name:       memberName,
			ShapeRef:   extractLocalName(member.Target),
			Required:   requiredSet[memberName],
			Deprecated: p.isDeprecated(member.Traits),
		}

		// Get type from target shape
		if targetShape, ok := shapes[member.Target]; ok {
			param.Type = targetShape.Type
		}

		// Check for location trait (header, query, etc.)
		if loc := p.extractLocation(member.Traits); loc != "" {
			param.Location = loc
		}

		params = append(params, param)
	}

	return params
}

func (p *AWSModelParser) extractServiceTitle(traits map[string]interface{}) string {
	if title, ok := traits["aws.api#service"]; ok {
		if svc, ok := title.(map[string]interface{}); ok {
			if name, ok := svc["sdkId"].(string); ok {
				return name
			}
		}
	}
	if title, ok := traits["smithy.api#title"].(string); ok {
		return title
	}
	return ""
}

func (p *AWSModelParser) extractAPIVersion(traits map[string]interface{}) string {
	if svc, ok := traits["aws.api#service"].(map[string]interface{}); ok {
		if version, ok := svc["arnNamespace"].(string); ok {
			return version
		}
	}
	return ""
}

func (p *AWSModelParser) extractProtocol(traits map[string]interface{}) string {
	protocols := []string{
		"aws.protocols#awsQuery",
		"aws.protocols#awsJson1_0",
		"aws.protocols#awsJson1_1",
		"aws.protocols#restJson1",
		"aws.protocols#restXml",
		"aws.protocols#ec2Query",
	}

	protocolNames := map[string]string{
		"aws.protocols#awsQuery":   "query",
		"aws.protocols#awsJson1_0": "json",
		"aws.protocols#awsJson1_1": "json",
		"aws.protocols#restJson1":  "rest-json",
		"aws.protocols#restXml":    "rest-xml",
		"aws.protocols#ec2Query":   "ec2",
	}

	for _, proto := range protocols {
		if _, ok := traits[proto]; ok {
			return protocolNames[proto]
		}
	}

	return "unknown"
}

func (p *AWSModelParser) extractDocumentation(traits map[string]interface{}) string {
	if doc, ok := traits["smithy.api#documentation"].(string); ok {
		return cleanDocumentation(doc)
	}
	return ""
}

func (p *AWSModelParser) isDeprecated(traits map[string]interface{}) bool {
	_, ok := traits["smithy.api#deprecated"]
	return ok
}

func (p *AWSModelParser) extractDeprecatedMessage(traits map[string]interface{}) string {
	if dep, ok := traits["smithy.api#deprecated"].(map[string]interface{}); ok {
		if msg, ok := dep["message"].(string); ok {
			return msg
		}
	}
	return ""
}

func (p *AWSModelParser) extractHTTPMethod(traits map[string]interface{}) string {
	if http, ok := traits["smithy.api#http"].(map[string]interface{}); ok {
		if method, ok := http["method"].(string); ok {
			return method
		}
	}
	return "POST" // Default for query protocol
}

func (p *AWSModelParser) extractHTTPPath(traits map[string]interface{}) string {
	if http, ok := traits["smithy.api#http"].(map[string]interface{}); ok {
		if uri, ok := http["uri"].(string); ok {
			return uri
		}
	}
	return "/"
}

func (p *AWSModelParser) extractRequired(traits map[string]interface{}) map[string]bool {
	required := make(map[string]bool)
	if req, ok := traits["smithy.api#required"]; ok {
		if reqList, ok := req.([]interface{}); ok {
			for _, r := range reqList {
				if name, ok := r.(string); ok {
					required[name] = true
				}
			}
		}
	}
	return required
}

func (p *AWSModelParser) extractLocation(traits map[string]interface{}) string {
	if _, ok := traits["smithy.api#httpHeader"]; ok {
		return "header"
	}
	if _, ok := traits["smithy.api#httpQuery"]; ok {
		return "querystring"
	}
	if _, ok := traits["smithy.api#httpLabel"]; ok {
		return "uri"
	}
	return ""
}

// extractNamespace extracts the namespace from a fully qualified shape name
// e.g., "com.amazonaws.rds#CreateDBInstance" -> "com.amazonaws.rds"
func extractNamespace(shapeName string) string {
	if idx := strings.Index(shapeName, "#"); idx > 0 {
		return shapeName[:idx]
	}
	return ""
}

// extractLocalName extracts the local name from a fully qualified shape name
// e.g., "com.amazonaws.rds#CreateDBInstance" -> "CreateDBInstance"
func extractLocalName(shapeName string) string {
	if idx := strings.Index(shapeName, "#"); idx >= 0 {
		return shapeName[idx+1:]
	}
	return shapeName
}

// cleanDocumentation removes HTML tags and cleans up documentation strings
func cleanDocumentation(doc string) string {
	// Simple HTML tag removal
	replacements := map[string]string{
		"<p>": "", "</p>": " ",
		"<code>": "`", "</code>": "`",
		"<i>": "_", "</i>": "_",
		"<b>": "**", "</b>": "**",
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

	// Remove remaining HTML tags (simple regex-free approach)
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
	if len(doc) > 200 {
		doc = doc[:197] + "..."
	}

	return doc
}
