package iacprovisioner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const skipJsonLogLine = " msg="

var (
	// ansiLineRegex matches lines starting with ANSI escape codes for text formatting (e.g., colors, styles).
	ansiLineRegex = regexp.MustCompile(`(?m)^\x1b\[[0-9;]*m.*`)
	// tgLogLevel matches log lines containing fields for time, level, prefix, binary, and message, each with non-whitespace values.
	tgLogLevel = regexp.MustCompile(`.*time=\S+ level=\S+ prefix=\S+ binary=\S+ msg=.*`)
)

// Output calls terraform output for the given variable and return its string value representation.
// It only designed to work with primitive terraform types: string, number and bool.
// Please use OutputStructE for anything else.
func Output(options *Options, key string) (string, error) {
	var val interface{}
	err := OutputStruct(options, key, &val)
	return fmt.Sprintf("%v", val), err
}

// OutputRequired calls terraform output for the given variable and return its value. If the value is empty, return an error.
func OutputRequired(options *Options, key string) (string, error) {
	out, err := Output(options, key)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", EmptyOutput(key)
	}

	return out, nil
}

// parseMap takes a map of interfaces and parses the types.
// It is recursive which allows it to support complex nested structures.
// At this time, this function uses https://golang.org/pkg/strconv/#ParseInt
// to determine if a number should be a float or an int. For this reason, if you are
// expecting a float with a zero as the "tenth" you will need to manually convert
// the return value to a float.
//
// This function exists to map return values of the terraform outputs to intuitive
// types. ie, if you are expecting a value of "1" you are implicitly expecting an int.
//
// This also allows the work to be executed recursively to support complex data types.
func parseMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range m {
		switch vt := v.(type) {
		case map[string]interface{}:
			nestedMap, err := parseMap(vt)
			if err != nil {
				return nil, err
			}
			result[k] = nestedMap
		case []interface{}:
			nestedList, err := parseList(vt)
			if err != nil {
				return nil, err
			}
			result[k] = nestedList
		case float64:
			result[k] = parseFloat(vt)
		default:
			result[k] = vt
		}
	}

	return result, nil
}

func parseList(items []interface{}) (_ []interface{}, err error) {
	for i, v := range items {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Map:
			items[i], err = parseMap(rv.Interface().(map[string]interface{}))
		case reflect.Slice, reflect.Array:
			items[i], err = parseList(rv.Interface().([]interface{}))
		case reflect.Float64:
			items[i] = parseFloat(v)
		}
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func parseFloat(v interface{}) interface{} {
	testInt, err := strconv.ParseInt((fmt.Sprintf("%v", v)), 10, 0)
	if err == nil {
		return int(testInt)
	}
	return v
}

// OutputMapOfObjects calls terraform output for the given variable and returns its value as a map of lists/maps.
// Also returns an error object if an error was generated.
// If the output value is not a map of lists/maps, then it fails the test.
func OutputMapOfObjects(options *Options, key string) (map[string]interface{}, error) {
	out, err := OutputJson(options, key)
	if err != nil {
		return nil, err
	}

	var output map[string]interface{}

	if err := json.Unmarshal([]byte(out), &output); err != nil {
		return nil, err
	}

	return parseMap(output)
}

// OutputListOfObjects calls terraform output for the given variable and returns its value as a list of maps/lists.
// Also returns an error object if an error was generated.
// If the output value is not a list of maps/lists, then it fails the test.
func OutputListOfObjects(options *Options, key string) ([]map[string]interface{}, error) {
	out, err := OutputJson(options, key)
	if err != nil {
		return nil, err
	}

	var output []map[string]interface{}

	if err := json.Unmarshal([]byte(out), &output); err != nil {
		return nil, err
	}

	var result []map[string]interface{}

	for _, m := range output {
		newMap, err := parseMap(m)
		if err != nil {
			return nil, err
		}

		result = append(result, newMap)
	}

	return result, nil
}

// OutputList calls terraform output for the given variable and returns its value as a list.
// If the output value is not a list type, then it returns an error.
func OutputList(options *Options, key string) ([]string, error) {
	out, err := OutputJson(options, key)
	if err != nil {
		return nil, err
	}

	var output interface{}
	if err := json.Unmarshal([]byte(out), &output); err != nil {
		return nil, err
	}

	if outputList, isList := output.([]interface{}); isList {
		return parseListOutputTerraform(outputList, key)
	}

	return nil, UnexpectedOutputType{Key: key, ExpectedType: "map or list", ActualType: reflect.TypeOf(output).String()}
}

// Parse a list output in the format it is returned by Terraform 0.12 and newer versions
func parseListOutputTerraform(outputList []interface{}, key string) ([]string, error) {
	list := []string{}

	for _, item := range outputList {
		list = append(list, fmt.Sprintf("%v", item))
	}

	return list, nil
}

// OutputMap calls terraform output for the given variable and returns its value as a map.
// If the output value is not a map type, then it returns an error.
func OutputMap(options *Options, key string) (map[string]string, error) {
	out, err := OutputJson(options, key)
	if err != nil {
		return nil, err
	}

	outputMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(out), &outputMap); err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	for k, v := range outputMap {
		resultMap[k] = fmt.Sprintf("%v", v)
	}
	return resultMap, nil
}

// OutputJson calls terraform output for the given variable and returns the
// result as the json string.
// If key is an empty string, it will return all the output variables.
func OutputJson(options *Options, key string) (string, error) {
	args := []string{"output", "-no-color", "-json"}
	if key != "" {
		args = append(args, key)
	}

	rawJson, err := RunCommand(options, prepend(options.ExtraArgs.Output, args...)...)
	if err != nil {
		return rawJson, err
	}
	return cleanJson(rawJson)
}

// OutputStruct calls terraform output for the given variable and stores the
// result in the value pointed to by v. If v is nil or not a pointer, or if
// the value returned by Terraform is not appropriate for a given target type,
// it returns an error.
func OutputStruct(options *Options, key string, v interface{}) error {
	out, err := OutputJson(options, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(out), &v)
}

// OutputForKeys calls terraform output for the given key list and returns values as a map.
// The returned values are of type interface{} and need to be type casted as necessary. Refer to output_test.go
func OutputForKeys(options *Options, keys []string) (map[string]interface{}, error) {
	out, err := OutputJson(options, "")
	if err != nil {
		return nil, err
	}

	outputMap := map[string]map[string]interface{}{}
	if err := json.Unmarshal([]byte(out), &outputMap); err != nil {
		return nil, err
	}

	if keys == nil {
		outputKeys := make([]string, 0, len(outputMap))
		for k := range outputMap {
			outputKeys = append(outputKeys, k)
		}
		keys = outputKeys
	}

	resultMap := make(map[string]interface{})
	for _, key := range keys {
		value, containsValue := outputMap[key]["value"]
		if !containsValue {
			return nil, OutputKeyNotFound(string(key))
		}
		resultMap[key] = value
	}
	return resultMap, nil
}

// OutputAll calls terraform and returns all the outputs as a map
func OutputAll(options *Options) (map[string]interface{}, error) {
	return OutputForKeys(options, nil)
}

// clean the ANSI characters from the JSON and update formatting
func cleanJson(input string) (string, error) {
	// Remove ANSI escape codes
	cleaned := ansiLineRegex.ReplaceAllString(input, "")
	cleaned = tgLogLevel.ReplaceAllString(cleaned, "")

	lines := strings.Split(cleaned, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.Contains(trimmed, skipJsonLogLine) {
			result = append(result, trimmed)
		}
	}
	ansiClean := strings.Join(result, "\n")

	var jsonObj interface{}
	if err := json.Unmarshal([]byte(ansiClean), &jsonObj); err != nil {
		return "", err
	}

	// Format JSON output with indentation
	normalized, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}
