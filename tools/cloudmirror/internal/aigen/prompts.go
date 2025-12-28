package aigen

import (
	"fmt"
	"strings"
)

// PromptContext contains context information for building prompts.
type PromptContext struct {
	// Existing handler examples from the codebase
	ExampleHandlers map[string]string

	// Response builder patterns
	ResponseBuilderPatterns string

	// State management patterns
	StatePatterns string

	// Test patterns
	TestPatterns string

	// AWS SDK type definitions
	SDKTypes map[string]string

	// Terraform example files for Create operations
	TerraformMainExample string // Example main.tf content
	TerraformTestExample string // Example test.tftest.hcl content
}

// BuildHandlerPrompt builds a prompt for generating a handler implementation.
func BuildHandlerPrompt(service, protocol string, op *OperationInfo, ctx *PromptContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# Generate AWS Handler Implementation

## Service Information
- **Service**: %s
- **Protocol**: %s
- **Operation**: %s
- **Priority**: %s

## Operation Documentation
%s

`, service, protocol, op.Name, op.Priority, op.Documentation))

	// Add parameters section
	sb.WriteString("## Parameters\n\n")
	if len(op.Parameters) > 0 {
		sb.WriteString("| Name | Type | Required | Location |\n")
		sb.WriteString("|------|------|----------|----------|\n")
		for _, p := range op.Parameters {
			required := "No"
			if p.Required {
				required = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", p.Name, p.Type, required, p.Location))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No parameters defined.\n\n")
	}

	// Add protocol-specific guidance
	sb.WriteString("## Protocol Requirements\n\n")
	switch protocol {
	case "query":
		sb.WriteString(`This is a Query Protocol service. Requirements:
- Content-Type: text/xml
- XML wrapper: <ActionResponse xmlns="..."><ActionResult>...</ActionResult><ResponseMetadata>...</ResponseMetadata></ActionResponse>
- Create Result types with XMLName tags
- Use successResponse() helper which calls emulator.BuildQueryResponse()
- Error format: Use errorResponse() helper

`)
	case "json":
		sb.WriteString(`This is a JSON Protocol service. Requirements:
- Content-Type: application/x-amz-json-1.0
- Extract action from X-Amz-Target header
- Return JSON response directly
- Use jsonResponse() helper

`)
	case "rest-xml":
		sb.WriteString(`This is a REST-XML Protocol service. Requirements:
- Content-Type: application/xml
- Direct XML (no wrapper)
- Use RESTXMLResponse helper

`)
	case "rest-json":
		sb.WriteString(`This is a REST-JSON Protocol service. Requirements:
- Content-Type: application/json
- Standard JSON response
- Use RESTJSONResponse helper

`)
	}

	// Add example handlers if available
	if len(ctx.ExampleHandlers) > 0 {
		sb.WriteString("## Example Handler Patterns\n\n")
		sb.WriteString("Follow these existing patterns from the codebase:\n\n")
		for name, code := range ctx.ExampleHandlers {
			sb.WriteString(fmt.Sprintf("### %s\n\n```go\n%s\n```\n\n", name, code))
		}
	}

	// Add response builder patterns
	if ctx.ResponseBuilderPatterns != "" {
		sb.WriteString("## Response Builder Patterns\n\n")
		sb.WriteString("```go\n")
		sb.WriteString(ctx.ResponseBuilderPatterns)
		sb.WriteString("\n```\n\n")
	}

	// Add state patterns
	if ctx.StatePatterns != "" {
		sb.WriteString("## State Management Patterns\n\n")
		sb.WriteString("```go\n")
		sb.WriteString(ctx.StatePatterns)
		sb.WriteString("\n```\n\n")
	}

	// Add SDK types if available
	if inputType, ok := ctx.SDKTypes[op.InputType]; ok {
		sb.WriteString("## SDK Input Type\n\n```go\n")
		sb.WriteString(inputType)
		sb.WriteString("\n```\n\n")
	}
	if outputType, ok := ctx.SDKTypes[op.OutputType]; ok {
		sb.WriteString("## SDK Output Type\n\n```go\n")
		sb.WriteString(outputType)
		sb.WriteString("\n```\n\n")
	}

	// Add generation instructions
	sb.WriteString(`## Generation Instructions

Generate a complete handler implementation following these rules:

1. **Function Signature**:
   func (s *ServiceType) operationName(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error)

2. **Parameter Extraction**:
   - Use getStringValue, getInt32Value, getBoolValue helpers
   - Validate required parameters first
   - Return appropriate error for missing required params

3. **Business Logic**:
   - Check for resource existence/conflicts as needed
   - Create/update/delete resources in state
   - Use state key pattern: <service>:<resource-type>:<id>

4. **Response Building**:
   - Create Result type with XMLName for Query protocol
   - Use successResponse() helper
   - Include all required fields in response

5. **Error Handling**:
   - Use errorResponse() with appropriate AWS error codes
   - Common codes: InvalidParameterValue, ResourceNotFound, ResourceAlreadyExists

6. **Graph Registration** (if applicable):
   - Call registerResource() after creating resources
   - Call addRelationship() for resource relationships
   - Call unregisterResource() before deleting

Generate the handler function and any required Result types. Output only Go code.
`)

	return sb.String()
}

// BuildTestPrompt builds a prompt for generating tests for a handler.
func BuildTestPrompt(service, operation, handlerCode string, ctx *PromptContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# Generate Unit Tests for Handler

## Service: %s
## Operation: %s

## Handler Implementation

`+"```go\n%s\n```\n\n", service, operation, handlerCode))

	// Add test patterns if available
	if ctx.TestPatterns != "" {
		sb.WriteString("## Existing Test Patterns\n\n")
		sb.WriteString("```go\n")
		sb.WriteString(ctx.TestPatterns)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString(`## Test Requirements

Generate comprehensive unit tests covering:

1. **Success Case**: Verify the operation works correctly with valid inputs

2. **Required Parameter Validation**: Test that missing required parameters return appropriate errors

3. **Resource Not Found**: Test behavior when referenced resources don't exist

4. **Error Cases**: Test specific error conditions (duplicates, conflicts, invalid values)

## Test Structure

`+"```go"+`
func TestOperationName_Success(t *testing.T) {
    state := emulator.NewMemoryStateManager()
    validator := emulator.NewSchemaValidator()
    service := NewServiceType(state, validator)

    req := &emulator.AWSRequest{
        Method: "POST",
        Headers: map[string]string{
            "Content-Type": "application/x-www-form-urlencoded",
        },
        Body:   []byte("Action=OperationName&Param1=value1"),
        Action: "OperationName",
    }

    resp, err := service.HandleRequest(context.Background(), req)
    if err != nil {
        t.Fatalf("HandleRequest failed: %v", err)
    }

    // Assert response
    if resp.StatusCode != 200 {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }
}
`+"```\n\n"+`

Generate test functions for this handler. Output only Go test code.
`)

	return sb.String()
}

// BuildTypesPrompt builds a prompt for generating types for an operation.
func BuildTypesPrompt(service, protocol string, op *OperationInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# Generate Types for AWS Operation

## Service: %s
## Protocol: %s
## Operation: %s

## Parameters

`, service, protocol, op.Name))

	if len(op.Parameters) > 0 {
		for _, p := range op.Parameters {
			sb.WriteString(fmt.Sprintf("- %s (%s, required=%v)\n", p.Name, p.Type, p.Required))
		}
	}

	sb.WriteString(`
## Requirements

1. Generate a Result type with XMLName for Query protocol:
`+"```go"+`
type OperationNameResult struct {
    XMLName xml.Name `+"`xml:\"OperationNameResult\"`"+`
    // Response fields
}
`+"```\n\n"+`

2. Generate any entity types needed for the response

3. Follow existing naming conventions from the service

Output only Go type definitions.
`)

	return sb.String()
}
