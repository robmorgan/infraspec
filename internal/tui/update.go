package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "enter":
			if m.state == stateIdle && m.textInput.Value() != "" {
				m.featurePath = m.textInput.Value()
				m.state = stateRunning
				m.startTime = time.Now()
				m.output = []string{}
				m.textInput.SetValue("")
				return m, tea.Batch(
					m.testRunner.RunTest(m.featurePath),
					m.spinner.Tick,
				)
			}
		case "r":
			if m.state == stateComplete || m.state == stateError {
				m.state = stateIdle
				m.output = []string{}
				m.duration = 0
				m.textInput.Focus()
				return m, nil
			}
		case "esc":
			if m.state == stateRunning {
				m.testRunner.Cancel()
				m.state = stateIdle
				m.output = append(m.output, m.styles.Warning.Render("✗ Test execution canceled"))
				m.duration = time.Since(m.startTime)
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight-4)
			m.viewport.YPosition = headerHeight + 2
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMarginHeight - 4
		}

	case TestStreamStartMsg:
		return m, m.handleTestStream(msg.OutputChan, msg.DoneChan)

	case TestStreamMsg:
		m.output = append(m.output, m.formatTestOutput(TestOutputMsg(msg)))

	case TestStreamDoneMsg:
		m.duration = time.Since(m.startTime)
		if msg.Error != nil {
			m.state = stateError
			m.output = append(m.output, m.styles.Error.Render(fmt.Sprintf("✗ Test execution failed: %s", msg.Error)))
		} else {
			m.state = stateComplete
			m.output = append(m.output, m.styles.Success.Render("✓ Test execution completed successfully"))
		}

	case testCompleteMsg:
		m.state = stateComplete
		m.duration = time.Since(m.startTime)
		m.output = append(m.output, m.styles.Success.Render("✓ Test execution completed successfully"))

	case testErrorMsg:
		m.state = stateError
		m.duration = time.Since(m.startTime)
		m.output = append(m.output, m.styles.Error.Render(fmt.Sprintf("✗ Test execution failed: %s", msg.err)))

	case testOutputMsg:
		m.output = append(m.output, string(msg))
	}

	// Update components
	if m.state == stateIdle {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == stateRunning {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport content
	content := strings.Join(m.output, "\n")
	m.viewport.SetContent(content)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
