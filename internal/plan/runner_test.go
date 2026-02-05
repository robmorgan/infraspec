package plan

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTerraformBinary(t *testing.T) {
	path, err := FindTerraformBinary()
	if err != nil {
		// terraform not in PATH - this is expected in some CI environments
		assert.ErrorIs(t, err, ErrTerraformNotFound)
		t.Skip("terraform not in PATH")
	}
	assert.NotEmpty(t, path)
}

func TestGeneratePlan_TerraformNotFound(t *testing.T) {
	// This test requires terraform to not be in PATH
	// We can't easily test this without PATH manipulation
	t.Skip("requires PATH manipulation to test terraform not found")
}

func TestGeneratePlan_DirectoryNotFound(t *testing.T) {
	// Skip if terraform not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not available")
	}

	_, err := GeneratePlan("/nonexistent/path/to/terraform", PlanOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestGeneratePlan_NotADirectory(t *testing.T) {
	// Skip if terraform not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not available")
	}

	// Use a file path instead of directory
	filePath := filepath.Join("testdata", "plans", "vpc_basic.json")
	_, err := GeneratePlan(filePath, PlanOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestGeneratePlanWithContext_Timeout(t *testing.T) {
	// Skip if terraform not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not available")
	}

	// Create a context that's already canceled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout fires

	_, err := GeneratePlanWithContext(ctx, ".", PlanOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestPlanOptions_Defaults(t *testing.T) {
	opts := PlanOptions{}

	assert.Nil(t, opts.VarFiles)
	assert.Nil(t, opts.Vars)
	assert.Zero(t, opts.Parallelism)
	assert.Zero(t, opts.Timeout)
	assert.Nil(t, opts.EnvVars)
}
