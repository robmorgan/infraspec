package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/robmorgan/infraspec/internal/config"
)

// Command is a command to be run in a shell.
type Command struct {
	Name       string
	Args       []string
	WorkingDir string
	Env        map[string]string
}

// ErrWithCmdOutput is an error that includes the output of the command.
type ErrWithCmdOutput struct {
	Underlying error
	Output     *output
}

func (e *ErrWithCmdOutput) Error() string {
	return fmt.Sprintf("error while running command: %v; %s", e.Underlying, e.Output.Stderr())
}

// Add this near the top of the file
var allowedCommands = map[string]bool{
	"terraform": true,
	"aws":       true,
	"kubectl":   true,
	"docker":    true,
	"git":       true,
}

// validateCommand checks if the command is in the list of allowed commands
func validateCommand(command Command) error {
	if !allowedCommands[command.Name] {
		return fmt.Errorf("command not allowed: %s", command.Name)
	}

	return nil
}

// RunCommandAndGetOutput runs the given command and returns the output as a string or an error if the command fails.
func RunCommandAndGetOutput(command Command) (string, error) {
	output, err := runCommand(command)
	if err != nil {
		return output.Stdout(), &ErrWithCmdOutput{err, output}
	}

	return output.Stdout(), nil
}

// runCommand runs the given command and returns an error if the command fails.
func runCommand(command Command) (*output, error) {
	config.Logging.Logger.Infof("Running command %s with args %s", command.Name, command.Args)

	err := validateCommand(command)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(command.Name, command.Args...) //nolint:gosec
	cmd.Dir = command.WorkingDir
	cmd.Stdin = os.Stdin
	cmd.Env = formatEnvVars(command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	output, err := readStdoutAndStderr(stdout, stderr)
	if err != nil {
		return output, err
	}

	return output, cmd.Wait()
}

// This function captures stdout and stderr into the given variables while still printing it to the stdout and stderr
// of this Go program
func readStdoutAndStderr(stdout, stderr io.ReadCloser) (*output, error) {
	out := newOutput()
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	wg := &sync.WaitGroup{}
	wgSize := 2 // stdout and stderr

	wg.Add(wgSize)
	var stdoutErr, stderrErr error
	go func() {
		defer wg.Done()
		stdoutErr = readData(stdoutReader, out.stdout)
	}()
	go func() {
		defer wg.Done()
		stderrErr = readData(stderrReader, out.stderr)
	}()
	wg.Wait()

	if stdoutErr != nil {
		return out, stdoutErr
	}
	if stderrErr != nil {
		return out, stderrErr
	}

	return out, nil
}

func readData(reader *bufio.Reader, writer io.StringWriter) error {
	var line string
	var readErr error
	for {
		line, readErr = reader.ReadString('\n')

		// remove newline, our output is in a slice, one element per line.
		line = strings.TrimSuffix(line, "\n")

		// only return early if the line does not have any contents. We could have a line that does not have a newline
		// before io.EOF, we still need to add it to the output.
		if line == "" && readErr == io.EOF {
			break
		}

		// logger.Logger has a Logf method, but not a Log method. We have to use the format string indirection to avoid
		// interpreting any possible formatting characters in the line.
		//
		// See https://github.com/gruntwork-io/terratest/issues/982.
		config.Logging.Logger.Infof("%s", line)

		if _, err := writer.WriteString(line); err != nil {
			return err
		}

		if readErr != nil {
			break
		}
	}
	if readErr != io.EOF {
		return readErr
	}
	return nil
}

func formatEnvVars(command Command) []string {
	env := os.Environ()
	for key, value := range command.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}
