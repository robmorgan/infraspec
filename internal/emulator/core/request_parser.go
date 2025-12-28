package emulator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// ParseJSONRequest parses a JSON request body into a typed struct.
// Used for DynamoDB, CloudWatch, and other JSON protocol services.
func ParseJSONRequest[T any](body []byte) (*T, error) {
	var input T
	if len(body) == 0 {
		return &input, nil
	}

	if err := json.Unmarshal(body, &input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON request: %w", err)
	}

	return &input, nil
}

// ParseQueryRequest parses a Query Protocol (form-encoded) request into a typed struct.
// Used for IAM, RDS, STS, and other Query protocol services.
//
// The struct fields should have xml tags that match the form field names.
// For example:
//
//	type CreateRoleRequest struct {
//	    RoleName                 *string `xml:"RoleName"`
//	    AssumeRolePolicyDocument *string `xml:"AssumeRolePolicyDocument"`
//	}
//
// Form data: "Action=CreateRole&RoleName=test&AssumeRolePolicyDocument=%7B%7D"
func ParseQueryRequest[T any](body []byte) (*T, error) {
	var input T

	if len(body) == 0 {
		return &input, nil
	}

	// Parse form data
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	// Use reflection to populate struct fields
	if err := populateStructFromForm(&input, values); err != nil {
		return nil, err
	}

	return &input, nil
}

// ParseEC2Request parses an EC2 Query Protocol request into a typed struct.
// EC2 uses a slightly different format than standard Query protocol.
func ParseEC2Request[T any](body []byte) (*T, error) {
	// EC2 uses the same form encoding as Query protocol
	return ParseQueryRequest[T](body)
}

// populateStructFromForm populates a struct from URL form values using reflection.
// It uses xml tags to match form field names to struct fields.
func populateStructFromForm(target interface{}, values url.Values) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Get the xml tag to determine the form field name
		xmlTag := field.Tag.Get("xml")
		if xmlTag == "" || xmlTag == "-" {
			continue
		}

		// Parse the xml tag (e.g., "RoleName", "RoleName,omitempty", "Tags>item,omitempty")
		fieldName := parseXMLTagName(xmlTag)
		if fieldName == "" {
			continue
		}

		// Check if this is a list field (e.g., "Tags>item")
		if strings.Contains(xmlTag, ">") {
			// Handle list fields with numbered suffixes (e.g., Tags.member.1, Tags.member.2)
			if err := populateListField(fieldVal, fieldName, values); err != nil {
				return err
			}
			continue
		}

		// Get the form value
		formValue := values.Get(fieldName)
		if formValue == "" {
			continue
		}

		// Set the field value based on type
		if err := setFieldValue(fieldVal, formValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// parseXMLTagName extracts the field name from an xml tag.
// Examples:
//   - "RoleName" -> "RoleName"
//   - "RoleName,omitempty" -> "RoleName"
//   - "Tags>item,omitempty" -> "Tags"
func parseXMLTagName(tag string) string {
	// Remove options after comma
	if idx := strings.Index(tag, ","); idx >= 0 {
		tag = tag[:idx]
	}

	// For nested tags like "Tags>item", return the parent name
	if idx := strings.Index(tag, ">"); idx >= 0 {
		tag = tag[:idx]
	}

	return tag
}

// setFieldValue sets a struct field value from a string.
func setFieldValue(field reflect.Value, value string) error {
	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// populateListField populates a slice field from numbered form parameters.
// AWS Query protocol uses numbered suffixes for list items:
//   - Tags.member.1.Key=env
//   - Tags.member.1.Value=prod
//   - Tags.member.2.Key=team
//   - Tags.member.2.Value=platform
func populateListField(field reflect.Value, baseName string, values url.Values) error {
	if field.Kind() != reflect.Slice {
		return nil
	}

	// Find all numbered items
	items := make(map[int]url.Values)
	prefix := baseName + ".member."

	for key, vals := range values {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Extract index and remaining key
		remaining := key[len(prefix):]
		dotIdx := strings.Index(remaining, ".")
		var indexStr, subKey string
		if dotIdx >= 0 {
			indexStr = remaining[:dotIdx]
			subKey = remaining[dotIdx+1:]
		} else {
			indexStr = remaining
			subKey = ""
		}

		index, err := strconv.Atoi(indexStr)
		if err != nil {
			continue
		}

		if _, ok := items[index]; !ok {
			items[index] = make(url.Values)
		}

		if subKey != "" {
			items[index].Set(subKey, vals[0])
		} else if len(vals) > 0 {
			// Simple value (e.g., SecurityGroupIds.member.1=sg-123)
			items[index].Set("", vals[0])
		}
	}

	if len(items) == 0 {
		return nil
	}

	// Create slice elements
	elemType := field.Type().Elem()
	for i := 1; ; i++ {
		itemValues, ok := items[i]
		if !ok {
			break
		}

		// Create new element
		elem := reflect.New(elemType)
		elemVal := elem.Elem()

		// Check if it's a simple value or a struct
		if simpleVal := itemValues.Get(""); simpleVal != "" {
			// Simple slice (e.g., []string)
			if err := setFieldValue(elemVal, simpleVal); err != nil {
				return err
			}
		} else {
			// Struct slice (e.g., []Tag)
			if elemVal.Kind() == reflect.Struct {
				if err := populateStructFromForm(elem.Interface(), itemValues); err != nil {
					return err
				}
			}
		}

		field.Set(reflect.Append(field, elemVal))
	}

	return nil
}

// ParseRequest parses a request body into a typed struct based on the protocol.
// This is the main entry point for type-safe request parsing.
func ParseRequest[T any](req *AWSRequest, protocol ProtocolType) (*T, error) {
	switch protocol {
	case ProtocolJSON:
		return ParseJSONRequest[T](req.Body)
	case ProtocolQuery:
		return ParseQueryRequest[T](req.Body)
	case ProtocolRESTXML:
		// REST-XML typically uses query parameters or path for input
		// For now, treat as Query protocol for form data
		return ParseQueryRequest[T](req.Body)
	case ProtocolRESTJSON:
		return ParseJSONRequest[T](req.Body)
	default:
		return ParseQueryRequest[T](req.Body)
	}
}

// GetStringParam extracts a string parameter from parsed parameters map.
// This is a convenience helper for the transition period.
func GetStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

// GetInt32Param extracts an int32 parameter from parsed parameters map.
func GetInt32Param(params map[string]interface{}, key string, defaultValue int32) int32 {
	if val, ok := params[key].(float64); ok {
		return int32(val)
	}
	if val, ok := params[key].(int); ok {
		return int32(val)
	}
	if val, ok := params[key].(int32); ok {
		return val
	}
	if val, ok := params[key].(string); ok {
		var parsed int32
		if _, err := fmt.Sscanf(val, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetBoolParam extracts a bool parameter from parsed parameters map.
func GetBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	if val, ok := params[key].(string); ok {
		if val == "true" || val == "True" || val == "TRUE" {
			return true
		}
		if val == "false" || val == "False" || val == "FALSE" {
			return false
		}
	}
	return defaultValue
}
