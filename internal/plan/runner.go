package plan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// ErrTerraformNotFound is returned when the terraform binary is not found in PATH.
var ErrTerraformNotFound = errors.New("terraform binary not found in PATH")

const (
	// dirPermissions is the default permission mode for created directories.
	dirPermissions = 0o755
	// filePermissions is the default permission mode for created files.
	filePermissions = 0o600
)

// PlanOptions provides a simplified configuration for running terraform plan.
type PlanOptions struct {
	// VarFiles are paths to .tfvars files.
	VarFiles []string
	// Vars are individual variable values.
	Vars map[string]string
	// Parallelism is the Terraform parallelism setting (0 = default).
	Parallelism int
	// Timeout is the maximum execution time (0 = no timeout).
	Timeout time.Duration
	// EnvVars are additional environment variables to set.
	EnvVars map[string]string
}

// Runner handles execution of terraform plan and parsing of the results.
type Runner struct {
	options *iacprovisioner.Options
}

// NewRunner creates a new plan runner with the given options.
func NewRunner(options *iacprovisioner.Options) *Runner {
	return &Runner{
		options: options,
	}
}

// Run executes terraform plan and returns the parsed plan.
// It creates a temporary plan file, runs terraform show to get JSON output,
// and parses the result.
func (r *Runner) Run() (*Plan, error) {
	// Create a temp file for the plan output
	planFile, err := os.CreateTemp("", "infraspec-plan-*.tfplan")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp plan file: %w", err)
	}
	planFilePath := planFile.Name()
	planFile.Close()
	defer os.Remove(planFilePath)

	// Run terraform plan with -out flag
	opts := r.options
	opts.PlanFilePath = planFilePath
	if _, err := RunPlan(opts); err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	// Run terraform show -json to get the plan JSON
	jsonOutput, err := ShowJSON(opts, planFilePath)
	if err != nil {
		return nil, fmt.Errorf("terraform show failed: %w", err)
	}

	// Parse the JSON output
	return ParsePlanBytes([]byte(jsonOutput))
}

// RunFromFile parses an existing terraform plan JSON file.
func (r *Runner) RunFromFile(jsonPath string) (*Plan, error) {
	return ParsePlanFile(jsonPath)
}

// RunPlan runs terraform plan and saves the output to the path specified in options.PlanFilePath.
func RunPlan(options *iacprovisioner.Options) (string, error) {
	args := []string{"plan", "-input=false"}

	if options.PlanFilePath != "" {
		args = append(args, fmt.Sprintf("-out=%s", options.PlanFilePath))
	}

	return iacprovisioner.RunCommand(options, iacprovisioner.FormatArgs(options, args...)...)
}

// PlanJSON runs terraform plan and returns the plan output as JSON.
// This runs plan followed by show -json on the resulting plan file.
func PlanJSON(options *iacprovisioner.Options) (string, error) {
	// Create a temp file for the plan output
	planFile, err := os.CreateTemp("", "infraspec-plan-*.tfplan")
	if err != nil {
		return "", fmt.Errorf("failed to create temp plan file: %w", err)
	}
	planFilePath := planFile.Name()
	planFile.Close()
	defer os.Remove(planFilePath)

	// Run terraform plan
	opts := *options // Copy to avoid modifying original
	opts.PlanFilePath = planFilePath
	if _, err := RunPlan(&opts); err != nil {
		return "", err
	}

	// Run terraform show -json
	return ShowJSON(&opts, planFilePath)
}

// ShowJSON runs terraform show -json on a plan file and returns the JSON output.
func ShowJSON(options *iacprovisioner.Options, planFilePath string) (string, error) {
	args := []string{"show", "-json"}

	// Add the plan file path if provided
	if planFilePath != "" {
		args = append(args, planFilePath)
	}

	// Add any extra args for the show command
	args = append(options.ExtraArgs.Show, args...)

	return iacprovisioner.RunCommand(options, args...)
}

// InitAndPlan runs terraform init followed by terraform plan.
func InitAndPlan(options *iacprovisioner.Options) (string, error) {
	if _, err := iacprovisioner.Init(options); err != nil {
		return "", err
	}
	return RunPlan(options)
}

// InitAndPlanJSON runs terraform init followed by terraform plan,
// then returns the plan as JSON.
func InitAndPlanJSON(options *iacprovisioner.Options) (string, error) {
	if _, err := iacprovisioner.Init(options); err != nil {
		return "", err
	}
	return PlanJSON(options)
}

// InitAndPlanParsed runs terraform init, plan, and returns a parsed Plan struct.
func InitAndPlanParsed(options *iacprovisioner.Options) (*Plan, error) {
	jsonOutput, err := InitAndPlanJSON(options)
	if err != nil {
		return nil, err
	}
	return ParsePlanBytes([]byte(jsonOutput))
}

// SavePlanJSON saves the plan JSON output to a file.
func SavePlanJSON(options *iacprovisioner.Options, outputPath string) error {
	jsonOutput, err := PlanJSON(options)
	if err != nil {
		return err
	}

	// Ensure the output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(jsonOutput), filePermissions)
}

// FindTerraformBinary returns the path to the terraform binary or an error.
func FindTerraformBinary() (string, error) {
	path, err := exec.LookPath("terraform")
	if err != nil {
		return "", ErrTerraformNotFound
	}
	return path, nil
}

// GeneratePlan is a high-level convenience function that:
// 1. Checks if terraform binary exists
// 2. Runs terraform init if .terraform/ doesn't exist
// 3. Runs terraform plan -out=tfplan -input=false
// 4. Runs terraform show -json tfplan
// 5. Parses and returns the Plan
// 6. Cleans up the tfplan file
func GeneratePlan(dir string, opts PlanOptions) (*Plan, error) {
	return GeneratePlanWithContext(context.Background(), dir, opts)
}

// GeneratePlanWithContext is like GeneratePlan but accepts a context for cancellation and timeout.
func GeneratePlanWithContext(ctx context.Context, dir string, opts PlanOptions) (*Plan, error) {
	if _, err := FindTerraformBinary(); err != nil {
		return nil, err
	}

	absDir, err := validateDirectory(dir)
	if err != nil {
		return nil, err
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	provOpts := buildProvisionerOptions(absDir, opts)

	if err := ensureInitialized(ctx, absDir, provOpts); err != nil {
		return nil, err
	}

	return runPlanAndParse(ctx, provOpts)
}

// validateDirectory checks that dir exists and is a directory, returning the absolute path.
func validateDirectory(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve directory path: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", absDir)
		}
		return "", fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absDir)
	}

	return absDir, nil
}

// buildProvisionerOptions creates iacprovisioner.Options from PlanOptions.
func buildProvisionerOptions(absDir string, opts PlanOptions) *iacprovisioner.Options {
	provOpts := &iacprovisioner.Options{
		WorkingDir:  absDir,
		VarFiles:    opts.VarFiles,
		Parallelism: opts.Parallelism,
		EnvVars:     opts.EnvVars,
	}

	if len(opts.Vars) > 0 {
		provOpts.Vars = make(map[string]interface{}, len(opts.Vars))
		for k, v := range opts.Vars {
			provOpts.Vars[k] = v
		}
	}

	return provOpts
}

// ensureInitialized runs terraform init if the .terraform directory doesn't exist.
func ensureInitialized(ctx context.Context, absDir string, provOpts *iacprovisioner.Options) error {
	terraformDir := filepath.Join(absDir, ".terraform")
	if _, err := os.Stat(terraformDir); os.IsNotExist(err) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context canceled before init: %w", err)
		}
		if _, err := iacprovisioner.Init(provOpts); err != nil {
			return fmt.Errorf("terraform init failed: %w", err)
		}
	}
	return nil
}

// runPlanAndParse executes terraform plan, show -json, and parses the result.
func runPlanAndParse(ctx context.Context, provOpts *iacprovisioner.Options) (*Plan, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before plan: %w", err)
	}

	planFile, err := os.CreateTemp("", "infraspec-plan-*.tfplan")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp plan file: %w", err)
	}
	planFilePath := planFile.Name()
	planFile.Close()
	defer os.Remove(planFilePath)

	provOpts.PlanFilePath = planFilePath
	if _, err := RunPlan(provOpts); err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before show: %w", err)
	}

	jsonOutput, err := ShowJSON(provOpts, planFilePath)
	if err != nil {
		return nil, fmt.Errorf("terraform show failed: %w", err)
	}

	return ParsePlanBytes([]byte(jsonOutput))
}
