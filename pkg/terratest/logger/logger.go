package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/testing"
)

var runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#24252E")).Background(lipgloss.Color("#F4F9A8"))

var Default = New()

func init() {
	// Create a new Logger that logs colorfully to stdout and is compatible with Terratest.
	logger.Default = New()
}

// New returns a new logger that logs colorfully to stdout and is compatible with Terratest.
func New() *logger.Logger {
	return logger.New(terratestLogger{})
}

type terratestLogger struct{}

func (_ terratestLogger) Logf(t testing.TestingT, format string, args ...interface{}) {
	DoLog(t, os.Stdout, fmt.Sprintf(format, args...))
}

// DoLog logs the given arguments to the given writer, prefixed by `RUNS“.
func DoLog(t testing.TestingT, writer io.Writer, args ...interface{}) {
	prefix := fmt.Sprintf("    %s ", runningStyle.Render(" RUNS "))
	allArgs := append([]interface{}{prefix}, args...)
	fmt.Fprintln(writer, allArgs...)
}
