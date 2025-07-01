package iacprovisioner

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/robmorgan/infraspec/pkg/retry"
	"github.com/robmorgan/infraspec/pkg/shell"
)

const (
	// TofuDefaultPath command to run tofu
	TofuDefaultPath = "tofu"

	// TerraformDefaultPath to run terraform
	TerraformDefaultPath = "terraform"

	// TerragruntDefaultPath to run terragrunt
	TerragruntDefaultPath = "terragrunt"
)

var commandsWithParallelism = []string{
	"plan",
	"apply",
	"destroy",
	"plan-all",
	"run-all",
	"apply-all",
	"destroy-all",
}

var DefaultExecutable = defaultExecutable()

func generateCommand(options *Options, args ...string) shell.Command {
	cmd := shell.Command{
		Name:       options.Binary,
		Args:       args,
		WorkingDir: options.WorkingDir,
		Env:        options.EnvVars,
	}
	return cmd
}

// GetCommonOptions extracts commons terraform options
func GetCommonOptions(options *Options, args ...string) (*Options, []string) {
	if options.Binary == "" {
		options.Binary = DefaultExecutable
	}

	if options.Binary == TerragruntDefaultPath {
		args = append(args, "--terragrunt-non-interactive")
		// for newer Terragrunt version, setting simplified log formatting
		if options.EnvVars == nil {
			options.EnvVars = map[string]string{}
		}
		_, tgLogSet := options.EnvVars["TERRAGRUNT_LOG_FORMAT"]
		if !tgLogSet {
			// key-value format for terragrunt logs to avoid colors and have plain form
			// https://terragrunt.gruntwork.io/docs/reference/cli-options/#terragrunt-log-format
			options.EnvVars["TERRAGRUNT_LOG_FORMAT"] = "key-value"
		}
		_, tgLogFormat := options.EnvVars["TERRAGRUNT_LOG_CUSTOM_FORMAT"]
		if !tgLogFormat {
			options.EnvVars["TERRAGRUNT_LOG_CUSTOM_FORMAT"] = "%msg(color=disable)"
		}
	}

	if options.Parallelism > 0 && len(args) > 0 && slices.Contains(commandsWithParallelism, args[0]) {
		args = append(args, fmt.Sprintf("--parallelism=%d", options.Parallelism))
	}

	// if SshAgent is provided, override the local SSH agent with the socket of our in-process agent
	if options.SshAgent != nil {
		// Initialize EnvVars, if it hasn't been set yet
		if options.EnvVars == nil {
			options.EnvVars = map[string]string{}
		}
		options.EnvVars["SSH_AUTH_SOCK"] = options.SshAgent.SocketFile()
	}
	return options, args
}

// RunCommand runs the IaC Provisioner with the given arguments and options and return stdout/stderr.
func RunCommand(additionalOptions *Options, additionalArgs ...string) (string, error) {
	options, args := GetCommonOptions(additionalOptions, additionalArgs...)

	cmd := generateCommand(options, args...)
	description := fmt.Sprintf("%s %v", options.Binary, args)

	return retry.DoWithRetryableErrors(description, options.RetryableTerraformErrors, options.MaxRetries, options.TimeBetweenRetries, func() (string, error) {
		s, err := shell.RunCommandAndGetOutput(cmd)
		if err != nil {
			return s, err
		}
		if err := hasWarning(additionalOptions, s); err != nil {
			return s, err
		}
		return s, err
	})

}

func defaultExecutable() string {
	cmd := exec.Command(TerraformDefaultPath, "-version")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err == nil {
		return TerraformDefaultPath
	}

	// fallback to Tofu if terraform is not available
	return TofuDefaultPath
}

func hasWarning(opts *Options, out string) error {
	for k, v := range opts.WarningsAsErrors {
		str := fmt.Sprintf("\n.*(?i:Warning): %s[^\n]*\n", k)
		re, err := regexp.Compile(str)
		if err != nil {
			return fmt.Errorf("cannot compile regex for warning detection: %w", err)
		}
		m := re.FindAllString(out, -1)
		if len(m) == 0 {
			continue
		}
		return fmt.Errorf("warning(s) were found: %s:\n%s", v, strings.Join(m, ""))
	}
	return nil
}
