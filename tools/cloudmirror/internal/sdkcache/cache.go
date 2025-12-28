// Package sdkcache provides functionality for caching the AWS SDK Go V2 repository.
// It automatically downloads and manages the SDK in ~/.cloudmirror/aws-sdk-go-v2,
// similar to how npm caches packages in ~/.npm.
package sdkcache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// SDKRepoURL is the GitHub repository URL for AWS SDK Go V2
	SDKRepoURL = "https://github.com/aws/aws-sdk-go-v2.git"

	// DefaultCacheDir is the default directory name for the cache
	DefaultCacheDir = ".cloudmirror"

	// SDKDirName is the name of the SDK directory within the cache
	SDKDirName = "aws-sdk-go-v2"
)

// SDKCache manages the cached AWS SDK repository
type SDKCache struct {
	CacheDir string // Base cache directory (e.g., ~/.cloudmirror)
	Verbose  bool   // Enable verbose output
	Quiet    bool   // Suppress non-essential output
}

// NewSDKCache creates a new SDKCache with the default cache directory
func NewSDKCache(verbose, quiet bool) (*SDKCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &SDKCache{
		CacheDir: filepath.Join(homeDir, DefaultCacheDir),
		Verbose:  verbose,
		Quiet:    quiet,
	}, nil
}

// GetSDKDir returns the full path to the cached SDK directory
func (c *SDKCache) GetSDKDir() string {
	return filepath.Join(c.CacheDir, SDKDirName)
}

// HasCache checks if the SDK is already cached and valid
func (c *SDKCache) HasCache() bool {
	return hasModelsDir(c.GetSDKDir())
}

// GetSDKPath returns the cached SDK path, cloning if necessary.
// If version is empty, uses the latest main branch.
func (c *SDKCache) GetSDKPath(version string) (string, error) {
	sdkDir := c.GetSDKDir()

	// Check if already cached
	if c.HasCache() {
		// If a specific version is requested, checkout that version
		if version != "" {
			if err := c.Checkout(version); err != nil {
				return "", err
			}
		}
		return sdkDir, nil
	}

	// Clone the repository
	if err := c.clone(); err != nil {
		return "", err
	}

	// Checkout specific version if requested
	if version != "" {
		if err := c.Checkout(version); err != nil {
			return "", err
		}
	}

	return sdkDir, nil
}

// clone performs a shallow clone of the AWS SDK repository
func (c *SDKCache) clone() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	sdkDir := c.GetSDKDir()

	// Remove any existing incomplete clone
	if _, err := os.Stat(sdkDir); err == nil {
		if err := os.RemoveAll(sdkDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "AWS SDK not found locally. Downloading to %s...\n", sdkDir)
		fmt.Fprintf(os.Stderr, "Cloning aws-sdk-go-v2 (this may take a moment)...\n")
	}

	// Shallow clone with single branch for speed
	args := []string{"clone", "--depth", "1", "--single-branch", SDKRepoURL, sdkDir}
	if err := c.runGit(args...); err != nil {
		return fmt.Errorf("failed to clone SDK: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "SDK cached successfully\n")
	}

	return nil
}

// Update fetches the latest changes from the remote repository
func (c *SDKCache) Update() error {
	if !c.HasCache() {
		return fmt.Errorf("SDK not cached. Run a command that requires the SDK first")
	}

	sdkDir := c.GetSDKDir()

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Updating cached SDK...\n")
	}

	// Fetch latest changes
	if err := c.runGitInDir(sdkDir, "fetch", "--depth", "1", "origin", "main"); err != nil {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	// Reset to latest main
	if err := c.runGitInDir(sdkDir, "reset", "--hard", "origin/main"); err != nil {
		return fmt.Errorf("failed to reset to latest: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "SDK updated to latest\n")
	}

	return nil
}

// Checkout switches to a specific version (tag or commit)
func (c *SDKCache) Checkout(version string) error {
	if !c.HasCache() {
		return fmt.Errorf("SDK not cached")
	}

	sdkDir := c.GetSDKDir()

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Checking out SDK version %s...\n", version)
	}

	// Fetch the specific tag with depth
	if err := c.runGitInDir(sdkDir, "fetch", "--depth", "1", "origin", "tag", version); err != nil {
		return fmt.Errorf("failed to fetch version %s: %w", version, err)
	}

	// Checkout the tag
	if err := c.runGitInDir(sdkDir, "checkout", version); err != nil {
		return fmt.Errorf("failed to checkout version %s: %w", version, err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Switched to SDK version %s\n", version)
	}

	return nil
}

// Clean removes the cached SDK
func (c *SDKCache) Clean() error {
	sdkDir := c.GetSDKDir()

	if _, err := os.Stat(sdkDir); os.IsNotExist(err) {
		if !c.Quiet {
			fmt.Fprintf(os.Stderr, "Cache is already clean\n")
		}
		return nil
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Removing cached SDK at %s...\n", sdkDir)
	}

	if err := os.RemoveAll(sdkDir); err != nil {
		return fmt.Errorf("failed to remove cached SDK: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Cache cleaned successfully\n")
	}

	return nil
}

// Status returns information about the cached SDK
func (c *SDKCache) Status() (*CacheStatus, error) {
	sdkDir := c.GetSDKDir()
	status := &CacheStatus{
		CacheDir: sdkDir,
		Exists:   false,
	}

	if !c.HasCache() {
		return status, nil
	}

	status.Exists = true

	// Get current commit/tag
	version, err := c.getCurrentVersion()
	if err == nil {
		status.Version = version
	}

	// Get current commit SHA
	sha, err := c.getCurrentCommit()
	if err == nil {
		status.CommitSHA = sha
	}

	// Calculate directory size
	size, err := dirSize(sdkDir)
	if err == nil {
		status.SizeBytes = size
		status.SizeHuman = humanReadableSize(size)
	}

	return status, nil
}

// ListVersions returns available SDK tags (versions)
func (c *SDKCache) ListVersions(limit int) ([]string, error) {
	if !c.HasCache() {
		return nil, fmt.Errorf("SDK not cached. Run 'cloudmirror sdk update' first")
	}

	sdkDir := c.GetSDKDir()

	// Fetch all tags
	if err := c.runGitInDir(sdkDir, "fetch", "--tags", "--depth", "1"); err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	// List tags
	output, err := c.runGitOutput(sdkDir, "tag", "-l", "--sort=-v:refname")
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(output), "\n")

	// Filter to only include release tags (e.g., v1.30.0, release-2024-01-01)
	var versions []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		// Include main version tags
		if strings.HasPrefix(tag, "v") || strings.HasPrefix(tag, "release-") {
			versions = append(versions, tag)
		}
	}

	// Sort by version (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	// Limit results if requested
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}

	return versions, nil
}

// CacheStatus contains information about the cached SDK
type CacheStatus struct {
	CacheDir  string
	Exists    bool
	Version   string
	CommitSHA string
	SizeBytes int64
	SizeHuman string
}

// getCurrentVersion attempts to get the current tag or branch name
func (c *SDKCache) getCurrentVersion() (string, error) {
	sdkDir := c.GetSDKDir()

	// Try to get tag name
	output, err := c.runGitOutput(sdkDir, "describe", "--tags", "--exact-match")
	if err == nil && output != "" {
		return strings.TrimSpace(output), nil
	}

	// Fall back to branch name
	output, err = c.runGitOutput(sdkDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		return strings.TrimSpace(output), nil
	}

	return "unknown", nil
}

// getCurrentCommit returns the current commit SHA
func (c *SDKCache) getCurrentCommit() (string, error) {
	sdkDir := c.GetSDKDir()
	output, err := c.runGitOutput(sdkDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// runGit executes a git command
func (c *SDKCache) runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	if c.Verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// runGitInDir executes a git command in a specific directory
func (c *SDKCache) runGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if c.Verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// runGitOutput executes a git command and returns its output
func (c *SDKCache) runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.Output()
	return string(output), err
}

// IsGitAvailable checks if git is installed and accessible
func IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// hasModelsDir checks if the SDK path contains the Smithy model files
func hasModelsDir(sdkPath string) bool {
	modelsPath := filepath.Join(sdkPath, "codegen", "sdk-codegen", "aws-models")
	info, err := os.Stat(modelsPath)
	return err == nil && info.IsDir()
}

// dirSize calculates the total size of a directory
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// humanReadableSize converts bytes to human-readable format
func humanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
