package iacprovisioner

import "fmt"

// Init calls terraform init and return stdout/stderr.
func Init(options *Options) (string, error) {
	args := []string{"init", fmt.Sprintf("-upgrade=%t", options.Upgrade)}

	// Append reconfigure option if specified
	if options.Reconfigure {
		args = append(args, "-reconfigure")
	}
	// Append combination of migrate-state and force-copy to suppress answer prompt
	if options.MigrateState {
		args = append(args, "-migrate-state", "-force-copy")
	}
	// Append no-color option if needed
	if options.NoColor {
		args = append(args, "-no-color")
	}

	args = append(args, FormatTerraformBackendConfigAsArgs(options.BackendConfig)...)
	args = append(args, FormatTerraformPluginDirAsArgs(options.PluginDir)...)
	return RunCommand(options, args...)
}
