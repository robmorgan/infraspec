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

// RunCommandAndGetOutput runs the given command and returns the output as a string.
func RunCommandAndGetOutput(command Command) (string, error) {
	return "", nil
}

// RunCommand runs the given command and returns an error if the command fails.
func RunCommand(command Command) (*output, error) {
	config.Logging.Logger.Infof("Running command %s with args %s", command.Name, command.Args)

	cmd := exec.Command(command.Name, command.Args...)
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

	//go streamOutputToLogger(stdout, config.Logging.Logger.Info)
	//go streamOutputToLogger(stderr, config.Logging.Logger.Error)

	return output, cmd.Wait()
}

// This function captures stdout and stderr into the given variables while still printing it to the stdout and stderr
// of this Go program
func readStdoutAndStderr(stdout, stderr io.ReadCloser) (*output, error) {
	out := newOutput()
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	wg := &sync.WaitGroup{}

	wg.Add(2)
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

		// remove newline, our output is in a slice,
		// one element per line.
		line = strings.TrimSuffix(line, "\n")

		// only return early if the line does not have
		// any contents. We could have a line that does
		// not not have a newline before io.EOF, we still
		// need to add it to the output.
		if len(line) == 0 && readErr == io.EOF {
			break
		}

		// logger.Logger has a Logf method, but not a Log method.
		// We have to use the format string indirection to avoid
		// interpreting any possible formatting characters in
		// the line.
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

func streamOutputToLogger(reader io.ReadCloser, logFunc func(args ...interface{})) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logFunc(scanner.Text())
	}
}
