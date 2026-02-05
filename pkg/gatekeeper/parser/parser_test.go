package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test Terraform file
	tfContent := `
resource "aws_s3_bucket" "example" {
  bucket = "my-bucket"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id

  versioning_configuration {
    status = "Enabled"
  }
}
`
	tfPath := filepath.Join(tmpDir, "main.tf")
	err := os.WriteFile(tfPath, []byte(tfContent), 0o644)
	require.NoError(t, err)

	// Parse the file
	p := New(Config{})
	resources, err := p.ParseFile(tfPath)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	// Check first resource
	s3Bucket := resources[0]
	assert.Equal(t, "aws_s3_bucket", s3Bucket.Type)
	assert.Equal(t, "example", s3Bucket.Name)
	assert.Equal(t, "my-bucket", s3Bucket.Attributes["bucket"])
	assert.Equal(t, tfPath, s3Bucket.Location.File)
	assert.Greater(t, s3Bucket.Location.Line, 0)

	// Check tags
	tags, ok := s3Bucket.Attributes["tags"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "My bucket", tags["Name"])
	assert.Equal(t, "Dev", tags["Environment"])

	// Check second resource with nested block
	versioning := resources[1]
	assert.Equal(t, "aws_s3_bucket_versioning", versioning.Type)
	assert.Equal(t, "example", versioning.Name)

	versioningConfig, ok := versioning.Attributes["versioning_configuration"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Enabled", versioningConfig["status"])
}

func TestParser_ParseFile_WithVariables(t *testing.T) {
	tmpDir := t.TempDir()

	// Create variables file
	varsContent := `
variable "bucket_name" {
  default = "default-bucket"
}

variable "environment" {
  type = string
}
`
	varsPath := filepath.Join(tmpDir, "variables.tf")
	err := os.WriteFile(varsPath, []byte(varsContent), 0o644)
	require.NoError(t, err)

	// Create main file using variables
	mainContent := `
resource "aws_s3_bucket" "example" {
  bucket = var.bucket_name

  tags = {
    Environment = var.environment
  }
}
`
	mainPath := filepath.Join(tmpDir, "main.tf")
	err = os.WriteFile(mainPath, []byte(mainContent), 0o644)
	require.NoError(t, err)

	// Parse - first collect variables
	p := New(Config{})
	err = p.collectVariables(varsPath, []byte(varsContent))
	require.NoError(t, err)

	resources, err := p.ParseFile(mainPath)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	// Check that variable with default was resolved
	bucket := resources[0]
	assert.Equal(t, "default-bucket", bucket.Attributes["bucket"])

	// Check that variable without default is unknown
	tags, ok := bucket.Attributes["tags"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, IsUnknown(tags["Environment"]))
}

func TestParser_ParseFile_WithTfvars(t *testing.T) {
	tmpDir := t.TempDir()

	// Create variables file
	varsContent := `
variable "bucket_name" {
  default = "default-bucket"
}

variable "environment" {
  type = string
}
`
	varsPath := filepath.Join(tmpDir, "variables.tf")
	err := os.WriteFile(varsPath, []byte(varsContent), 0o644)
	require.NoError(t, err)

	// Create tfvars file
	tfvarsContent := `
bucket_name = "custom-bucket"
environment = "production"
`
	tfvarsPath := filepath.Join(tmpDir, "terraform.tfvars")
	err = os.WriteFile(tfvarsPath, []byte(tfvarsContent), 0o644)
	require.NoError(t, err)

	// Create main file
	mainContent := `
resource "aws_s3_bucket" "example" {
  bucket = var.bucket_name

  tags = {
    Environment = var.environment
  }
}
`
	mainPath := filepath.Join(tmpDir, "main.tf")
	err = os.WriteFile(mainPath, []byte(mainContent), 0o644)
	require.NoError(t, err)

	// Parse with tfvars
	p := New(Config{VarFile: tfvarsPath})
	err = p.collectVariables(varsPath, []byte(varsContent))
	require.NoError(t, err)
	err = p.loadTfvars(tfvarsPath)
	require.NoError(t, err)

	resources, err := p.ParseFile(mainPath)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	// Check that tfvars overrides default
	bucket := resources[0]
	assert.Equal(t, "custom-bucket", bucket.Attributes["bucket"])

	// Check that tfvars provides value for variable without default
	tags, ok := bucket.Attributes["tags"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "production", tags["Environment"])
}

func TestParser_ParseFile_WithNestedBlocks(t *testing.T) {
	tmpDir := t.TempDir()

	tfContent := `
resource "aws_security_group" "example" {
  name        = "example"
  description = "Example security group"
  vpc_id      = "vpc-123"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`
	tfPath := filepath.Join(tmpDir, "main.tf")
	err := os.WriteFile(tfPath, []byte(tfContent), 0o644)
	require.NoError(t, err)

	p := New(Config{})
	resources, err := p.ParseFile(tfPath)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	sg := resources[0]
	assert.Equal(t, "aws_security_group", sg.Type)
	assert.Equal(t, "example", sg.Name)
	assert.Equal(t, "example", sg.Attributes["name"])

	// Check that multiple ingress blocks are collected as an array
	ingress, ok := sg.Attributes["ingress"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ingress, 2)

	// Check first ingress
	ing1, ok := ingress[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 22, ing1["from_port"])

	// Check second ingress
	ing2, ok := ingress[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 443, ing2["from_port"])

	// Check egress (single block)
	egress, ok := sg.Attributes["egress"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, egress["from_port"])
}

func TestGetAttribute(t *testing.T) {
	attrs := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep",
			},
		},
		"array": []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 2},
			map[string]interface{}{"id": 3},
		},
		"number": 42,
		"bool":   true,
	}

	tests := []struct {
		path     string
		expected interface{}
		exists   bool
	}{
		{"simple", "value", true},
		{"nested.level1.level2", "deep", true},
		{"nested.level1", map[string]interface{}{"level2": "deep"}, true},
		{"array[0].id", 1, true},
		{"array[1].id", 2, true},
		{"array[*].id", []interface{}{1, 2, 3}, true},
		{"number", 42, true},
		{"bool", true, true},
		{"missing", nil, false},
		{"nested.missing", nil, false},
		{"array[10].id", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			val, exists := GetAttribute(attrs, tt.path)
			assert.Equal(t, tt.exists, exists)
			if tt.exists {
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"simple", []string{"simple"}},
		{"nested.path", []string{"nested", "path"}},
		{"a.b.c", []string{"a", "b", "c"}},
		{"array[0]", []string{"array", "[0]"}},
		{"array[*].field", []string{"array", "[*]", "field"}},
		{"a.b[0].c[*].d", []string{"a", "b", "[0]", "c", "[*]", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := splitPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnknownAndComputed(t *testing.T) {
	assert.True(t, IsUnknown(UnknownValue{}))
	assert.False(t, IsUnknown("hello"))
	assert.False(t, IsUnknown(42))

	assert.True(t, IsComputed(ComputedValue{}))
	assert.False(t, IsComputed("hello"))
	assert.False(t, IsComputed(42))

	assert.Equal(t, "<unknown>", UnknownValue{}.String())
	assert.Equal(t, "<computed>", ComputedValue{}.String())
}
