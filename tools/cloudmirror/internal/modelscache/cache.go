// Package modelscache provides functionality for caching the AWS API Models repository.
// It automatically downloads and manages the models in ~/.cloudmirror/api-models-aws.
package modelscache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// ModelsRepoURL is the GitHub repository URL for AWS API Models
	ModelsRepoURL = "https://github.com/aws/api-models-aws.git"

	// DefaultCacheDir is the default directory name for the cache
	DefaultCacheDir = ".cloudmirror"

	// ModelsDirName is the name of the models directory within the cache
	ModelsDirName = "api-models-aws"
)

// ModelsCache manages the cached AWS API Models repository
type ModelsCache struct {
	CacheDir string // Base cache directory (e.g., ~/.cloudmirror)
	Verbose  bool   // Enable verbose output
	Quiet    bool   // Suppress non-essential output
}

// NewModelsCache creates a new ModelsCache with the default cache directory
func NewModelsCache(verbose, quiet bool) (*ModelsCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &ModelsCache{
		CacheDir: filepath.Join(homeDir, DefaultCacheDir),
		Verbose:  verbose,
		Quiet:    quiet,
	}, nil
}

// GetModelsDir returns the full path to the cached models directory
func (c *ModelsCache) GetModelsDir() string {
	return filepath.Join(c.CacheDir, ModelsDirName)
}

// HasCache checks if the models are already cached and valid
func (c *ModelsCache) HasCache() bool {
	return hasModelsDir(c.GetModelsDir())
}

// GetModelsPath returns the cached models path, cloning if necessary
func (c *ModelsCache) GetModelsPath() (string, error) {
	modelsDir := c.GetModelsDir()

	// Check if already cached
	if c.HasCache() {
		return modelsDir, nil
	}

	// Clone the repository
	if err := c.clone(); err != nil {
		return "", err
	}

	return modelsDir, nil
}

// clone performs a shallow clone of the AWS API Models repository
func (c *ModelsCache) clone() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	modelsDir := c.GetModelsDir()

	// Remove any existing incomplete clone
	if _, err := os.Stat(modelsDir); err == nil {
		if err := os.RemoveAll(modelsDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "AWS API Models not found locally. Downloading to %s...\n", modelsDir)
		fmt.Fprintf(os.Stderr, "Cloning api-models-aws (this may take a moment)...\n")
	}

	// Shallow clone with single branch for speed
	args := []string{"clone", "--depth", "1", "--single-branch", ModelsRepoURL, modelsDir}
	if err := c.runGit(args...); err != nil {
		return fmt.Errorf("failed to clone models: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Models cached successfully\n")
	}

	return nil
}

// Update fetches the latest changes from the remote repository
func (c *ModelsCache) Update() error {
	if !c.HasCache() {
		return fmt.Errorf("models not cached. Run a command that requires the models first")
	}

	modelsDir := c.GetModelsDir()

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Updating cached models...\n")
	}

	// Fetch latest changes
	if err := c.runGitInDir(modelsDir, "fetch", "--depth", "1", "origin", "main"); err != nil {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	// Reset to latest main
	if err := c.runGitInDir(modelsDir, "reset", "--hard", "origin/main"); err != nil {
		return fmt.Errorf("failed to reset to latest: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Models updated to latest\n")
	}

	return nil
}

// Clean removes the cached models
func (c *ModelsCache) Clean() error {
	modelsDir := c.GetModelsDir()

	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		if !c.Quiet {
			fmt.Fprintf(os.Stderr, "Models cache is already clean\n")
		}
		return nil
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Removing cached models at %s...\n", modelsDir)
	}

	if err := os.RemoveAll(modelsDir); err != nil {
		return fmt.Errorf("failed to remove cached models: %w", err)
	}

	if !c.Quiet {
		fmt.Fprintf(os.Stderr, "Models cache cleaned successfully\n")
	}

	return nil
}

// Status returns information about the cached models
func (c *ModelsCache) Status() (*CacheStatus, error) {
	modelsDir := c.GetModelsDir()
	status := &CacheStatus{
		CacheDir: modelsDir,
		Exists:   false,
	}

	if !c.HasCache() {
		return status, nil
	}

	status.Exists = true

	// Get current commit SHA
	sha, err := c.getCurrentCommit()
	if err == nil {
		status.CommitSHA = sha
	}

	// Calculate directory size
	size, err := dirSize(modelsDir)
	if err == nil {
		status.SizeBytes = size
		status.SizeHuman = humanReadableSize(size)
	}

	return status, nil
}

// FindModelPath finds the model file for a given service
// Models are located at: models/<service>/service/<version>/<service>-<version>.json
func (c *ModelsCache) FindModelPath(serviceName string) (string, error) {
	modelsDir := c.GetModelsDir()
	serviceName = strings.ToLower(serviceName)

	// Map common service name variations
	serviceNameMap := map[string]string{
		"applicationautoscaling": "application-auto-scaling",
		"autoscaling":            "auto-scaling",
	}

	if mapped, ok := serviceNameMap[serviceName]; ok {
		serviceName = mapped
	}

	// Look for the service directory
	serviceDir := filepath.Join(modelsDir, "models", serviceName, "service")
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		// Try with hyphens converted
		serviceName = strings.ReplaceAll(serviceName, "_", "-")
		serviceDir = filepath.Join(modelsDir, "models", serviceName, "service")
		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			return "", fmt.Errorf("service %s not found in models", serviceName)
		}
	}

	// Find the latest version directory
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return "", fmt.Errorf("failed to read service directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for service %s", serviceName)
	}

	// Sort versions and get the latest
	sort.Strings(versions)
	latestVersion := versions[len(versions)-1]

	// Find the JSON model file
	versionDir := filepath.Join(serviceDir, latestVersion)
	jsonFiles, err := filepath.Glob(filepath.Join(versionDir, "*.json"))
	if err != nil || len(jsonFiles) == 0 {
		return "", fmt.Errorf("no model file found for service %s version %s", serviceName, latestVersion)
	}

	return jsonFiles[0], nil
}

// ListServices returns a list of available services in the models repo
func (c *ModelsCache) ListServices() ([]string, error) {
	modelsDir := c.GetModelsDir()
	modelsPath := filepath.Join(modelsDir, "models")

	entries, err := os.ReadDir(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models directory: %w", err)
	}

	var services []string
	for _, entry := range entries {
		if entry.IsDir() {
			services = append(services, entry.Name())
		}
	}

	sort.Strings(services)
	return services, nil
}

// CacheStatus contains information about the cached models
type CacheStatus struct {
	CacheDir  string
	Exists    bool
	CommitSHA string
	SizeBytes int64
	SizeHuman string
}

// getCurrentCommit returns the current commit SHA
func (c *ModelsCache) getCurrentCommit() (string, error) {
	modelsDir := c.GetModelsDir()
	output, err := c.runGitOutput(modelsDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// runGit executes a git command
func (c *ModelsCache) runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	if c.Verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// runGitInDir executes a git command in a specific directory
func (c *ModelsCache) runGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if c.Verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// runGitOutput executes a git command and returns its output
func (c *ModelsCache) runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.Output()
	return string(output), err
}

// hasModelsDir checks if the models path contains the expected directory structure
func hasModelsDir(modelsPath string) bool {
	modelsSubdir := filepath.Join(modelsPath, "models")
	info, err := os.Stat(modelsSubdir)
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
