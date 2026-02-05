package typegen

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/smithy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Smithy model for generator tests
const generatorTestModel = `{
	"smithy": "2.0",
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.api#service": {
					"sdkId": "Test"
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
			"members": {}
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
						"smithy.api#xmlName": "cidrBlock"
					}
				},
				"State": {
					"target": "smithy.api#String",
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
		}
	}
}`

func createTestModelFile(t *testing.T) string {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.json")
	err := os.WriteFile(modelPath, []byte(generatorTestModel), 0o644)
	require.NoError(t, err)
	return modelPath
}

func TestGenerator_Generate(t *testing.T) {
	modelPath := createTestModelFile(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Check package declaration
	assert.Contains(t, code, "package test")

	// Check types are generated
	assert.Contains(t, code, "type DescribeVpcsResult struct")
	assert.Contains(t, code, "type Vpc struct")
	assert.Contains(t, code, "type Tag struct")

	// Check XML tags use camelCase
	assert.Contains(t, code, `xml:"vpcId`)
	assert.Contains(t, code, `xml:"cidrBlock`)
	assert.Contains(t, code, `xml:"state`)
	assert.Contains(t, code, `xml:"isDefault`)
	assert.Contains(t, code, `xml:"key`)
	assert.Contains(t, code, `xml:"value`)

	// Check list XML tags have >item syntax
	assert.Contains(t, code, `xml:"tagSet>item`)
	assert.Contains(t, code, `xml:"vpcSet>item`)

	// Should not contain PascalCase XML tags
	assert.NotContains(t, code, `xml:"VpcId"`)
	assert.NotContains(t, code, `xml:"CidrBlock"`)
}

func TestGenerator_GenerateWithSuffix(t *testing.T) {
	modelPath := createTestModelFile(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
		TypeSuffix:   "XML",
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Check types have suffix
	assert.Contains(t, code, "type DescribeVpcsResultXML struct")
	assert.Contains(t, code, "type VpcXML struct")
	assert.Contains(t, code, "type TagXML struct")

	// Check field types also have suffix
	assert.Contains(t, code, "[]VpcXML")
	assert.Contains(t, code, "[]TagXML")
}

func TestGenerator_GenerateToFile(t *testing.T) {
	modelPath := createTestModelFile(t)
	outputPath := filepath.Join(t.TempDir(), "output", "types.go")

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		OutputPath:   outputPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	err := generator.GenerateToFile()
	require.NoError(t, err)

	// Check file was created
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Check content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "type Vpc struct")
}

func TestCleanDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes HTML tags",
			input:    "<p>The ID of the VPC.</p>",
			expected: "The ID of the VPC.",
		},
		{
			name:     "converts code tags",
			input:    "Use <code>vpcId</code> to identify.",
			expected: "Use `vpcId` to identify.",
		},
		{
			name:     "handles entities",
			input:    "Values &amp; strings with &quot;quotes&quot;",
			expected: "Values & strings with \"quotes\"",
		},
		{
			name:     "truncates long text",
			input:    strings.Repeat("a", 200),
			expected: strings.Repeat("a", 97) + "...",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "removes nested tags",
			input:    "<p><b>Important:</b> This is <i>critical</i>.</p>",
			expected: "Important: This is critical.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanDocumentation(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPrimitiveType(t *testing.T) {
	primitives := []string{
		"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "Time",
	}

	for _, p := range primitives {
		assert.True(t, isPrimitiveType(p), "Expected %s to be primitive", p)
	}

	nonPrimitives := []string{
		"Vpc", "Tag", "Instance", "CustomType",
	}

	for _, np := range nonPrimitives {
		assert.False(t, isPrimitiveType(np), "Expected %s to not be primitive", np)
	}
}

// Test that generated types serialize correctly to XML
func TestGeneratedTypes_XMLSerialization(t *testing.T) {
	// Simulate the generated types
	type Tag struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	}

	type Vpc struct {
		VpcId     string `xml:"vpcId"`
		CidrBlock string `xml:"cidrBlock"`
		State     string `xml:"state"`
		IsDefault bool   `xml:"isDefault"`
		Tags      []Tag  `xml:"tagSet>item,omitempty"`
	}

	vpc := Vpc{
		VpcId:     "vpc-12345",
		CidrBlock: "10.0.0.0/16",
		State:     "available",
		IsDefault: false,
		Tags: []Tag{
			{Key: "Name", Value: "MyVpc"},
		},
	}

	data, err := xml.MarshalIndent(vpc, "", "  ")
	require.NoError(t, err)

	xmlStr := string(data)

	// Verify camelCase element names
	assert.Contains(t, xmlStr, "<vpcId>vpc-12345</vpcId>")
	assert.Contains(t, xmlStr, "<cidrBlock>10.0.0.0/16</cidrBlock>")
	assert.Contains(t, xmlStr, "<state>available</state>")
	assert.Contains(t, xmlStr, "<isDefault>false</isDefault>")

	// Verify list serialization
	assert.Contains(t, xmlStr, "<tagSet>")
	assert.Contains(t, xmlStr, "<item>")
	assert.Contains(t, xmlStr, "<key>Name</key>")
	assert.Contains(t, xmlStr, "<value>MyVpc</value>")

	// Should NOT contain PascalCase
	assert.NotContains(t, xmlStr, "<VpcId>")
	assert.NotContains(t, xmlStr, "<CidrBlock>")
	assert.NotContains(t, xmlStr, "<State>")
}

// Test deserialization from AWS-style XML
func TestGeneratedTypes_XMLDeserialization(t *testing.T) {
	type Tag struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	}

	type Vpc struct {
		VpcId     string `xml:"vpcId"`
		CidrBlock string `xml:"cidrBlock"`
		State     string `xml:"state"`
		IsDefault bool   `xml:"isDefault"`
		Tags      []Tag  `xml:"tagSet>item,omitempty"`
	}

	// AWS-style XML response (camelCase)
	awsXML := `<Vpc>
		<vpcId>vpc-67890</vpcId>
		<cidrBlock>172.16.0.0/16</cidrBlock>
		<state>pending</state>
		<isDefault>true</isDefault>
		<tagSet>
			<item>
				<key>Environment</key>
				<value>Production</value>
			</item>
		</tagSet>
	</Vpc>`

	var vpc Vpc
	err := xml.Unmarshal([]byte(awsXML), &vpc)
	require.NoError(t, err)

	assert.Equal(t, "vpc-67890", vpc.VpcId)
	assert.Equal(t, "172.16.0.0/16", vpc.CidrBlock)
	assert.Equal(t, "pending", vpc.State)
	assert.True(t, vpc.IsDefault)
	require.Len(t, vpc.Tags, 1)
	assert.Equal(t, "Environment", vpc.Tags[0].Key)
	assert.Equal(t, "Production", vpc.Tags[0].Value)
}

func TestGenerator_HeaderComments(t *testing.T) {
	modelPath := createTestModelFile(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Check header comments
	assert.Contains(t, code, "Code generated by CloudMirror")
	assert.Contains(t, code, "DO NOT EDIT")
	assert.Contains(t, code, "Service: test")
	assert.Contains(t, code, "Protocol: ec2")
}

// Test model with enums for pointer and enum tests
const generatorTestModelWithEnums = `{
	"smithy": "2.0",
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.api#service": { "sdkId": "Test" },
				"aws.protocols#ec2Query": {}
			}
		},
		"com.amazonaws.test#DescribeInstances": {
			"type": "operation",
			"input": { "target": "com.amazonaws.test#DescribeInstancesRequest" },
			"output": { "target": "com.amazonaws.test#DescribeInstancesResult" }
		},
		"com.amazonaws.test#DescribeInstancesRequest": {
			"type": "structure",
			"members": {}
		},
		"com.amazonaws.test#DescribeInstancesResult": {
			"type": "structure",
			"members": {
				"Instances": {
					"target": "com.amazonaws.test#InstanceList",
					"traits": { "smithy.api#xmlName": "instanceSet" }
				}
			},
			"traits": { "smithy.api#output": {} }
		},
		"com.amazonaws.test#Instance": {
			"type": "structure",
			"members": {
				"InstanceId": {
					"target": "smithy.api#String",
					"traits": { "smithy.api#xmlName": "instanceId" }
				},
				"InstanceType": {
					"target": "smithy.api#String",
					"traits": { "smithy.api#xmlName": "instanceType" }
				},
				"State": {
					"target": "com.amazonaws.test#InstanceState",
					"traits": { "smithy.api#xmlName": "state" }
				},
				"StateReason": {
					"target": "com.amazonaws.test#StateReason",
					"traits": { "smithy.api#xmlName": "stateReason" }
				},
				"LaunchTime": {
					"target": "smithy.api#Timestamp",
					"traits": { "smithy.api#xmlName": "launchTime" }
				},
				"CoreCount": {
					"target": "smithy.api#Integer",
					"traits": { "smithy.api#xmlName": "coreCount" }
				},
				"EbsOptimized": {
					"target": "smithy.api#Boolean",
					"traits": { "smithy.api#xmlName": "ebsOptimized" }
				},
				"Tags": {
					"target": "com.amazonaws.test#TagList",
					"traits": { "smithy.api#xmlName": "tagSet" }
				},
				"SecurityGroups": {
					"target": "com.amazonaws.test#GroupIdentifierList",
					"traits": { "smithy.api#xmlName": "groupSet" }
				},
				"StateHistory": {
					"target": "com.amazonaws.test#InstanceStateList",
					"traits": { "smithy.api#xmlName": "stateHistory" }
				}
			}
		},
		"com.amazonaws.test#InstanceState": {
			"type": "enum",
			"members": {
				"PENDING": { "target": "smithy.api#Unit", "traits": { "smithy.api#enumValue": "pending" } },
				"RUNNING": { "target": "smithy.api#Unit", "traits": { "smithy.api#enumValue": "running" } },
				"STOPPED": { "target": "smithy.api#Unit", "traits": { "smithy.api#enumValue": "stopped" } },
				"TERMINATED": { "target": "smithy.api#Unit", "traits": { "smithy.api#enumValue": "terminated" } }
			}
		},
		"com.amazonaws.test#StateReason": {
			"type": "structure",
			"members": {
				"Code": {
					"target": "smithy.api#String",
					"traits": { "smithy.api#xmlName": "code" }
				},
				"Message": {
					"target": "smithy.api#String",
					"traits": { "smithy.api#xmlName": "message" }
				}
			}
		},
		"com.amazonaws.test#Tag": {
			"type": "structure",
			"members": {
				"Key": { "target": "smithy.api#String", "traits": { "smithy.api#xmlName": "key" } },
				"Value": { "target": "smithy.api#String", "traits": { "smithy.api#xmlName": "value" } }
			}
		},
		"com.amazonaws.test#GroupIdentifier": {
			"type": "structure",
			"members": {
				"GroupId": { "target": "smithy.api#String", "traits": { "smithy.api#xmlName": "groupId" } },
				"GroupName": { "target": "smithy.api#String", "traits": { "smithy.api#xmlName": "groupName" } }
			}
		},
		"com.amazonaws.test#TagList": {
			"type": "list",
			"member": { "target": "com.amazonaws.test#Tag", "traits": { "smithy.api#xmlName": "item" } }
		},
		"com.amazonaws.test#GroupIdentifierList": {
			"type": "list",
			"member": { "target": "com.amazonaws.test#GroupIdentifier", "traits": { "smithy.api#xmlName": "item" } }
		},
		"com.amazonaws.test#InstanceList": {
			"type": "list",
			"member": { "target": "com.amazonaws.test#Instance", "traits": { "smithy.api#xmlName": "item" } }
		},
		"com.amazonaws.test#InstanceStateList": {
			"type": "list",
			"member": { "target": "com.amazonaws.test#InstanceState", "traits": { "smithy.api#xmlName": "item" } }
		}
	}
}`

func createTestModelFileWithEnums(t *testing.T) string {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model-enums.json")
	err := os.WriteFile(modelPath, []byte(generatorTestModelWithEnums), 0o644)
	require.NoError(t, err)
	return modelPath
}

// TestGenerator_EnumTypeAliases verifies that enum types generate type aliases
func TestGenerator_EnumTypeAliases(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify enum type alias is generated
	assert.Contains(t, code, "type InstanceState string", "Enum type alias should be generated")

	// Verify the type alias section exists
	assert.Contains(t, code, "// Enum type aliases", "Enum section header should exist")
}

// TestGenerator_PointerTypes verifies that primitive and struct fields use pointers
func TestGenerator_PointerTypes(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify primitive types use pointers
	assert.Contains(t, code, "*string", "String fields should use pointer")
	assert.Contains(t, code, "*int32", "Integer fields should use pointer")
	assert.Contains(t, code, "*bool", "Boolean fields should use pointer")
	assert.Contains(t, code, "*time.Time", "Timestamp fields should use pointer")

	// Verify nested struct types use pointers
	assert.Contains(t, code, "*StateReason", "Nested struct fields should use pointer")

	// Verify time import is present when time.Time is used
	assert.Contains(t, code, `"time"`, "time package should be imported")
}

// TestGenerator_NonPointerTypes verifies slices, maps, and enums don't use pointers
func TestGenerator_NonPointerTypes(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify slices don't use pointers (look for []Type not *[]Type)
	assert.Contains(t, code, "[]Tag", "Slice fields should not use pointer")
	assert.Contains(t, code, "[]GroupIdentifier", "Slice of struct fields should not use pointer")
	assert.NotContains(t, code, "*[]Tag", "Slice fields should not have pointer prefix")
	assert.NotContains(t, code, "*[]GroupIdentifier", "Slice of struct fields should not have pointer prefix")

	// Verify enum fields don't use pointers
	// The State field should be "State InstanceState" not "State *InstanceState"
	assert.Regexp(t, `State\s+InstanceState\s+`, code, "Enum fields should not use pointer")
	assert.NotContains(t, code, "*InstanceState", "Enum fields should not have pointer prefix")
}

// TestGenerator_EnumSlices verifies slices of enums work correctly
func TestGenerator_EnumSlices(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify slice of enum type (StateHistory []InstanceState)
	assert.Contains(t, code, "[]InstanceState", "Slice of enum should work")
	assert.NotContains(t, code, "*[]InstanceState", "Slice of enum should not use pointer")
}

// TestGenerator_EnumWithSuffix verifies enums don't get type suffix applied
func TestGenerator_EnumWithSuffix(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
		TypeSuffix:   "Output",
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify enum type alias does NOT have suffix
	assert.Contains(t, code, "type InstanceState string", "Enum type alias should NOT have suffix")
	assert.NotContains(t, code, "type InstanceStateOutput string", "Enum type alias should NOT have suffix")

	// Verify enum fields use the non-suffixed type
	assert.Regexp(t, `State\s+InstanceState\s+`, code, "Enum field should use non-suffixed type")
	assert.NotContains(t, code, "InstanceStateOutput", "Enum should not have suffix anywhere")

	// Verify struct types DO have suffix
	assert.Contains(t, code, "type InstanceOutput struct", "Struct types should have suffix")
	assert.Contains(t, code, "type StateReasonOutput struct", "Nested struct should have suffix")

	// Verify struct field references have suffix
	assert.Contains(t, code, "*StateReasonOutput", "Nested struct field should reference suffixed type")
	assert.Contains(t, code, "[]TagOutput", "Slice of struct should reference suffixed type")
}

// TestGenerator_EnumSliceWithSuffix verifies enum slices don't get suffix
func TestGenerator_EnumSliceWithSuffix(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
		TypeSuffix:   "Output",
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify slice of enum does NOT have suffix on element type
	assert.Contains(t, code, "[]InstanceState", "Slice of enum should NOT have suffix on element")
	assert.NotContains(t, code, "[]InstanceStateOutput", "Slice of enum element should NOT have suffix")
}

// TestGenerator_MixedPointerFields verifies a struct can have both pointer and non-pointer fields
func TestGenerator_MixedPointerFields(t *testing.T) {
	modelPath := createTestModelFileWithEnums(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Instance struct should have:
	// - Pointer fields: InstanceId (*string), CoreCount (*int32), StateReason (*StateReason)
	// - Non-pointer fields: State (InstanceState enum), Tags ([]Tag slice)

	// Verify Instance struct exists
	assert.Contains(t, code, "type Instance struct", "Instance struct should exist")

	// Check that the struct has the expected mixed pointer/non-pointer pattern
	// This regex checks for the pattern within the Instance struct
	assert.Regexp(t, `InstanceId\s+\*string`, code, "InstanceId should be *string")
	assert.Regexp(t, `CoreCount\s+\*int32`, code, "CoreCount should be *int32")
	assert.Regexp(t, `StateReason\s+\*StateReason`, code, "StateReason should be *StateReason")
	assert.Regexp(t, `State\s+InstanceState\s+`, code, "State should be InstanceState (no pointer)")
	assert.Regexp(t, `Tags\s+\[\]Tag`, code, "Tags should be []Tag (no pointer)")
}

// TestShouldUsePointer verifies the pointer determination logic
func TestShouldUsePointer(t *testing.T) {
	tests := []struct {
		name    string
		goType  string
		isEnum  bool
		wantPtr bool
	}{
		{"string primitive", "string", false, true},
		{"int32 primitive", "int32", false, true},
		{"bool primitive", "bool", false, true},
		{"time.Time", "time.Time", false, true},
		{"nested struct", "StateReason", false, true},
		{"slice of struct", "[]Tag", false, false},
		{"slice of string", "[]string", false, false},
		{"map", "map[string]string", false, false},
		{"enum type", "InstanceState", true, false},
		{"slice of enum", "[]InstanceState", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldUsePointer(tc.goType, tc.isEnum)
			assert.Equal(t, tc.wantPtr, result, "shouldUsePointer(%q, %v)", tc.goType, tc.isEnum)
		})
	}
}

// Test model with validation traits
const generatorTestModelWithValidation = `{
	"smithy": "2.0",
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.api#service": { "sdkId": "Test" },
				"aws.protocols#ec2Query": {}
			}
		},
		"com.amazonaws.test#CreateUser": {
			"type": "operation",
			"input": { "target": "com.amazonaws.test#CreateUserRequest" },
			"output": { "target": "com.amazonaws.test#CreateUserResult" }
		},
		"com.amazonaws.test#CreateUserRequest": {
			"type": "structure",
			"members": {},
			"traits": { "smithy.api#input": {} }
		},
		"com.amazonaws.test#CreateUserResult": {
			"type": "structure",
			"members": {
				"User": {
					"target": "com.amazonaws.test#User",
					"traits": { "smithy.api#xmlName": "user" }
				}
			},
			"traits": { "smithy.api#output": {} }
		},
		"com.amazonaws.test#User": {
			"type": "structure",
			"members": {
				"Username": {
					"target": "com.amazonaws.test#UsernameType",
					"traits": {
						"smithy.api#xmlName": "username",
						"smithy.api#required": {}
					}
				},
				"Email": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "email",
						"smithy.api#pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+$"
					}
				},
				"Age": {
					"target": "smithy.api#Integer",
					"traits": {
						"smithy.api#xmlName": "age",
						"smithy.api#range": { "min": 0, "max": 150 }
					}
				},
				"Tags": {
					"target": "com.amazonaws.test#TagList",
					"traits": {
						"smithy.api#xmlName": "tags",
						"smithy.api#length": { "min": 0, "max": 50 }
					}
				}
			}
		},
		"com.amazonaws.test#UsernameType": {
			"type": "string",
			"traits": {
				"smithy.api#length": { "min": 3, "max": 64 },
				"smithy.api#pattern": "^[a-zA-Z][a-zA-Z0-9_-]*$"
			}
		},
		"com.amazonaws.test#Tag": {
			"type": "structure",
			"members": {
				"Key": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "key",
						"smithy.api#length": { "min": 1, "max": 128 }
					}
				},
				"Value": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#xmlName": "value",
						"smithy.api#length": { "max": 256 }
					}
				}
			}
		},
		"com.amazonaws.test#TagList": {
			"type": "list",
			"member": {
				"target": "com.amazonaws.test#Tag",
				"traits": { "smithy.api#xmlName": "item" }
			}
		}
	}
}`

func createTestModelFileWithValidation(t *testing.T) string {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model-validation.json")
	err := os.WriteFile(modelPath, []byte(generatorTestModelWithValidation), 0o644)
	require.NoError(t, err)
	return modelPath
}

// TestGenerator_ValidationConstraints verifies that validation traits generate doc comments and Validate methods
func TestGenerator_ValidationConstraints(t *testing.T) {
	modelPath := createTestModelFileWithValidation(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Check validation doc comments are present
	assert.Contains(t, code, "[Length: 3-64", "Length constraint should be in doc comment")
	assert.Contains(t, code, "[Pattern:", "Pattern constraint should be in doc comment")
	assert.Contains(t, code, "[Range: 0-150]", "Range constraint should be in doc comment")

	// Check Validate() methods are generated
	assert.Contains(t, code, "func (s *User) Validate() []error", "Validate method should be generated for User")
	assert.Contains(t, code, "func (s *Tag) Validate() []error", "Validate method should be generated for Tag")

	// Check imports
	assert.Contains(t, code, `"fmt"`, "fmt should be imported for validation errors")
	assert.Contains(t, code, `"regexp"`, "regexp should be imported for pattern validation")

	// Check ValidationError type
	assert.Contains(t, code, "type ValidationError struct", "ValidationError type should be defined")
}

// TestGenerator_ValidationLogic verifies the generated validation logic patterns
func TestGenerator_ValidationLogic(t *testing.T) {
	modelPath := createTestModelFileWithValidation(t)

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Check length validation logic
	assert.Contains(t, code, "len(*s.", "Length validation should dereference pointers")
	assert.Contains(t, code, "len(s.", "Length validation for slices")

	// Check pattern validation logic
	assert.Contains(t, code, "MatchString", "Pattern validation should use MatchString")
	assert.Contains(t, code, "validate", "Pattern variable should be prefixed with validate")

	// Check range validation logic
	assert.Contains(t, code, "float64(", "Range validation should convert to float64")
}

// TestGenerator_NoValidationWhenNoConstraints verifies types without constraints don't have Validate methods
func TestGenerator_NoValidationWhenNoConstraints(t *testing.T) {
	modelPath := createTestModelFile(t) // Original model without validation

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Validate method should not be generated for types without constraints
	assert.NotContains(t, code, "func (s *Vpc) Validate()", "Validate should not be generated without constraints")
	assert.NotContains(t, code, "ValidationError", "ValidationError should not be present without validation")
}

// TestGenerator_ValidationImportsOnlyWhenNeeded verifies imports are only added when validation is present
func TestGenerator_ValidationImportsOnlyWhenNeeded(t *testing.T) {
	modelPath := createTestModelFile(t) // Original model without validation

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// fmt and regexp should not be imported without validation
	assert.NotContains(t, code, `"fmt"`, "fmt should not be imported without validation")
	assert.NotContains(t, code, `"regexp"`, "regexp should not be imported without validation")
}

// TestFormatValidationComment verifies validation comment formatting
func TestFormatValidationComment(t *testing.T) {
	tests := []struct {
		name     string
		traits   smithy.ValidationTraits
		expected string
	}{
		{
			name:     "empty traits",
			traits:   smithy.ValidationTraits{},
			expected: "",
		},
		{
			name:     "length min and max",
			traits:   smithy.ValidationTraits{LengthMin: ptrInt64(1), LengthMax: ptrInt64(100)},
			expected: "[Length: 1-100]",
		},
		{
			name:     "length min only",
			traits:   smithy.ValidationTraits{LengthMin: ptrInt64(1)},
			expected: "[Min length: 1]",
		},
		{
			name:     "length max only",
			traits:   smithy.ValidationTraits{LengthMax: ptrInt64(256)},
			expected: "[Max length: 256]",
		},
		{
			name:     "pattern",
			traits:   smithy.ValidationTraits{Pattern: "^[a-z]+$"},
			expected: "[Pattern: ^[a-z]+$]",
		},
		{
			name:     "range min and max",
			traits:   smithy.ValidationTraits{RangeMin: ptrFloat64(0), RangeMax: ptrFloat64(100)},
			expected: "[Range: 0-100]",
		},
		{
			name: "combined constraints",
			traits: smithy.ValidationTraits{
				LengthMin: ptrInt64(1),
				LengthMax: ptrInt64(50),
				Pattern:   "^[a-z]+$",
			},
			expected: "[Length: 1-50, Pattern: ^[a-z]+$]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatValidationComment(tc.traits)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsNumericType verifies numeric type detection
func TestIsNumericType(t *testing.T) {
	numericTypes := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
	}

	for _, numType := range numericTypes {
		assert.True(t, isNumericType(numType), "Expected %s to be numeric", numType)
	}

	nonNumericTypes := []string{
		"string", "bool", "time.Time", "Vpc", "[]int", "map[string]int",
	}

	for _, nonNumType := range nonNumericTypes {
		assert.False(t, isNumericType(nonNumType), "Expected %s to not be numeric", nonNumType)
	}
}

// Helper functions for test pointer creation
func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

// Test model with HTTP location traits (REST-JSON protocol)
const generatorTestModelRESTJSON = `{
	"smithy": "2.0",
	"shapes": {
		"com.amazonaws.test#TestService": {
			"type": "service",
			"traits": {
				"aws.api#service": { "sdkId": "Test" },
				"aws.protocols#restJson1": {}
			}
		},
		"com.amazonaws.test#Invoke": {
			"type": "operation",
			"input": { "target": "com.amazonaws.test#InvokeRequest" },
			"output": { "target": "com.amazonaws.test#InvokeResponse" },
			"traits": {
				"smithy.api#http": { "method": "POST", "uri": "/functions/{FunctionName}/invocations" }
			}
		},
		"com.amazonaws.test#InvokeRequest": {
			"type": "structure",
			"members": {
				"FunctionName": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#httpLabel": {},
						"smithy.api#required": {}
					}
				},
				"InvocationType": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#httpHeader": "X-Amz-Invocation-Type"
					}
				},
				"Qualifier": {
					"target": "smithy.api#String",
					"traits": {
						"smithy.api#httpQuery": "Qualifier"
					}
				},
				"Payload": {
					"target": "smithy.api#Blob",
					"traits": {
						"smithy.api#httpPayload": {}
					}
				}
			},
			"traits": { "smithy.api#input": {} }
		},
		"com.amazonaws.test#InvokeResponse": {
			"type": "structure",
			"members": {
				"StatusCode": {
					"target": "smithy.api#Integer",
					"traits": {
						"smithy.api#httpResponseCode": {}
					}
				},
				"ResponsePayload": {
					"target": "smithy.api#Blob",
					"traits": {
						"smithy.api#httpPayload": {}
					}
				}
			},
			"traits": { "smithy.api#output": {} }
		}
	}
}`

func createTestModelFileWithHTTP(t *testing.T) string {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model-http.json")
	err := os.WriteFile(modelPath, []byte(generatorTestModelRESTJSON), 0o644)
	require.NoError(t, err)
	return modelPath
}

func TestGenerator_HTTPLocationTags_RESTJSON(t *testing.T) {
	modelPath := createTestModelFileWithHTTP(t)

	config := &Config{
		ServiceName:   "test",
		PackageName:   "test",
		ModelPath:     modelPath,
		IncludeInputs: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Verify HTTP location tags are present in InvokeRequest
	assert.Contains(t, code, `uri:"FunctionName"`, "httpLabel should generate uri tag")
	assert.Contains(t, code, `header:"X-Amz-Invocation-Type"`, "httpHeader should generate header tag")
	assert.Contains(t, code, `query:"Qualifier"`, "httpQuery should generate query tag")
	assert.Contains(t, code, `payload:"true"`, "httpPayload should generate payload tag")

	// Verify payload fields have json:"-" to exclude from JSON serialization
	assert.Contains(t, code, `json:"-"`, "payload fields should have json:\"-\"")
}

func TestGenerator_NoHTTPTags_QueryProtocol(t *testing.T) {
	// Query protocol should NOT have HTTP location tags
	modelPath := createTestModelFile(t) // Uses ec2Query protocol

	config := &Config{
		ServiceName:  "test",
		PackageName:  "test",
		ModelPath:    modelPath,
		ResponseOnly: true,
	}

	generator := NewGenerator(config)
	code, err := generator.Generate()
	require.NoError(t, err)

	// Should NOT contain HTTP location tags
	assert.NotContains(t, code, `uri:"`, "Query protocol should not have uri tags")
	assert.NotContains(t, code, `header:"`, "Query protocol should not have header tags")
	assert.NotContains(t, code, `query:"`, "Query protocol should not have query tags")
	assert.NotContains(t, code, `payload:"`, "Query protocol should not have payload tags")
}
