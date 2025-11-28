package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverFeatureFiles finds all .feature files in the given path.
// If path is a file, returns a single-element slice.
// If path is a directory, recursively finds all .feature files.
func DiscoverFeatureFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if !info.IsDir() {
		// Single file provided
		if !strings.HasSuffix(path, ".feature") {
			return nil, fmt.Errorf("file must have .feature extension: %s", path)
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		return []string{absPath}, nil
	}

	// Directory provided - discover all feature files
	var features []string
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".feature") {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			features = append(features, absPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover feature files: %w", err)
	}

	if len(features) == 0 {
		return nil, fmt.Errorf("no .feature files found in directory: %s", path)
	}

	return features, nil
}

// UniqueStrings removes duplicate strings from a slice while preserving order.
func UniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(input))
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
