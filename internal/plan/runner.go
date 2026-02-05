package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

const (
	// dirPermissions is the default permission mode for created directories.
	dirPermissions = 0o755
	// filePermissions is the default permission mode for created files.
	filePermissions = 0o600
)

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
