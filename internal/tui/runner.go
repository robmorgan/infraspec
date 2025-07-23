package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/robmorgan/infraspec/internal/config"
)

type TestRunner struct {
	cfg    *config.Config
	cancel context.CancelFunc
}

func NewTestRunner(cfg *config.Config) *TestRunner {
	return &TestRunner{
		cfg: cfg,
	}
}

type TestOutputMsg struct {
	Line      string
	Timestamp time.Time
	Type      OutputType
}

type OutputType int

const (
	OutputNormal OutputType = iota
	OutputError
	OutputSuccess
	OutputWarning
	OutputDebug
)

func (t *TestRunner) RunTest(featurePath string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		t.cancel = cancel

		// Create command to run infraspec
		cmd := exec.CommandContext(ctx, os.Args[0], featurePath)
		if t.cfg.Verbose {
			cmd.Args = append(cmd.Args, "-v")
		}

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return testErrorMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return testErrorMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return testErrorMsg{err: fmt.Errorf("failed to start test: %w", err)}
		}

		// Create channels for output
		outputChan := make(chan TestOutputMsg, 100)
		doneChan := make(chan error, 1)

		// Start goroutines to read output
		go t.readOutput(stdout, outputChan, OutputNormal)
		go t.readOutput(stderr, outputChan, OutputError)

		// Wait for command completion
		go func() {
			doneChan <- cmd.Wait()
		}()

		// Return the first message to start the stream
		return TestStreamStartMsg{
			OutputChan: outputChan,
			DoneChan:   doneChan,
		}
	}
}

func (t *TestRunner) readOutput(r io.Reader, outputChan chan<- TestOutputMsg, outputType OutputType) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		outputChan <- TestOutputMsg{
			Line:      line,
			Timestamp: time.Now(),
			Type:      outputType,
		}
	}
}

func (t *TestRunner) Cancel() {
	if t.cancel != nil {
		t.cancel()
	}
}

type TestStreamStartMsg struct {
	OutputChan <-chan TestOutputMsg
	DoneChan   <-chan error
}

type (
	TestStreamMsg     TestOutputMsg
	TestStreamDoneMsg struct {
		Error error
	}
)

func ListenForTestOutput(outputChan <-chan TestOutputMsg, doneChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case output := <-outputChan:
			return TestStreamMsg(output)
		case err := <-doneChan:
			return TestStreamDoneMsg{Error: err}
		}
	}
}

func (m *Model) handleTestStream(outputChan <-chan TestOutputMsg, doneChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case output := <-outputChan:
			return TestStreamMsg(output)
		case err := <-doneChan:
			return TestStreamDoneMsg{Error: err}
		default:
			// Continue listening
			time.Sleep(50 * time.Millisecond)
			return m.handleTestStream(outputChan, doneChan)()
		}
	}
}

func (m *Model) formatTestOutput(msg TestOutputMsg) string {
	timestamp := msg.Timestamp.Format("15:04:05")

	var style lipgloss.Style
	var prefix string

	switch msg.Type {
	case OutputError:
		style = m.styles.Error
		prefix = "✗"
	case OutputSuccess:
		style = m.styles.Success
		prefix = "✓"
	case OutputWarning:
		style = m.styles.Warning
		prefix = "⚠"
	case OutputDebug:
		style = m.styles.Muted
		prefix = "•"
	case OutputNormal:
		style = lipgloss.NewStyle()
		prefix = "•"
	default:
		style = lipgloss.NewStyle()
		prefix = "•"
	}

	timeStyle := m.styles.Muted.Render(fmt.Sprintf("[%s]", timestamp))
	contentStyle := style.Render(fmt.Sprintf("%s %s", prefix, msg.Line))

	return fmt.Sprintf("%s %s", timeStyle, contentStyle)
}
