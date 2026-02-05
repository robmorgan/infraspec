package emulator

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
)

// ResponseValidator validates AWS API responses against SDK output types
type ResponseValidator struct {
	registry *ResponseRegistry
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(registry *ResponseRegistry) *ResponseValidator {
	return &ResponseValidator{
		registry: registry,
	}
}

// ValidateResponse validates a response against the expected output type for a service and action
func (v *ResponseValidator) ValidateResponse(serviceName, action string, resp *AWSResponse) error {
	outputType := v.registry.GetResponseType(serviceName, action)
	if outputType == nil {
		// Response type not registered - skip validation (permissive mode)
		return nil
	}

	return v.validateResponseStructure(resp, outputType, serviceName)
}

// ValidateResponseForAction validates a response by searching for the action across all services
func (v *ResponseValidator) ValidateResponseForAction(action string, resp *AWSResponse) error {
	outputType := v.registry.GetResponseTypeForAction(action)
	if outputType == nil {
		// Response type not registered - skip validation (permissive mode)
		return nil
	}

	return v.validateResponseStructure(resp, outputType, "")
}

// validateResponseStructure validates the response body structure against the expected output type
func (v *ResponseValidator) validateResponseStructure(resp *AWSResponse, outputType reflect.Type, serviceName string) error {
	// Determine protocol from service name or Content-Type header
	protocol := GetProtocolForService(serviceName)
	contentType := resp.Headers["Content-Type"]

	// Override protocol detection based on Content-Type if available
	if strings.Contains(contentType, "application/x-amz-json") {
		protocol = ProtocolJSON
	} else if strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "x-amz-json") {
		protocol = ProtocolRESTJSON
	} else if strings.Contains(contentType, "xml") {
		if protocol == ProtocolRESTXML {
			protocol = ProtocolRESTXML
		} else {
			protocol = ProtocolQuery
		}
	}

	// Parse response body based on protocol
	var parsedData interface{}
	var err error

	switch protocol {
	case ProtocolJSON, ProtocolRESTJSON:
		err = json.Unmarshal(resp.Body, &parsedData)
		if err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}
		return v.validateJSONStructure(parsedData, outputType)

	case ProtocolQuery, ProtocolRESTXML:
		// For XML, we need to extract the actual data from the wrapper
		// This is a simplified validation - full XML validation would require
		// unmarshaling into the expected type
		return v.validateXMLStructure(resp.Body, outputType, protocol)

	default:
		// Unknown protocol - skip validation
		return nil
	}
}

// validateJSONStructure validates JSON response structure against output type
func (v *ResponseValidator) validateJSONStructure(data interface{}, outputType reflect.Type) error {
	if outputType.Kind() == reflect.Ptr {
		outputType = outputType.Elem()
	}

	if outputType.Kind() != reflect.Struct {
		// Not a struct type - basic validation only
		return nil
	}

	// Convert data to map for validation
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	// Validate fields exist and types match
	return v.validateStructFields(dataMap, outputType)
}

// validateXMLStructure validates XML response structure
// This is a simplified validation - full validation would require unmarshaling
func (v *ResponseValidator) validateXMLStructure(body []byte, outputType reflect.Type, protocol ProtocolType) error {
	// For Query protocol, XML is wrapped in {Action}Response/{Action}Result
	// For REST-XML, it's direct XML

	// Basic validation: check that XML is well-formed
	var dummy interface{}
	err := xml.Unmarshal(body, &dummy)
	if err != nil {
		return fmt.Errorf("invalid XML structure: %w", err)
	}

	// Additional validation could be done by unmarshaling into the expected type
	// For now, we just verify XML is well-formed
	return nil
}

// validateStructFields validates that fields in dataMap match the struct type
func (v *ResponseValidator) validateStructFields(dataMap map[string]interface{}, structType reflect.Type) error {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name

		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		// Check if field exists in data (optional fields may not be present)
		value, exists := dataMap[fieldName]
		if !exists {
			// Field not present - this is OK for optional fields
			continue
		}

		// Validate field type
		if err := v.validateFieldType(value, field.Type, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType validates a single field's type
func (v *ResponseValidator) validateFieldType(value interface{}, fieldType reflect.Type, fieldName string) error {
	if value == nil {
		return nil
	}

	valueType := reflect.TypeOf(value)

	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	switch fieldType.Kind() {
	case reflect.String:
		if valueType.Kind() != reflect.String {
			return fmt.Errorf("field %s expected string, got %s", fieldName, valueType.Kind())
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		// Accept int types and float64 (JSON numbers)
		if valueType.Kind() != reflect.Int && valueType.Kind() != reflect.Int32 &&
			valueType.Kind() != reflect.Int64 && valueType.Kind() != reflect.Float64 {
			return fmt.Errorf("field %s expected integer, got %s", fieldName, valueType.Kind())
		}
	case reflect.Bool:
		if valueType.Kind() != reflect.Bool {
			return fmt.Errorf("field %s expected boolean, got %s", fieldName, valueType.Kind())
		}
	case reflect.Slice:
		if valueType.Kind() != reflect.Slice {
			return fmt.Errorf("field %s expected slice, got %s", fieldName, valueType.Kind())
		}
	case reflect.Map:
		if valueType.Kind() != reflect.Map {
			return fmt.Errorf("field %s expected map, got %s", fieldName, valueType.Kind())
		}
	case reflect.Struct:
		// Nested struct - validate recursively
		if valueMap, ok := value.(map[string]interface{}); ok {
			return v.validateStructFields(valueMap, fieldType)
		}
	}

	return nil
}

// ValidateResponseHeaders validates that required headers are present
func (v *ResponseValidator) ValidateResponseHeaders(resp *AWSResponse, serviceName string) error {
	protocol := GetProtocolForService(serviceName)
	contentType := resp.Headers["Content-Type"]

	// Override protocol detection based on Content-Type
	if strings.Contains(contentType, "application/x-amz-json") {
		protocol = ProtocolJSON
	} else if strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "x-amz-json") {
		protocol = ProtocolRESTJSON
	}

	// Check Content-Type header
	if contentType == "" {
		return fmt.Errorf("missing Content-Type header")
	}

	// Protocol-specific header requirements
	switch protocol {
	case ProtocolJSON:
		// JSON Protocol requires x-amzn-RequestId
		if resp.Headers["x-amzn-RequestId"] == "" {
			return fmt.Errorf("missing x-amzn-RequestId header for JSON protocol")
		}

	case ProtocolRESTJSON:
		// REST-JSON Protocol requires x-amzn-RequestId
		if resp.Headers["x-amzn-RequestId"] == "" {
			return fmt.Errorf("missing x-amzn-RequestId header for REST-JSON protocol")
		}

	case ProtocolQuery, ProtocolRESTXML:
		// XML protocols don't require specific headers (RequestId is in body)
		// But Content-Type should be xml
		if !strings.Contains(contentType, "xml") {
			return fmt.Errorf("Content-Type should contain 'xml' for XML protocol, got: %s", contentType)
		}
	}

	return nil
}

// ValidateResponseStatusCode validates that status code is appropriate
func (v *ResponseValidator) ValidateResponseStatusCode(statusCode int) error {
	// Valid status codes: 200, 201, 204, 400, 403, 404, 409, 500
	validCodes := map[int]bool{
		200: true,
		201: true,
		204: true,
		400: true,
		403: true,
		404: true,
		409: true,
		500: true,
	}

	if !validCodes[statusCode] {
		return fmt.Errorf("unexpected status code: %d (expected 200, 201, 204, 400, 403, 404, 409, or 500)", statusCode)
	}

	return nil
}
