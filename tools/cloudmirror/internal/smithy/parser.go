package smithy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Parser parses Smithy JSON AST models
type Parser struct {
	model *Model
}

// NewParser creates a new Smithy parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a Smithy JSON model from a file path
func (p *Parser) ParseFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses a Smithy JSON model from bytes
func (p *Parser) Parse(data []byte) (*Model, error) {
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse Smithy model: %w", err)
	}

	p.model = &model
	return &model, nil
}

// ServiceInfo contains metadata about the service
type ServiceInfo struct {
	Name       string
	FullName   string
	APIVersion string
	Protocol   string
	Namespace  string
}

// GetServiceInfo extracts service metadata from the model
func (p *Parser) GetServiceInfo() (*ServiceInfo, error) {
	if p.model == nil {
		return nil, fmt.Errorf("no model parsed")
	}

	info := &ServiceInfo{}

	for shapeName, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeService {
			info.Namespace = ExtractNamespace(shapeName)
			info.Name = ExtractLocalName(shapeName)

			// Extract service name from aws.api#service trait
			if svc, ok := shape.Traits[TraitAWSService].(map[string]interface{}); ok {
				if sdkId, ok := svc["sdkId"].(string); ok {
					info.FullName = sdkId
				}
			}

			// Extract protocol
			info.Protocol = p.extractProtocol(shape.Traits)

			break
		}
	}

	return info, nil
}

// extractProtocol determines the service protocol from traits
func (p *Parser) extractProtocol(traits map[string]interface{}) string {
	protocols := map[string]string{
		TraitAWSQuery:  "query",
		TraitAWSJSON10: "json",
		TraitAWSJSON11: "json",
		TraitRestJSON:  "rest-json",
		TraitRestXML:   "rest-xml",
		TraitEC2Query:  "ec2",
	}

	for traitKey, protocol := range protocols {
		if _, ok := traits[traitKey]; ok {
			return protocol
		}
	}

	return "unknown"
}

// GetOperations returns all operations in the model
func (p *Parser) GetOperations() map[string]*OperationInfo {
	if p.model == nil {
		return nil
	}

	operations := make(map[string]*OperationInfo)

	for shapeName, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeOperation {
			name := ExtractLocalName(shapeName)
			op := &OperationInfo{
				Name: name,
			}

			if shape.Input != nil {
				op.InputShape = ExtractLocalName(shape.Input.Target)
			}
			if shape.Output != nil {
				op.OutputShape = ExtractLocalName(shape.Output.Target)
			}

			operations[name] = op
		}
	}

	return operations
}

// OperationInfo contains information about an operation
type OperationInfo struct {
	Name        string
	InputShape  string
	OutputShape string
}

// GetShape returns a shape by its local name
func (p *Parser) GetShape(localName string) (*Shape, bool) {
	if p.model == nil {
		return nil, false
	}

	// Try to find by full name first
	for shapeName, shape := range p.model.Shapes {
		if ExtractLocalName(shapeName) == localName {
			return &shape, true
		}
	}

	return nil, false
}

// GetShapeByFullName returns a shape by its full qualified name
func (p *Parser) GetShapeByFullName(fullName string) (*Shape, bool) {
	if p.model == nil {
		return nil, false
	}

	shape, ok := p.model.Shapes[fullName]
	return &shape, ok
}

// GetAllShapes returns all shapes in the model
func (p *Parser) GetAllShapes() map[string]Shape {
	if p.model == nil {
		return nil
	}
	return p.model.Shapes
}

// GetStructureShapes returns all structure shapes
func (p *Parser) GetStructureShapes() map[string]Shape {
	if p.model == nil {
		return nil
	}

	structures := make(map[string]Shape)
	for name, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeStructure {
			structures[name] = shape
		}
	}
	return structures
}

// GetOutputShapes returns all shapes used as operation outputs (response types)
func (p *Parser) GetOutputShapes() []string {
	if p.model == nil {
		return nil
	}

	var outputs []string
	seen := make(map[string]bool)

	for _, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeOperation && shape.Output != nil {
			outputName := ExtractLocalName(shape.Output.Target)
			if !seen[outputName] {
				outputs = append(outputs, outputName)
				seen[outputName] = true
			}
		}
	}

	return outputs
}

// GetOutputShapeToOperationMap returns a map from output shape name to operation name.
// This allows looking up which operation a given output shape belongs to.
func (p *Parser) GetOutputShapeToOperationMap() map[string]string {
	if p.model == nil {
		return nil
	}

	outputToOp := make(map[string]string)

	for shapeName, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeOperation && shape.Output != nil {
			opName := ExtractLocalName(shapeName)
			outputName := ExtractLocalName(shape.Output.Target)
			outputToOp[outputName] = opName
		}
	}

	return outputToOp
}

// GetInputShapes returns all shapes used as operation inputs (request types)
func (p *Parser) GetInputShapes() []string {
	if p.model == nil {
		return nil
	}

	var inputs []string
	seen := make(map[string]bool)

	for _, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeOperation && shape.Input != nil {
			inputName := ExtractLocalName(shape.Input.Target)
			if !seen[inputName] {
				inputs = append(inputs, inputName)
				seen[inputName] = true
			}
		}
	}

	return inputs
}

// GetInputShapeToOperationMap returns a map from input shape name to operation name.
// This allows looking up which operation a given input shape belongs to.
func (p *Parser) GetInputShapeToOperationMap() map[string]string {
	if p.model == nil {
		return nil
	}

	inputToOp := make(map[string]string)

	for shapeName, shape := range p.model.Shapes {
		if shape.Type == ShapeTypeOperation && shape.Input != nil {
			opName := ExtractLocalName(shapeName)
			inputName := ExtractLocalName(shape.Input.Target)
			inputToOp[inputName] = opName
		}
	}

	return inputToOp
}

// ExtractNamespace extracts the namespace from a fully qualified shape name
// e.g., "com.amazonaws.ec2#CreateVpc" -> "com.amazonaws.ec2"
func ExtractNamespace(shapeName string) string {
	if idx := strings.Index(shapeName, "#"); idx > 0 {
		return shapeName[:idx]
	}
	return ""
}

// ExtractLocalName extracts the local name from a fully qualified shape name
// e.g., "com.amazonaws.ec2#CreateVpc" -> "CreateVpc"
func ExtractLocalName(shapeName string) string {
	if idx := strings.Index(shapeName, "#"); idx >= 0 {
		return shapeName[idx+1:]
	}
	return shapeName
}

// ResolveTarget resolves a target reference to its local name
// Handles both full qualified names and simple type names
func ResolveTarget(target string) string {
	// Handle smithy API types
	if strings.HasPrefix(target, "smithy.api#") {
		return ExtractLocalName(target)
	}

	return ExtractLocalName(target)
}
