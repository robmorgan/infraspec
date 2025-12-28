package smithy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test model for resolver tests
const resolverTestModel = `{
	"smithy": "2.0",
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.protocols#ec2Query": {}
			}
		},
		"com.amazonaws.test#Vpc": {
			"type": "structure",
			"members": {
				"VpcId": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "vpcId",
						"smithy.api#required": {}
					}
				},
				"CidrBlock": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "cidrBlock"
					}
				},
				"State": {
					"target": "com.amazonaws.test#VpcState",
					"traits": {
						"smithy.api#xmlName": "state"
					}
				},
				"IsDefault": {
					"target": "smithy.api#Boolean",
					"traits": {
						"smithy.api#xmlName": "isDefault"
					}
				},
				"Tags": {
					"target": "com.amazonaws.test#TagList",
					"traits": {
						"smithy.api#xmlName": "tagSet"
					}
				},
				"CreatedAt": {
					"target": "smithy.api#Timestamp",
					"traits": {
						"smithy.api#xmlName": "createdAt"
					}
				},
				"InstanceCount": {
					"target": "smithy.api#Integer",
					"traits": {
						"smithy.api#xmlName": "instanceCount"
					}
				}
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
		"com.amazonaws.test#StringList": {
			"type": "list",
			"member": {
				"target": "smithy.api#String"
			}
		},
		"com.amazonaws.test#StringMap": {
			"type": "map",
			"key": {
				"target": "smithy.api#String"
			},
			"value": {
				"target": "smithy.api#String"
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
				},
				"NextToken": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "nextToken"
					}
				}
			}
		}
	}
}`

func setupResolver(t *testing.T) (*Resolver, *Parser) {
	parser := NewParser()
	_, err := parser.Parse([]byte(resolverTestModel))
	require.NoError(t, err)

	resolver := NewResolver(parser, "ec2")
	return resolver, parser
}

func TestResolver_ResolveShape_Structure(t *testing.T) {
	resolver, _ := setupResolver(t)

	resolved, deps, err := resolver.ResolveShape("Vpc")
	require.NoError(t, err)
	require.NotNil(t, resolved)

	assert.Equal(t, "Vpc", resolved.Name)
	assert.Equal(t, ShapeTypeStructure, resolved.ShapeType)
	assert.Equal(t, "Vpc", resolved.GoType)

	// Check fields
	assert.NotEmpty(t, resolved.Fields)

	// Find VpcId field
	var vpcIdField *ResolvedField
	var tagsField *ResolvedField
	var stateField *ResolvedField
	var isDefaultField *ResolvedField
	var createdAtField *ResolvedField
	var instanceCountField *ResolvedField

	for i := range resolved.Fields {
		switch resolved.Fields[i].MemberName {
		case "VpcId":
			vpcIdField = &resolved.Fields[i]
		case "Tags":
			tagsField = &resolved.Fields[i]
		case "State":
			stateField = &resolved.Fields[i]
		case "IsDefault":
			isDefaultField = &resolved.Fields[i]
		case "CreatedAt":
			createdAtField = &resolved.Fields[i]
		case "InstanceCount":
			instanceCountField = &resolved.Fields[i]
		}
	}

	// Verify VpcId field
	require.NotNil(t, vpcIdField)
	assert.Equal(t, "VpcId", vpcIdField.Name)
	assert.Equal(t, "string", vpcIdField.GoType)
	assert.Equal(t, "vpcId", vpcIdField.XMLName)
	assert.True(t, vpcIdField.IsRequired)

	// Verify Tags field (list of structures)
	require.NotNil(t, tagsField)
	assert.Equal(t, "[]Tag", tagsField.GoType)
	assert.Contains(t, tagsField.XMLTag, "tagSet>item") // May have ,omitempty

	// Verify State field (enum)
	require.NotNil(t, stateField)
	assert.Equal(t, "VpcState", stateField.GoType) // Enums use their type name

	// Verify IsDefault field (boolean)
	require.NotNil(t, isDefaultField)
	assert.Equal(t, "bool", isDefaultField.GoType)

	// Verify CreatedAt field (timestamp)
	require.NotNil(t, createdAtField)
	assert.Equal(t, "time.Time", createdAtField.GoType)

	// Verify InstanceCount field (integer)
	require.NotNil(t, instanceCountField)
	assert.Equal(t, "int32", instanceCountField.GoType)

	// Check dependencies
	assert.Contains(t, deps, "Tag")
}

func TestResolver_ResolveShape_Enum(t *testing.T) {
	resolver, _ := setupResolver(t)

	resolved, _, err := resolver.ResolveShape("VpcState")
	require.NoError(t, err)
	require.NotNil(t, resolved)

	assert.Equal(t, ShapeTypeEnum, resolved.ShapeType)
	assert.Equal(t, "string", resolved.GoType)
	assert.Len(t, resolved.EnumValues, 2)

	// Check enum values
	values := make(map[string]string)
	for _, ev := range resolved.EnumValues {
		values[ev.Name] = ev.Value
	}
	assert.Equal(t, "pending", values["PENDING"])
	assert.Equal(t, "available", values["AVAILABLE"])
}

func TestResolver_ResolveShape_List(t *testing.T) {
	resolver, _ := setupResolver(t)

	// List of structures
	resolved, deps, err := resolver.ResolveShape("VpcList")
	require.NoError(t, err)
	require.NotNil(t, resolved)

	assert.Equal(t, ShapeTypeList, resolved.ShapeType)
	assert.Equal(t, "[]Vpc", resolved.GoType)
	assert.Equal(t, "Vpc", resolved.ListItemType)
	assert.Contains(t, deps, "Vpc")

	// List of strings
	stringList, _, err := resolver.ResolveShape("StringList")
	require.NoError(t, err)
	assert.Equal(t, "[]string", stringList.GoType)
}

func TestResolver_ResolveShape_Map(t *testing.T) {
	resolver, _ := setupResolver(t)

	resolved, _, err := resolver.ResolveShape("StringMap")
	require.NoError(t, err)
	require.NotNil(t, resolved)

	assert.Equal(t, ShapeTypeMap, resolved.ShapeType)
	assert.Equal(t, "map[string]string", resolved.GoType)
	assert.Equal(t, "string", resolved.MapKeyType)
	assert.Equal(t, "string", resolved.MapValueType)
}

func TestResolver_PrimitiveToGoType(t *testing.T) {
	resolver, _ := setupResolver(t)

	tests := []struct {
		smithyType string
		expected   string
	}{
		{ShapeTypeString, "string"},
		{"String", "string"},
		{ShapeTypeInteger, "int32"},
		{"Integer", "int32"},
		{ShapeTypeLong, "int64"},
		{"Long", "int64"},
		{ShapeTypeShort, "int16"},
		{ShapeTypeByte, "int8"},
		{ShapeTypeFloat, "float32"},
		{ShapeTypeDouble, "float64"},
		{ShapeTypeBoolean, "bool"},
		{"Boolean", "bool"},
		{ShapeTypeTimestamp, "time.Time"},
		{"Timestamp", "time.Time"},
		{ShapeTypeBlob, "[]byte"},
		{ShapeTypeBigInt, "string"},
		{ShapeTypeBigDec, "string"},
		{ShapeTypeDocument, "interface{}"},
		{"unknown", "string"},
	}

	for _, tc := range tests {
		t.Run(tc.smithyType, func(t *testing.T) {
			result := resolver.primitiveToGoType(tc.smithyType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResolver_CollectDependencies(t *testing.T) {
	resolver, _ := setupResolver(t)

	// Collect dependencies for Vpc
	deps, err := resolver.CollectDependencies("Vpc")
	require.NoError(t, err)

	// Should include Vpc and Tag
	assert.Contains(t, deps, "Vpc")
	assert.Contains(t, deps, "Tag")

	// Should not include primitive types or enums
	for _, dep := range deps {
		assert.NotEqual(t, "VpcState", dep) // Enum, not structure
		assert.NotEqual(t, "String", dep)
	}
}

func TestResolver_CollectDependencies_Nested(t *testing.T) {
	resolver, _ := setupResolver(t)

	// Collect dependencies for DescribeVpcsResult
	deps, err := resolver.CollectDependencies("DescribeVpcsResult")
	require.NoError(t, err)

	// Should include the result, Vpc, and Tag
	assert.Contains(t, deps, "DescribeVpcsResult")
	assert.Contains(t, deps, "Vpc")
	assert.Contains(t, deps, "Tag")
}

func TestResolver_XMLTagGeneration(t *testing.T) {
	resolver, _ := setupResolver(t)

	resolved, _, err := resolver.ResolveShape("DescribeVpcsResult")
	require.NoError(t, err)

	// Find Vpcs field
	var vpcsField *ResolvedField
	for i := range resolved.Fields {
		if resolved.Fields[i].MemberName == "Vpcs" {
			vpcsField = &resolved.Fields[i]
			break
		}
	}

	require.NotNil(t, vpcsField)
	// List of Vpc with xmlName "vpcSet" and item xmlName "item"
	assert.Contains(t, vpcsField.XMLTag, "vpcSet>item") // May have ,omitempty
}

func TestResolver_NonExistentShape(t *testing.T) {
	resolver, _ := setupResolver(t)

	_, _, err := resolver.ResolveShape("NonExistent")
	assert.Error(t, err)
}

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.amazonaws.ec2#Vpc", "Vpc"},
		{"smithy.api#String", "String"},
		{"String", "String"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ResolveTarget(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
