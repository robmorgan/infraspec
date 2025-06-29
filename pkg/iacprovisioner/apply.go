package iacprovisioner

// InitAndApply runs terraform init and apply with the given options and return stdout/stderr from the apply command. Note that this
// method does NOT call destroy and assumes the caller is responsible for cleaning up any resources created by running
// apply.
func InitAndApply(options *Options) (string, error) {
	if _, err := Init(options); err != nil {
		return "", err
	}

	return Apply(options)
}

// Apply runs apply with the given options and return stdout/stderr. Note that this method does NOT call destroy and
// assumes the caller is responsible for cleaning up any resources created by running apply.
func Apply(options *Options) (string, error) {
	return RunCommand(options, FormatArgs(options, "apply", "-input=false", "-auto-approve")...)
}
