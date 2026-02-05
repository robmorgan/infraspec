package testfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile_BasicStructure(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	// Check source file is set
	assert.Equal(t, "testdata/testfiles/vpc_test.infraspec.hcl", tf.SourceFile)

	// Check variables
	assert.Len(t, tf.Variables, 2)
	assert.Equal(t, "test", tf.Variables["environment"])
	assert.Equal(t, "10.0.0.0/16", tf.Variables["cidr_block"])

	// Check runs
	assert.Len(t, tf.Runs, 2)
	assert.Equal(t, []string{"vpc_dns_enabled", "idempotent"}, tf.RunNames())
}

func TestParseFile_RunBlock(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)

	// First run block
	run := tf.RunByName("vpc_dns_enabled")
	require.NotNil(t, run)
	assert.Equal(t, "./modules/vpc", run.Module)
	assert.Equal(t, "", run.State)
	assert.Equal(t, "plan", run.Command) // default value
	assert.Len(t, run.Asserts, 2)

	// Second run block with state
	run2 := tf.RunByName("idempotent")
	require.NotNil(t, run2)
	assert.Equal(t, "./modules/vpc", run2.Module)
	assert.Equal(t, "./fixtures/applied.tfstate.json", run2.State)
	assert.Len(t, run2.Asserts, 1)
}

func TestParseFile_AssertBlock(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)

	run := tf.RunByName("vpc_dns_enabled")
	require.NotNil(t, run)
	require.Len(t, run.Asserts, 2)

	// First assertion
	assert1 := run.Asserts[0]
	assert.NotNil(t, assert1.Condition)
	assert.NotEmpty(t, assert1.ConditionRaw)
	assert.Equal(t, "VPC ID must not be empty", assert1.ErrorMessage)

	// Second assertion
	assert2 := run.Asserts[1]
	assert.NotNil(t, assert2.Condition)
	assert.Equal(t, "VPC must have DNS support enabled", assert2.ErrorMessage)
}

func TestParseFile_StateFixture(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)

	run := tf.RunByName("idempotent")
	require.NotNil(t, run)
	assert.Equal(t, "./fixtures/applied.tfstate.json", run.State)
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("testdata/testfiles/nonexistent.infraspec.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read test file")
}

func TestParseFile_InvalidHCL(t *testing.T) {
	_, err := ParseFile("testdata/testfiles/invalid.infraspec.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse HCL")
}

func TestParseFile_MissingRequired(t *testing.T) {
	_, err := ParseFile("testdata/testfiles/missing_module.infraspec.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module")
}

func TestParseBytes_Success(t *testing.T) {
	content := []byte(`
run "test" {
  module = "./modules/test"

  assert {
    condition     = true
    error_message = "always passes"
  }
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)
	assert.Len(t, tf.Runs, 1)
	assert.Equal(t, "test", tf.Runs[0].Name)
}

func TestParseBytes_WithCommand(t *testing.T) {
	content := []byte(`
run "apply_test" {
  module  = "./modules/test"
  command = "apply"

  assert {
    condition     = true
    error_message = "test"
  }
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)
	assert.Equal(t, "apply", tf.Runs[0].Command)
}

func TestParseDir_Success(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create first test file
	file1 := filepath.Join(tmpDir, "first_test.infraspec.hcl")
	content1 := []byte(`
run "first" {
  module = "./modules/a"

  assert {
    condition     = true
    error_message = "first"
  }
}
`)
	require.NoError(t, os.WriteFile(file1, content1, 0o600))

	// Create second test file
	file2 := filepath.Join(tmpDir, "second_test.infraspec.hcl")
	content2 := []byte(`
run "second" {
  module = "./modules/b"

  assert {
    condition     = true
    error_message = "second"
  }
}
`)
	require.NoError(t, os.WriteFile(file2, content2, 0o600))

	// Parse directory
	testFiles, err := ParseDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, testFiles, 2)
}

func TestParseDir_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	testFiles, err := ParseDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, testFiles)
}

func TestParseDir_NotADirectory(t *testing.T) {
	_, err := ParseDir("testdata/testfiles/vpc_test.infraspec.hcl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestParseDir_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid test file
	invalidFile := filepath.Join(tmpDir, "invalid_test.infraspec.hcl")
	content := []byte(`run "broken" { # unclosed`)
	require.NoError(t, os.WriteFile(invalidFile, content, 0o600))

	_, err := ParseDir(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestValidate_DuplicateRunNames(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/duplicate_runs.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	err = tf.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate run name")
	assert.Contains(t, err.Error(), "same_name")
}

func TestValidate_Success(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)

	err = tf.Validate()
	require.NoError(t, err)
}

func TestRunByName_NotFound(t *testing.T) {
	tf, err := ParseFile("testdata/testfiles/vpc_test.infraspec.hcl")
	require.NoError(t, err)

	run := tf.RunByName("nonexistent")
	assert.Nil(t, run)
}

func TestParseBytes_VariableTypes(t *testing.T) {
	content := []byte(`
variables {
  string_var = "hello"
  int_var    = 42
  float_var  = 3.14
  bool_var   = true
  list_var   = ["a", "b", "c"]
  map_var    = {
    key1 = "value1"
    key2 = "value2"
  }
}

run "types_test" {
  module = "./modules/test"

  assert {
    condition     = true
    error_message = "test"
  }
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	assert.Equal(t, "hello", tf.Variables["string_var"])
	assert.Equal(t, int64(42), tf.Variables["int_var"])
	assert.Equal(t, 3.14, tf.Variables["float_var"])
	assert.Equal(t, true, tf.Variables["bool_var"])

	listVar, ok := tf.Variables["list_var"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, []interface{}{"a", "b", "c"}, listVar)

	mapVar, ok := tf.Variables["map_var"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value1", mapVar["key1"])
	assert.Equal(t, "value2", mapVar["key2"])
}

func TestParseBytes_NoVariables(t *testing.T) {
	content := []byte(`
run "no_vars" {
  module = "./modules/test"

  assert {
    condition     = true
    error_message = "no variables"
  }
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	assert.Empty(t, tf.Variables)
	assert.Len(t, tf.Runs, 1)
}

func TestParseBytes_MultipleAsserts(t *testing.T) {
	content := []byte(`
run "multi_assert" {
  module = "./modules/test"

  assert {
    condition     = true
    error_message = "first"
  }

  assert {
    condition     = false
    error_message = "second"
  }

  assert {
    condition     = 1 == 1
    error_message = "third"
  }
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	run := tf.Runs[0]
	assert.Len(t, run.Asserts, 3)
	assert.Equal(t, "first", run.Asserts[0].ErrorMessage)
	assert.Equal(t, "second", run.Asserts[1].ErrorMessage)
	assert.Equal(t, "third", run.Asserts[2].ErrorMessage)
}

func TestParseBytes_NoAsserts(t *testing.T) {
	content := []byte(`
run "no_asserts" {
  module = "./modules/test"
}
`)
	tf, err := ParseBytes(content, "test.infraspec.hcl")
	require.NoError(t, err)
	require.NotNil(t, tf)

	run := tf.Runs[0]
	assert.Empty(t, run.Asserts)
}
