package emulator

import (
	"fmt"
	"reflect"
	"strings"
)

// ActionRegistry holds the mapping of actions to their input types for each service
type ActionRegistry struct {
	actions map[string]reflect.Type // key: "servicename:ActionName"
}

// SchemaValidator validates AWS API requests against SDK types
type SchemaValidator struct {
	registry *ActionRegistry
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		registry: &ActionRegistry{
			actions: make(map[string]reflect.Type),
		},
	}
}

// RegisterService registers all actions for a service
func (v *SchemaValidator) RegisterService(serviceName string, actionInputTypes map[string]reflect.Type) {
	for action, inputType := range actionInputTypes {
		key := fmt.Sprintf("%s:%s", strings.ToLower(serviceName), action)
		v.registry.actions[key] = inputType
	}
}

// RegisterAction registers a single action for a service
func (v *SchemaValidator) RegisterAction(serviceName, action string, inputType reflect.Type) {
	key := fmt.Sprintf("%s:%s", strings.ToLower(serviceName), action)
	v.registry.actions[key] = inputType
}

// ValidateRequest performs basic request validation
func (v *SchemaValidator) ValidateRequest(req *AWSRequest) error {
	if req.Action == "" {
		return fmt.Errorf("action is required")
	}

	if req.Method == "" {
		return fmt.Errorf("method is required")
	}

	return nil
}

// ValidateAction validates action parameters against the registered input type
func (v *SchemaValidator) ValidateAction(action string, params map[string]interface{}) error {
	// Try to find the input type for this action
	// We don't require service name here for backward compatibility
	inputType := v.findInputTypeForAction(action)
	if inputType == nil {
		// Action not registered - skip validation (permissive mode)
		return nil
	}

	return v.validateAgainstType(params, inputType)
}

// ValidateServiceAction validates action with explicit service name
func (v *SchemaValidator) ValidateServiceAction(serviceName, action string, params map[string]interface{}) error {
	key := fmt.Sprintf("%s:%s", strings.ToLower(serviceName), action)
	inputType, exists := v.registry.actions[key]

	if !exists {
		// Action not registered - skip validation (permissive mode)
		return nil
	}

	return v.validateAgainstType(params, inputType)
}

// findInputTypeForAction searches for action across all services
func (v *SchemaValidator) findInputTypeForAction(action string) reflect.Type {
	// Search through all registered actions
	for key, inputType := range v.registry.actions {
		// key format: "servicename:ActionName"
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[1] == action {
			return inputType
		}
	}
	return nil
}

// validateAgainstType validates params against a struct type
func (v *SchemaValidator) validateAgainstType(params map[string]interface{}, structType reflect.Type) error {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		// Check for JSON tag (most common)
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name

		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		// Also check XML tag (for services like RDS, S3)
		if jsonTag == "" || jsonTag == "-" {
			xmlTag := field.Tag.Get("xml")
			if xmlTag != "" && xmlTag != "-" {
				fieldName = strings.Split(xmlTag, ",")[0]
			}
		}

		value, exists := params[fieldName]
		if !exists {
			continue
		}

		if err := v.validateFieldType(value, field.Type, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType validates a single field's type
func (v *SchemaValidator) validateFieldType(value interface{}, fieldType reflect.Type, fieldName string) error {
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
		// Accept int types, float64 (JSON numbers), and strings (form-encoded data)
		if valueType.Kind() != reflect.Int && valueType.Kind() != reflect.Int32 &&
			valueType.Kind() != reflect.Int64 && valueType.Kind() != reflect.Float64 &&
			valueType.Kind() != reflect.String {
			return fmt.Errorf("field %s expected integer, got %s", fieldName, valueType.Kind())
		}
	case reflect.Bool:
		// Accept bool and strings (form-encoded data sends "true"/"false" as strings)
		if valueType.Kind() != reflect.Bool && valueType.Kind() != reflect.String {
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
	}

	return nil
}

// GetRegisteredActions returns all registered actions for debugging
func (v *SchemaValidator) GetRegisteredActions() []string {
	actions := make([]string, 0, len(v.registry.actions))
	for key := range v.registry.actions {
		actions = append(actions, key)
	}
	return actions
}
