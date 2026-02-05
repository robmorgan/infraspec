package plan

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePlanFile(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)

	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, "1.2", plan.FormatVersion)
	assert.Equal(t, "1.7.0", plan.TerraformVersion)
	assert.NotNil(t, plan.PlannedValues)
	assert.NotNil(t, plan.ResourceChanges)
	assert.NotNil(t, plan.Configuration)
}

func TestParsePlanFile_NotFound(t *testing.T) {
	_, err := ParsePlanFile("nonexistent/path/plan.json")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read plan file")
}

func TestParsePlanBytes_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"format_version": "1.0", invalid}`)
	_, err := ParsePlanBytes(invalidJSON)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse plan JSON")
}

func TestParsePlanBytes_EmptyJSON(t *testing.T) {
	emptyJSON := []byte(`{}`)
	plan, err := ParsePlanBytes(emptyJSON)

	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.Empty(t, plan.FormatVersion)
	assert.Nil(t, plan.ResourceChanges)
}

func TestResourcesByType(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	tests := []struct {
		name         string
		resourceType string
		expectedLen  int
	}{
		{"VPC resources", "aws_vpc", 1},
		{"Subnet resources", "aws_subnet", 2},
		{"Security group resources", "aws_security_group", 1},
		{"Internet gateway resources", "aws_internet_gateway", 1},
		{"Availability zones data source", "aws_availability_zones", 1},
		{"Nonexistent type", "aws_nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := plan.ResourcesByType(tt.resourceType)
			assert.Len(t, resources, tt.expectedLen)
		})
	}
}

func TestResourceByAddress(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	tests := []struct {
		name     string
		address  string
		found    bool
		expected string
	}{
		{"Find VPC", "aws_vpc.main", true, "aws_vpc"},
		{"Find public subnet", "aws_subnet.public", true, "aws_subnet"},
		{"Find private subnet", "aws_subnet.private", true, "aws_subnet"},
		{"Find security group", "aws_security_group.ssh", true, "aws_security_group"},
		{"Find data source", "data.aws_availability_zones.available", true, "aws_availability_zones"},
		{"Nonexistent address", "aws_vpc.nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := plan.ResourceByAddress(tt.address)
			if tt.found {
				require.NotNil(t, rc)
				assert.Equal(t, tt.expected, rc.Type)
			} else {
				assert.Nil(t, rc)
			}
		})
	}
}

func TestResourcesByModule(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	// All resources in vpc_basic.json are in the root module (empty ModuleAddress)
	rootResources := plan.ResourcesByModule("")
	assert.Len(t, rootResources, 6) // vpc, 2 subnets, sg, igw, data source

	// No resources in a non-existent module
	moduleResources := plan.ResourcesByModule("module.nonexistent")
	assert.Len(t, moduleResources, 0)
}

func TestGetAfter(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)

	// Test string value
	cidr, ok := vpc.GetAfter("cidr_block")
	assert.True(t, ok)
	assert.Equal(t, "10.0.0.0/16", cidr)

	// Test bool value
	dnsSupport, ok := vpc.GetAfter("enable_dns_support")
	assert.True(t, ok)
	assert.Equal(t, true, dnsSupport)

	// Test nonexistent key
	_, ok = vpc.GetAfter("nonexistent_key")
	assert.False(t, ok)
}

func TestGetAfterNested(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	tests := []struct {
		name     string
		address  string
		path     []string
		expected interface{}
		found    bool
	}{
		{
			name:     "VPC tag Name",
			address:  "aws_vpc.main",
			path:     []string{"tags", "Name"},
			expected: "main-vpc",
			found:    true,
		},
		{
			name:     "VPC tag Environment",
			address:  "aws_vpc.main",
			path:     []string{"tags", "Environment"},
			expected: "production",
			found:    true,
		},
		{
			name:     "Security group ingress from_port",
			address:  "aws_security_group.ssh",
			path:     []string{"ingress", "0", "from_port"},
			expected: float64(22), // JSON numbers are float64
			found:    true,
		},
		{
			name:     "Security group ingress to_port",
			address:  "aws_security_group.ssh",
			path:     []string{"ingress", "0", "to_port"},
			expected: float64(22),
			found:    true,
		},
		{
			name:     "Security group ingress protocol",
			address:  "aws_security_group.ssh",
			path:     []string{"ingress", "0", "protocol"},
			expected: "tcp",
			found:    true,
		},
		{
			name:     "Security group ingress cidr_blocks first",
			address:  "aws_security_group.ssh",
			path:     []string{"ingress", "0", "cidr_blocks", "0"},
			expected: "0.0.0.0/0",
			found:    true,
		},
		{
			name:     "Nonexistent nested path",
			address:  "aws_vpc.main",
			path:     []string{"tags", "Nonexistent"},
			expected: nil,
			found:    false,
		},
		{
			name:     "Invalid array index",
			address:  "aws_security_group.ssh",
			path:     []string{"ingress", "99"},
			expected: nil,
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := plan.ResourceByAddress(tt.address)
			require.NotNil(t, rc)

			val, ok := rc.GetAfterNested(tt.path...)
			assert.Equal(t, tt.found, ok)
			if tt.found {
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestGetAfterNested_EmptyPath(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)

	_, ok := vpc.GetAfterNested()
	assert.False(t, ok)
}

func TestGetBefore(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)

	// Before is null for create actions
	_, ok := vpc.GetBefore("cidr_block")
	assert.False(t, ok)
}

func TestGetBeforeNested(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)

	// Before is null for create actions
	_, ok := vpc.GetBeforeNested("tags", "Name")
	assert.False(t, ok)
}

func TestActionDetection(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	// Test create action
	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)
	assert.True(t, vpc.IsCreate())
	assert.False(t, vpc.IsUpdate())
	assert.False(t, vpc.IsDelete())
	assert.False(t, vpc.IsReplace())
	assert.False(t, vpc.IsNoOp())
	assert.False(t, vpc.IsRead())

	// Test read action (data source)
	dataSource := plan.ResourceByAddress("data.aws_availability_zones.available")
	require.NotNil(t, dataSource)
	assert.False(t, dataSource.IsCreate())
	assert.False(t, dataSource.IsUpdate())
	assert.False(t, dataSource.IsDelete())
	assert.False(t, dataSource.IsReplace())
	assert.False(t, dataSource.IsNoOp())
	assert.True(t, dataSource.IsRead())
}

func TestActionDetection_SyntheticCases(t *testing.T) {
	tests := []struct {
		name      string
		actions   []Action
		isCreate  bool
		isUpdate  bool
		isDelete  bool
		isReplace bool
		isNoOp    bool
		isRead    bool
	}{
		{
			name:     "Create only",
			actions:  []Action{ActionCreate},
			isCreate: true,
		},
		{
			name:     "Update only",
			actions:  []Action{ActionUpdate},
			isUpdate: true,
		},
		{
			name:     "Delete only",
			actions:  []Action{ActionDelete},
			isDelete: true,
		},
		{
			name:      "Replace (delete then create)",
			actions:   []Action{ActionDelete, ActionCreate},
			isReplace: true,
		},
		{
			name:      "Replace (create then delete)",
			actions:   []Action{ActionCreate, ActionDelete},
			isReplace: true,
		},
		{
			name:    "No-op",
			actions: []Action{ActionNoOp},
			isNoOp:  true,
		},
		{
			name:    "Read",
			actions: []Action{ActionRead},
			isRead:  true,
		},
		{
			name:    "Empty actions",
			actions: []Action{},
			isNoOp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &ResourceChange{
				Change: &Change{
					Actions: tt.actions,
				},
			}
			assert.Equal(t, tt.isCreate, rc.IsCreate(), "IsCreate")
			assert.Equal(t, tt.isUpdate, rc.IsUpdate(), "IsUpdate")
			assert.Equal(t, tt.isDelete, rc.IsDelete(), "IsDelete")
			assert.Equal(t, tt.isReplace, rc.IsReplace(), "IsReplace")
			assert.Equal(t, tt.isNoOp, rc.IsNoOp(), "IsNoOp")
			assert.Equal(t, tt.isRead, rc.IsRead(), "IsRead")
		})
	}
}

func TestGetAfterTypedHelpers(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	t.Run("GetAfterString", func(t *testing.T) {
		vpc := plan.ResourceByAddress("aws_vpc.main")
		require.NotNil(t, vpc)

		assert.Equal(t, "10.0.0.0/16", vpc.GetAfterString("cidr_block"))
		assert.Equal(t, "default", vpc.GetAfterString("instance_tenancy"))
		assert.Equal(t, "", vpc.GetAfterString("nonexistent"))
		// Bool value should return empty string
		assert.Equal(t, "", vpc.GetAfterString("enable_dns_support"))
	})

	t.Run("GetAfterBool", func(t *testing.T) {
		vpc := plan.ResourceByAddress("aws_vpc.main")
		require.NotNil(t, vpc)

		assert.True(t, vpc.GetAfterBool("enable_dns_support"))
		assert.True(t, vpc.GetAfterBool("enable_dns_hostnames"))
		assert.False(t, vpc.GetAfterBool("nonexistent"))
		// String value should return false
		assert.False(t, vpc.GetAfterBool("cidr_block"))
	})

	t.Run("GetAfterInt", func(t *testing.T) {
		sg := plan.ResourceByAddress("aws_security_group.ssh")
		require.NotNil(t, sg)

		// Get from nested for this test
		ingressVal, ok := sg.GetAfterNested("ingress", "0", "from_port")
		require.True(t, ok)
		assert.Equal(t, float64(22), ingressVal)

		// Test with a simple case using constructed resource
		rc := &ResourceChange{
			Change: &Change{
				After: map[string]interface{}{
					"port":  float64(443),
					"count": float64(10),
				},
			},
		}
		assert.Equal(t, 443, rc.GetAfterInt("port"))
		assert.Equal(t, 10, rc.GetAfterInt("count"))
		assert.Equal(t, 0, rc.GetAfterInt("nonexistent"))
	})

	t.Run("GetAfterFloat", func(t *testing.T) {
		rc := &ResourceChange{
			Change: &Change{
				After: map[string]interface{}{
					"ratio": float64(3.14),
				},
			},
		}
		assert.Equal(t, 3.14, rc.GetAfterFloat("ratio"))
		assert.Equal(t, float64(0), rc.GetAfterFloat("nonexistent"))
	})

	t.Run("GetAfterStringSlice", func(t *testing.T) {
		rc := &ResourceChange{
			Change: &Change{
				After: map[string]interface{}{
					"security_groups": []interface{}{"sg-1", "sg-2", "sg-3"},
					"mixed":           []interface{}{"str", 123, true},
				},
			},
		}
		slice := rc.GetAfterStringSlice("security_groups")
		assert.Equal(t, []string{"sg-1", "sg-2", "sg-3"}, slice)

		// Mixed types - only strings are included
		mixedSlice := rc.GetAfterStringSlice("mixed")
		assert.Equal(t, []string{"str"}, mixedSlice)

		assert.Nil(t, rc.GetAfterStringSlice("nonexistent"))
	})

	t.Run("GetAfterMap", func(t *testing.T) {
		vpc := plan.ResourceByAddress("aws_vpc.main")
		require.NotNil(t, vpc)

		tags := vpc.GetAfterMap("tags")
		require.NotNil(t, tags)
		assert.Equal(t, "main-vpc", tags["Name"])
		assert.Equal(t, "production", tags["Environment"])

		assert.Nil(t, vpc.GetAfterMap("nonexistent"))
	})
}

func TestResourceChangeWithNilChange(t *testing.T) {
	rc := &ResourceChange{
		Address: "aws_vpc.test",
		Type:    "aws_vpc",
		Name:    "test",
		Change:  nil,
	}

	// All these should handle nil gracefully
	_, ok := rc.GetAfter("key")
	assert.False(t, ok)

	_, ok = rc.GetAfterNested("path", "to", "value")
	assert.False(t, ok)

	_, ok = rc.GetBefore("key")
	assert.False(t, ok)

	_, ok = rc.GetBeforeNested("path", "to", "value")
	assert.False(t, ok)

	assert.Equal(t, "", rc.GetAfterString("key"))
	assert.Equal(t, false, rc.GetAfterBool("key"))
	assert.Equal(t, 0, rc.GetAfterInt("key"))
	assert.Equal(t, float64(0), rc.GetAfterFloat("key"))
	assert.Nil(t, rc.GetAfterStringSlice("key"))
	assert.Nil(t, rc.GetAfterMap("key"))

	assert.False(t, rc.IsCreate())
	assert.False(t, rc.IsUpdate())
	assert.False(t, rc.IsDelete())
	assert.False(t, rc.IsReplace())
	assert.True(t, rc.IsNoOp())
	assert.False(t, rc.IsRead())
	assert.False(t, rc.HasAction(ActionCreate))
}

func TestHasAction(t *testing.T) {
	rc := &ResourceChange{
		Change: &Change{
			Actions: []Action{ActionDelete, ActionCreate},
		},
	}

	assert.True(t, rc.HasAction(ActionCreate))
	assert.True(t, rc.HasAction(ActionDelete))
	assert.False(t, rc.HasAction(ActionUpdate))
	assert.False(t, rc.HasAction(ActionNoOp))
	assert.False(t, rc.HasAction(ActionRead))
}

func TestIsDataSource(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	dataSource := plan.ResourceByAddress("data.aws_availability_zones.available")
	require.NotNil(t, dataSource)
	assert.True(t, dataSource.IsDataSource())
	assert.False(t, dataSource.IsManaged())

	vpc := plan.ResourceByAddress("aws_vpc.main")
	require.NotNil(t, vpc)
	assert.False(t, vpc.IsDataSource())
	assert.True(t, vpc.IsManaged())
}

func TestVariables(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	require.NotNil(t, plan.Variables)
	assert.Len(t, plan.Variables, 2)

	envVar, ok := plan.Variables["environment"]
	require.True(t, ok)
	assert.Equal(t, "production", envVar.Value)

	cidrVar, ok := plan.Variables["vpc_cidr"]
	require.True(t, ok)
	assert.Equal(t, "10.0.0.0/16", cidrVar.Value)
}

func TestPlannedValues(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	require.NotNil(t, plan.PlannedValues)
	require.NotNil(t, plan.PlannedValues.Outputs)

	vpcOutput, ok := plan.PlannedValues.Outputs["vpc_id"]
	require.True(t, ok)
	assert.Equal(t, "vpc-12345678", vpcOutput.Value)
	assert.False(t, vpcOutput.Sensitive)

	require.NotNil(t, plan.PlannedValues.RootModule)
	assert.Len(t, plan.PlannedValues.RootModule.Resources, 3) // vpc, 2 subnets in planned_values
}

func TestConfiguration(t *testing.T) {
	planPath := filepath.Join("testdata", "plans", "vpc_basic.json")
	plan, err := ParsePlanFile(planPath)
	require.NoError(t, err)

	require.NotNil(t, plan.Configuration)
	require.NotNil(t, plan.Configuration.ProviderConfig)

	awsProvider, ok := plan.Configuration.ProviderConfig["aws"]
	require.True(t, ok)
	assert.Equal(t, "aws", awsProvider.Name)
	assert.Equal(t, "registry.terraform.io/hashicorp/aws", awsProvider.FullName)

	require.NotNil(t, plan.Configuration.RootModule)
	require.NotNil(t, plan.Configuration.RootModule.Variables)

	envVar, ok := plan.Configuration.RootModule.Variables["environment"]
	require.True(t, ok)
	assert.Equal(t, "development", envVar.Default)
	assert.Equal(t, "Environment name", envVar.Description)
}

func TestFullType(t *testing.T) {
	tests := []struct {
		name         string
		rc           *ResourceChange
		expectedType string
	}{
		{
			name: "With provider and underscore",
			rc: &ResourceChange{
				Type:         "aws_vpc",
				ProviderName: "registry.terraform.io/hashicorp/aws",
			},
			expectedType: "aws_vpc",
		},
		{
			name: "Without underscore",
			rc: &ResourceChange{
				Type:         "vpc",
				ProviderName: "registry.terraform.io/hashicorp/aws",
			},
			expectedType: "registry.terraform.io/hashicorp/aws_vpc",
		},
		{
			name: "Without provider",
			rc: &ResourceChange{
				Type:         "aws_vpc",
				ProviderName: "",
			},
			expectedType: "aws_vpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedType, tt.rc.FullType())
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"float64", float64(55.9), 55},
		{"string valid", "123", 123},
		{"string invalid", "not a number", 0},
		{"bool", true, 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
