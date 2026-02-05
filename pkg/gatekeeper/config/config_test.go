package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigBytes_Basic(t *testing.T) {
	hcl := `
config {
  min_severity = "warning"
  format       = "json"
  strict       = true
  no_builtin   = true
}
`

	config, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)

	assert.Equal(t, "warning", config.MinSeverity)
	assert.Equal(t, "json", config.Format)
	assert.True(t, config.Strict)
	assert.True(t, config.NoBuiltin)
	assert.Empty(t, config.Rules)
}

func TestLoadConfigBytes_WithRules(t *testing.T) {
	hcl := `
config {
  min_severity = "error"
}

rule "TEST_001" {
  name          = "Test rule"
  severity      = "error"
  resource_type = "aws_s3_bucket"
  condition {
    check {
      attribute = "bucket"
      operator  = "exists"
    }
  }
  message = "Test message"
}

rule "TEST_002" {
  name          = "Another rule"
  severity      = "warning"
  resource_type = "aws_instance"
  condition {
    check {
      attribute = "instance_type"
      operator  = "exists"
    }
  }
  message = "Instance message"
}
`

	config, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)

	assert.Equal(t, "error", config.MinSeverity)
	assert.Len(t, config.Rules, 2)
	assert.Equal(t, "TEST_001", config.Rules[0].ID)
	assert.Equal(t, "TEST_002", config.Rules[1].ID)
}

func TestLoadConfigBytes_DefaultValues(t *testing.T) {
	hcl := `
config {
}
`

	config, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)

	// Should have defaults from DefaultConfig()
	assert.Equal(t, "error", config.MinSeverity)
	assert.Equal(t, "text", config.Format)
	assert.False(t, config.Strict)
	assert.False(t, config.NoBuiltin)
}

func TestLoadConfigBytes_NoConfigBlock(t *testing.T) {
	hcl := `
rule "TEST_001" {
  name          = "Test rule"
  severity      = "error"
  resource_type = "aws_s3_bucket"
  condition {
    check {
      attribute = "bucket"
      operator  = "exists"
    }
  }
  message = "Test message"
}
`

	config, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)

	// Should have defaults
	assert.Equal(t, "error", config.MinSeverity)
	assert.Equal(t, "text", config.Format)

	// But should have the rule
	assert.Len(t, config.Rules, 1)
	assert.Equal(t, "TEST_001", config.Rules[0].ID)
}

func TestLoadConfigBytes_ComplexConditions(t *testing.T) {
	hcl := `
rule "TEST_001" {
  name          = "Complex rule"
  severity      = "error"
  resource_type = "aws_s3_bucket"
  condition {
    all {
      check {
        attribute = "bucket"
        operator  = "exists"
      }
      check {
        attribute = "versioning"
        operator  = "equals"
        value     = true
      }
    }
  }
  message = "Test message"
}
`

	config, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.NoError(t, err)

	require.Len(t, config.Rules, 1)
	rule := config.Rules[0]
	assert.Equal(t, "all", string(rule.Condition.Operator))
	assert.Len(t, rule.Condition.Conditions, 2)
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create nested directories
	subDir := filepath.Join(tempDir, "subdir", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	// Create a .infraspec.hcl in the temp directory
	configPath := filepath.Join(tempDir, ".infraspec.hcl")
	require.NoError(t, os.WriteFile(configPath, []byte(`config {}`), 0o644))

	// Test finding from the nested directory
	found, err := FindConfigFile(subDir)
	require.NoError(t, err)
	assert.Equal(t, configPath, found)

	// Test finding from a file in the nested directory
	filePath := filepath.Join(subDir, "main.tf")
	require.NoError(t, os.WriteFile(filePath, []byte(`resource "test" {}`), 0o644))

	found, err = FindConfigFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, configPath, found)
}

func TestFindConfigFile_NotFound(t *testing.T) {
	// Create a temporary directory without a config file
	tempDir := t.TempDir()

	// Should return empty string, no error
	found, err := FindConfigFile(tempDir)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestLoadConfigBytes_DuplicateRuleIDs(t *testing.T) {
	hcl := `
rule "TEST_001" {
  name          = "First rule"
  severity      = "error"
  resource_type = "aws_s3_bucket"
  condition {
    check {
      attribute = "bucket"
      operator  = "exists"
    }
  }
  message = "First message"
}

rule "TEST_001" {
  name          = "Duplicate rule"
  severity      = "warning"
  resource_type = "aws_instance"
  condition {
    check {
      attribute = "id"
      operator  = "exists"
    }
  }
  message = "Duplicate message"
}
`

	_, err := LoadConfigBytes([]byte(hcl), "test.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate rule ID")
}
