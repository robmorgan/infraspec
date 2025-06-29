package iacprovisioner

// Destroy runs terraform destroy with the given options and return stdout/stderr.
func Destroy(options *Options) (string, error) {
	return RunCommand(options, FormatArgs(options, prepend(options.ExtraArgs.Destroy, "destroy", "-auto-approve", "-input=false")...)...)
}
