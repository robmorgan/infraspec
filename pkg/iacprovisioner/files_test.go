package iacprovisioner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldFilterTerraformPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{
			name:     "allows regular .tf files",
			path:     "main.tf",
			isDir:    false,
			expected: false,
		},
		{
			name:     "allows .terraform-version file",
			path:     ".terraform-version",
			isDir:    false,
			expected: false,
		},
		{
			name:     "allows .terraform.lock.hcl file",
			path:     ".terraform.lock.hcl",
			isDir:    false,
			expected: false,
		},
		{
			name:     "filters terraform.tfstate file",
			path:     "terraform.tfstate",
			isDir:    false,
			expected: true,
		},
		{
			name:     "filters terraform.tfstate.backup file",
			path:     "terraform.tfstate.backup",
			isDir:    false,
			expected: true,
		},
		{
			name:     "filters terraform.tfvars file",
			path:     "terraform.tfvars",
			isDir:    false,
			expected: true,
		},
		{
			name:     "filters terraform.tfvars.json file",
			path:     "terraform.tfvars.json",
			isDir:    false,
			expected: true,
		},
		{
			name:     "filters .terraform directory",
			path:     ".terraform",
			isDir:    true,
			expected: true,
		},
		{
			name:     "filters .git directory",
			path:     ".git",
			isDir:    true,
			expected: true,
		},
		{
			name:     "filters hidden files",
			path:     ".hidden",
			isDir:    false,
			expected: true,
		},
		{
			name:     "allows nested .tf files",
			path:     "modules/vpc/main.tf",
			isDir:    false,
			expected: false,
		},
		{
			name:     "filters nested state files",
			path:     "modules/vpc/terraform.tfstate",
			isDir:    false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &mockFileInfo{isDir: tt.isDir}
			result := shouldFilterTerraformPath(tt.path, info)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCopyTerraformFolderToTemp(t *testing.T) {
	// Create a temporary source directory with test files
	srcDir, err := os.MkdirTemp("", "infraspec-test-src-")
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)

	// Create test files and directories
	testFiles := map[string]string{
		"main.tf":                    "resource \"null_resource\" \"test\" {}",
		"variables.tf":               "variable \"test\" {}",
		"outputs.tf":                 "output \"test\" { value = \"test\" }",
		".terraform-version":         "1.5.0",
		".terraform.lock.hcl":        "# Lock file",
		"terraform.tfstate":          "should be filtered",
		"terraform.tfstate.backup":   "should be filtered",
		"terraform.tfvars":           "should be filtered",
		"terraform.tfvars.json":      "should be filtered",
		".git/config":                "should be filtered",
		".hidden":                    "should be filtered",
		"modules/vpc/main.tf":        "module content",
		"modules/vpc/.terraform/foo": "should be filtered",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(srcDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			require.NoError(t, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			require.NoError(t, err)
		}
	}

	// Copy to temp directory
	tempDir, err := CopyTerraformFolderToTemp(srcDir, "infraspec-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Verify that the temp directory exists
	_, err = os.Stat(tempDir)
	require.NoError(t, err)

	// Verify that expected files were copied
	expectedFiles := []string{
		"main.tf",
		"variables.tf",
		"outputs.tf",
		".terraform-version",
		".terraform.lock.hcl",
		"modules/vpc/main.tf",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tempDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected file %s to exist", file)
	}

	// Verify that filtered files were NOT copied
	filteredFiles := []string{
		"terraform.tfstate",
		"terraform.tfstate.backup",
		"terraform.tfvars",
		"terraform.tfvars.json",
		".git/config",
		".hidden",
		"modules/vpc/.terraform/foo",
	}

	for _, file := range filteredFiles {
		path := filepath.Join(tempDir, file)
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), "expected file %s to be filtered out", file)
	}
}

func TestCopyTerraformFolderToDest(t *testing.T) {
	// Create a temporary source directory
	srcDir, err := os.MkdirTemp("", "infraspec-test-src-")
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)

	// Create a test file
	testFile := filepath.Join(srcDir, "main.tf")
	err = os.WriteFile(testFile, []byte("resource \"null_resource\" \"test\" {}"), 0644)
	require.NoError(t, err)

	// Create a temporary destination directory
	destDir, err := os.MkdirTemp("", "infraspec-test-dest-")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	// Copy the folder
	err = CopyTerraformFolderToDest(srcDir, destDir)
	require.NoError(t, err)

	// Verify that the file was copied
	copiedFile := filepath.Join(destDir, "main.tf")
	content, err := os.ReadFile(copiedFile)
	require.NoError(t, err)
	assert.Equal(t, "resource \"null_resource\" \"test\" {}", string(content))
}

func TestCopySymlink(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "infraspec-symlink-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("target content"), 0644)
	require.NoError(t, err)

	// Create a symlink
	symlinkPath := filepath.Join(tmpDir, "link.txt")
	err = os.Symlink(targetFile, symlinkPath)
	require.NoError(t, err)

	// Copy the symlink
	destSymlink := filepath.Join(tmpDir, "dest-link.txt")
	err = copySymlink(symlinkPath, destSymlink)
	require.NoError(t, err)

	// Verify that the destination is a symlink
	linkInfo, err := os.Lstat(destSymlink)
	require.NoError(t, err)
	assert.True(t, linkInfo.Mode()&os.ModeSymlink != 0)

	// Verify that the symlink points to the correct target
	linkTarget, err := os.Readlink(destSymlink)
	require.NoError(t, err)
	assert.Equal(t, targetFile, linkTarget)
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
