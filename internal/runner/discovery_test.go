package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverFeatureFiles_SingleFile(t *testing.T) {
	// Create a temporary directory with a feature file
	tmpDir := t.TempDir()
	featureFile := filepath.Join(tmpDir, "test.feature")
	err := os.WriteFile(featureFile, []byte("Feature: Test"), 0o644)
	require.NoError(t, err)

	files, err := DiscoverFeatureFiles(featureFile)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, files[0], "test.feature")
}

func TestDiscoverFeatureFiles_Directory(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create feature files in root and subdirectory
	files := []string{
		filepath.Join(tmpDir, "test1.feature"),
		filepath.Join(tmpDir, "test2.feature"),
		filepath.Join(tmpDir, "subdir", "test3.feature"),
	}

	// Create subdirectory
	err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o755)
	require.NoError(t, err)

	// Create feature files
	for _, f := range files {
		err := os.WriteFile(f, []byte("Feature: Test"), 0o644)
		require.NoError(t, err)
	}

	// Create a non-feature file that should be ignored
	err = os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# README"), 0o644)
	require.NoError(t, err)

	discovered, err := DiscoverFeatureFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, discovered, 3)
}

func TestDiscoverFeatureFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := DiscoverFeatureFiles(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .feature files found")
}

func TestDiscoverFeatureFiles_NonFeatureFile(t *testing.T) {
	tmpDir := t.TempDir()
	textFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(textFile, []byte("text"), 0o644)
	require.NoError(t, err)

	_, err = DiscoverFeatureFiles(textFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have .feature extension")
}

func TestDiscoverFeatureFiles_NonExistent(t *testing.T) {
	_, err := DiscoverFeatureFiles("/nonexistent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access path")
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
