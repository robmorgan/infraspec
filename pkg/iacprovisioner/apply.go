package iacprovisioner

import "fmt"

// InitAndApply runs terraform init and apply with the given options and return stdout/stderr from the apply command. Note that this
// method does NOT call destroy and assumes the caller is responsible for cleaning up any resources created by running
// apply.
//
// If options.CopyToTemp is true, this will copy the Terraform configuration to a temporary directory before running
// init and apply. This is useful for running tests in parallel without file conflicts or to avoid polluting the
// original source directory with generated files.
func InitAndApply(options *Options) (string, error) {
	// If CopyToTemp is enabled, copy the Terraform folder to a temporary directory
	if options.CopyToTemp {
		if err := prepareTempWorkingDir(options); err != nil {
			return "", fmt.Errorf("failed to prepare temp working directory: %w", err)
		}
	}

	if _, err := Init(options); err != nil {
		return "", err
	}

	return Apply(options)
}

// prepareTempWorkingDir copies the Terraform configuration to a temporary directory and updates the options
// to use the new directory. This allows running Terraform in isolation without affecting the original source.
func prepareTempWorkingDir(options *Options) error {
	// Set a default prefix if not provided
	if options.TempFolderPrefix == "" {
		options.TempFolderPrefix = "infraspec-terraform-"
	}

	// Store the original working directory
	options.OriginalWorkingDir = options.WorkingDir

	// Copy the Terraform folder to a temp directory
	tempDir, err := CopyTerraformFolderToTemp(options.WorkingDir, options.TempFolderPrefix)
	if err != nil {
		return fmt.Errorf("failed to copy terraform folder to temp: %w", err)
	}

	// Update the working directory to point to the temp directory
	options.WorkingDir = tempDir

	return nil
}

// Apply runs apply with the given options and return stdout/stderr. Note that this method does NOT call destroy and
// assumes the caller is responsible for cleaning up any resources created by running apply.
func Apply(options *Options) (string, error) {
	return RunCommand(options, FormatArgs(options, "apply", "-input=false", "-auto-approve")...)
}
