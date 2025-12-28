package smithy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample Smithy JSON for testing
const sampleSmithyModel = `{
	"smithy": "2.0",
	"metadata": {},
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.api#service": {
					"sdkId": "Test Service"
				},
				"aws.protocols#ec2Query": {}
			}
		},
		"com.amazonaws.test#DescribeVpcs": {
			"type": "operation",
			"input": {
				"target": "com.amazonaws.test#DescribeVpcsRequest"
			},
			"output": {
				"target": "com.amazonaws.test#DescribeVpcsResult"
			}
		},
		"com.amazonaws.test#DescribeVpcsRequest": {
			"type": "structure",
			"members": {
				"VpcIds": {
					"target": "com.amazonaws.test#VpcIdList",
					"traits": {
						"smithy.api#xmlName": "VpcId"
					}
				}
			}
		},
		"com.amazonaws.test#DescribeVpcsResult": {
			"type": "structure",
			"members": {
				"Vpcs": {
					"target": "com.amazonaws.test#VpcList",
					"traits": {
						"smithy.api#xmlName": "vpcSet"
					}
				}
			},
			"traits": {
				"smithy.api#output": {}
			}
		},
		"com.amazonaws.test#Vpc": {
			"type": "structure",
			"members": {
				"VpcId": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#documentation": "The ID of the VPC.",
						"smithy.api#xmlName": "vpcId"
					}
				},
				"CidrBlock": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "cidrBlock",
						"smithy.api#required": {}
					}
				},
				"State": {
					"target": "com.amazonaws.test#VpcState",
					"traits": {
						"smithy.api#xmlName": "state"
					}
				},
				"Tags": {
					"target": "com.amazonaws.test#TagList",
					"traits": {
						"smithy.api#xmlName": "tagSet",
						"smithy.api#xmlFlattened": {}
					}
				}
			}
		},
		"com.amazonaws.test#VpcList": {
			"type": "list",
			"member": {
				"target": "com.amazonaws.test#Vpc",
				"traits": {
					"smithy.api#xmlName": "item"
				}
			}
		},
		"com.amazonaws.test#VpcIdList": {
			"type": "list",
			"member": {
				"target": "smithy.api#String"
			}
		},
		"com.amazonaws.test#VpcState": {
			"type": "enum",
			"members": {
				"PENDING": {
					"target": "smithy.api#Unit",
					"traits": {
						"smithy.api#enumValue": "pending"
					}
				},
				"AVAILABLE": {
					"target": "smithy.api#Unit",
					"traits": {
						"smithy.api#enumValue": "available"
					}
				}
			}
		},
		"com.amazonaws.test#Tag": {
			"type": "structure",
			"members": {
				"Key": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "key"
					}
				},
				"Value": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "value"
					}
				}
			}
		},
		"com.amazonaws.test#TagList": {
			"type": "list",
			"member": {
				"target": "com.amazonaws.test#Tag",
				"traits": {
					"smithy.api#xmlName": "item"
				}
			}
		}
	}
}`

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	model, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "2.0", model.Smithy)
	assert.NotEmpty(t, model.Shapes)
}

func TestParser_GetServiceInfo(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)

	info, err := parser.GetServiceInfo()
	require.NoError(t, err)

	assert.Equal(t, "TestService", info.Name)
	assert.Equal(t, "Test Service", info.FullName)
	assert.Equal(t, "ec2", info.Protocol)
	assert.Equal(t, "com.amazonaws.test", info.Namespace)
}

func TestParser_GetOperations(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)

	ops := parser.GetOperations()
	require.NotNil(t, ops)

	describeVpcs, ok := ops["DescribeVpcs"]
	require.True(t, ok)
	assert.Equal(t, "DescribeVpcsRequest", describeVpcs.InputShape)
	assert.Equal(t, "DescribeVpcsResult", describeVpcs.OutputShape)
}

func TestParser_GetShape(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)

	// Test getting a structure shape
	shape, ok := parser.GetShape("Vpc")
	require.True(t, ok)
	assert.Equal(t, ShapeTypeStructure, shape.Type)
	assert.Len(t, shape.Members, 4)

	// Test getting an enum shape
	enumShape, ok := parser.GetShape("VpcState")
	require.True(t, ok)
	assert.Equal(t, ShapeTypeEnum, enumShape.Type)

	// Test getting a list shape
	listShape, ok := parser.GetShape("VpcList")
	require.True(t, ok)
	assert.Equal(t, ShapeTypeList, listShape.Type)
	assert.NotNil(t, listShape.Member)

	// Test non-existent shape
	_, ok = parser.GetShape("NonExistent")
	assert.False(t, ok)
}

func TestParser_GetOutputShapes(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)

	outputs := parser.GetOutputShapes()
	assert.Contains(t, outputs, "DescribeVpcsResult")
}

func TestParser_GetStructureShapes(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte(sampleSmithyModel))
	require.NoError(t, err)

	structures := parser.GetStructureShapes()
	assert.NotEmpty(t, structures)

	// Check that Vpc is included
	found := false
	for name := range structures {
		if ExtractLocalName(name) == "Vpc" {
			found = true
			break
		}
	}
	assert.True(t, found, "Vpc structure should be in the list")
}

func TestExtractNamespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.amazonaws.ec2#Vpc", "com.amazonaws.ec2"},
		{"com.amazonaws.rds#DBInstance", "com.amazonaws.rds"},
		{"smithy.api#String", "smithy.api"},
		{"NoNamespace", ""},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ExtractNamespace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractLocalName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.amazonaws.ec2#Vpc", "Vpc"},
		{"com.amazonaws.rds#DBInstance", "DBInstance"},
		{"smithy.api#String", "String"},
		{"NoNamespace", "NoNamespace"},
		{"#JustHash", "JustHash"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ExtractLocalName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractXMLTraits(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected XMLTraits
	}{
		{
			name: "xmlName only",
			traits: map[string]interface{}{
				"smithy.api#xmlName": "vpcId",
			},
			expected: XMLTraits{XMLName: "vpcId"},
		},
		{
			name: "ec2QueryName only",
			traits: map[string]interface{}{
				"aws.protocols#ec2QueryName": "VpcId",
			},
			expected: XMLTraits{EC2Name: "VpcId"},
		},
		{
			name: "xmlFlattened",
			traits: map[string]interface{}{
				"smithy.api#xmlFlattened": map[string]interface{}{},
			},
			expected: XMLTraits{IsFlattened: true},
		},
		{
			name: "xmlAttribute",
			traits: map[string]interface{}{
				"smithy.api#xmlAttribute": map[string]interface{}{},
			},
			expected: XMLTraits{IsAttribute: true},
		},
		{
			name: "all traits",
			traits: map[string]interface{}{
				"smithy.api#xmlName":       "items",
				"aws.protocols#ec2QueryName": "ItemSet",
				"smithy.api#xmlFlattened":  map[string]interface{}{},
			},
			expected: XMLTraits{
				XMLName:     "items",
				EC2Name:     "ItemSet",
				IsFlattened: true,
			},
		},
		{
			name:     "empty traits",
			traits:   map[string]interface{}{},
			expected: XMLTraits{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractXMLTraits(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetXMLElementName(t *testing.T) {
	tests := []struct {
		name       string
		memberName string
		traits     map[string]interface{}
		protocol   string
		expected   string
	}{
		{
			name:       "with xmlName",
			memberName: "VpcId",
			traits: map[string]interface{}{
				"smithy.api#xmlName": "vpcId",
			},
			protocol: "ec2",
			expected: "vpcId",
		},
		{
			name:       "without traits - falls back to camelCase",
			memberName: "VpcId",
			traits:     map[string]interface{}{},
			protocol:   "ec2",
			expected:   "vpcId",
		},
		{
			name:       "xmlName takes precedence",
			memberName: "VpcId",
			traits: map[string]interface{}{
				"smithy.api#xmlName":       "vpcIdentifier",
				"aws.protocols#ec2QueryName": "VpcId",
			},
			protocol: "ec2",
			expected: "vpcIdentifier",
		},
		{
			name:       "lowercase member name unchanged",
			memberName: "state",
			traits:     map[string]interface{}{},
			protocol:   "query",
			expected:   "state",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetXMLElementName(tc.memberName, tc.traits, tc.protocol)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRequired(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected bool
	}{
		{
			name: "required present",
			traits: map[string]interface{}{
				"smithy.api#required": map[string]interface{}{},
			},
			expected: true,
		},
		{
			name:     "required absent",
			traits:   map[string]interface{}{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRequired(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name: "documentation present",
			traits: map[string]interface{}{
				"smithy.api#documentation": "The ID of the VPC.",
			},
			expected: "The ID of the VPC.",
		},
		{
			name:     "documentation absent",
			traits:   map[string]interface{}{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetDocumentation(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDeprecated(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected bool
	}{
		{
			name: "deprecated present",
			traits: map[string]interface{}{
				"smithy.api#deprecated": map[string]interface{}{
					"message": "Use something else",
				},
			},
			expected: true,
		},
		{
			name:     "deprecated absent",
			traits:   map[string]interface{}{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDeprecated(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetEnumValue(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name: "enumValue present",
			traits: map[string]interface{}{
				"smithy.api#enumValue": "pending",
			},
			expected: "pending",
		},
		{
			name:     "enumValue absent",
			traits:   map[string]interface{}{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetEnumValue(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToLowerFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VpcId", "vpcId"},
		{"DBInstance", "dBInstance"},
		{"State", "state"},
		{"state", "state"},
		{"A", "a"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := toLowerFirst(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper functions for creating pointers in tests
func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func TestExtractValidationTraits(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected ValidationTraits
	}{
		{
			name: "length with min and max",
			traits: map[string]interface{}{
				"smithy.api#length": map[string]interface{}{
					"min": float64(1),
					"max": float64(100),
				},
			},
			expected: ValidationTraits{
				LengthMin: ptrInt64(1),
				LengthMax: ptrInt64(100),
			},
		},
		{
			name: "length with min only",
			traits: map[string]interface{}{
				"smithy.api#length": map[string]interface{}{
					"min": float64(1),
				},
			},
			expected: ValidationTraits{
				LengthMin: ptrInt64(1),
			},
		},
		{
			name: "length with max only",
			traits: map[string]interface{}{
				"smithy.api#length": map[string]interface{}{
					"max": float64(50),
				},
			},
			expected: ValidationTraits{
				LengthMax: ptrInt64(50),
			},
		},
		{
			name: "pattern trait",
			traits: map[string]interface{}{
				"smithy.api#pattern": "^[A-Za-z0-9]+$",
			},
			expected: ValidationTraits{
				Pattern: "^[A-Za-z0-9]+$",
			},
		},
		{
			name: "range with min and max",
			traits: map[string]interface{}{
				"smithy.api#range": map[string]interface{}{
					"min": float64(0),
					"max": float64(65535),
				},
			},
			expected: ValidationTraits{
				RangeMin: ptrFloat64(0),
				RangeMax: ptrFloat64(65535),
			},
		},
		{
			name: "range with decimal values",
			traits: map[string]interface{}{
				"smithy.api#range": map[string]interface{}{
					"min": float64(0.0),
					"max": float64(1.0),
				},
			},
			expected: ValidationTraits{
				RangeMin: ptrFloat64(0.0),
				RangeMax: ptrFloat64(1.0),
			},
		},
		{
			name: "all constraints combined",
			traits: map[string]interface{}{
				"smithy.api#length": map[string]interface{}{
					"min": float64(1),
					"max": float64(256),
				},
				"smithy.api#pattern": "^[a-z]+$",
			},
			expected: ValidationTraits{
				LengthMin: ptrInt64(1),
				LengthMax: ptrInt64(256),
				Pattern:   "^[a-z]+$",
			},
		},
		{
			name:     "empty traits",
			traits:   map[string]interface{}{},
			expected: ValidationTraits{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractValidationTraits(tc.traits)
			assert.Equal(t, tc.expected.LengthMin, result.LengthMin)
			assert.Equal(t, tc.expected.LengthMax, result.LengthMax)
			assert.Equal(t, tc.expected.Pattern, result.Pattern)
			assert.Equal(t, tc.expected.RangeMin, result.RangeMin)
			assert.Equal(t, tc.expected.RangeMax, result.RangeMax)
		})
	}
}

func TestValidationTraits_HasConstraints(t *testing.T) {
	tests := []struct {
		name     string
		traits   ValidationTraits
		expected bool
	}{
		{"empty", ValidationTraits{}, false},
		{"length min", ValidationTraits{LengthMin: ptrInt64(1)}, true},
		{"length max", ValidationTraits{LengthMax: ptrInt64(100)}, true},
		{"pattern", ValidationTraits{Pattern: ".*"}, true},
		{"range min", ValidationTraits{RangeMin: ptrFloat64(0)}, true},
		{"range max", ValidationTraits{RangeMax: ptrFloat64(100)}, true},
		{"all constraints", ValidationTraits{
			LengthMin: ptrInt64(1),
			LengthMax: ptrInt64(100),
			Pattern:   ".*",
			RangeMin:  ptrFloat64(0),
			RangeMax:  ptrFloat64(100),
		}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.traits.HasConstraints())
		})
	}
}

func TestExtractHTTPTraits(t *testing.T) {
	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected HTTPTraits
	}{
		{
			name:     "no HTTP traits",
			traits:   map[string]interface{}{},
			expected: HTTPTraits{},
		},
		{
			name: "httpHeader trait",
			traits: map[string]interface{}{
				"smithy.api#httpHeader": "X-Amz-Invocation-Type",
			},
			expected: HTTPTraits{Location: "header", LocationName: "X-Amz-Invocation-Type"},
		},
		{
			name: "httpQuery trait",
			traits: map[string]interface{}{
				"smithy.api#httpQuery": "nextToken",
			},
			expected: HTTPTraits{Location: "query", LocationName: "nextToken"},
		},
		{
			name: "httpLabel trait (empty object)",
			traits: map[string]interface{}{
				"smithy.api#httpLabel": map[string]interface{}{},
			},
			expected: HTTPTraits{Location: "uri"},
		},
		{
			name: "httpLabel trait (true)",
			traits: map[string]interface{}{
				"smithy.api#httpLabel": true,
			},
			expected: HTTPTraits{Location: "uri"},
		},
		{
			name: "httpPayload trait (empty object)",
			traits: map[string]interface{}{
				"smithy.api#httpPayload": map[string]interface{}{},
			},
			expected: HTTPTraits{Location: "payload", IsPayload: true},
		},
		{
			name: "httpPayload trait (true)",
			traits: map[string]interface{}{
				"smithy.api#httpPayload": true,
			},
			expected: HTTPTraits{Location: "payload", IsPayload: true},
		},
		{
			name: "priority: header over query",
			traits: map[string]interface{}{
				"smithy.api#httpHeader": "X-Custom",
				"smithy.api#httpQuery":  "param",
			},
			expected: HTTPTraits{Location: "header", LocationName: "X-Custom"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractHTTPTraits(tc.traits)
			assert.Equal(t, tc.expected.Location, result.Location)
			assert.Equal(t, tc.expected.LocationName, result.LocationName)
			assert.Equal(t, tc.expected.IsPayload, result.IsPayload)
		})
	}
}

func TestHTTPTraits_HasHTTPTraits(t *testing.T) {
	tests := []struct {
		name     string
		traits   HTTPTraits
		expected bool
	}{
		{"empty", HTTPTraits{}, false},
		{"header", HTTPTraits{Location: "header", LocationName: "X-Foo"}, true},
		{"query", HTTPTraits{Location: "query", LocationName: "param"}, true},
		{"uri", HTTPTraits{Location: "uri"}, true},
		{"payload", HTTPTraits{Location: "payload", IsPayload: true}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.traits.HasHTTPTraits())
		})
	}
}
