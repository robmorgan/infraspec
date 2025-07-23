package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.headerView()
	body := m.bodyView()
	footer := m.footerView()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		footer,
	)
}

func (m Model) headerView() string {
	title := "InfraSpec Test Runner"

	var status string
	switch m.state {
	case stateIdle:
		status = m.styles.Muted.Render("Ready")
	case stateRunning:
		status = m.styles.Spinner.Render(m.spinner.View()) + " Running tests..."
	case stateComplete:
		status = m.styles.Success.Render("✓ Complete")
	case stateError:
		status = m.styles.Error.Render("✗ Failed")
	}

	if m.duration > 0 {
		status += m.styles.Muted.Render(fmt.Sprintf(" (%s)", m.duration.Round(time.Millisecond)))
	}

	titleStyle := m.styles.Header.Width(m.width - lipgloss.Width(status) - 4)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render(title),
		status,
	)
}

func (m Model) bodyView() string {
	if m.showHelp {
		return m.helpView()
	}

	var sections []string

	// Input section (only show when idle)
	if m.state == stateIdle {
		inputSection := lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.Status.Render("Enter feature file path:"),
			m.styles.Input.Render(m.textInput.View()),
		)
		sections = append(sections, inputSection)
	}

	// Output section
	if len(m.output) > 0 || m.state == stateRunning {
		outputTitle := "Test Output:"
		if m.featurePath != "" {
			outputTitle = fmt.Sprintf("Test Output: %s", m.featurePath)
		}

		outputSection := lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.Status.Render(outputTitle),
			m.styles.Output.Width(m.width-6).Height(m.viewport.Height).Render(m.viewport.View()),
		)
		sections = append(sections, outputSection)
	}

	if len(sections) == 0 {
		return m.styles.Muted.Render("Enter a feature file path to start testing...")
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) footerView() string {
	var shortcuts []string

	switch m.state {
	case stateIdle:
		shortcuts = []string{
			"enter: run test",
			"?: help",
			"q: quit",
		}
	case stateRunning:
		shortcuts = []string{
			"esc: cancel",
			"q: quit",
		}
	case stateComplete, stateError:
		shortcuts = []string{
			"r: run again",
			"?: help",
			"q: quit",
		}
	}

	help := strings.Join(shortcuts, " • ")
	return m.styles.Help.Width(m.width).Render(help)
}

func (m Model) helpView() string {
	help := `
InfraSpec Test Runner - Claude Code Inspired Interface

KEYBOARD SHORTCUTS:
  enter     Run the test with the specified feature file
  r         Run the test again (when complete)
  esc       Cancel running test
  ?         Toggle this help screen
  q         Quit the application
  ctrl+c    Force quit

USAGE:
  1. Enter the path to your feature file
  2. Press Enter to start the test execution
  3. Watch real-time output as tests run
  4. Press 'r' to run again or 'q' to quit

EXAMPLES:
  features/aws/s3/s3_bucket.feature
  features/terraform/hello_world.feature
  examples/aws/dynamodb/table.feature

Press ? again to return to the main view.
`

	return m.styles.Output.Width(m.width - 6).Render(strings.TrimSpace(help))
}

type (
	testCompleteMsg struct{}
	testErrorMsg    struct{ err error }
	testOutputMsg   string
)
