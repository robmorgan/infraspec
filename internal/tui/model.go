package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/internal/tui/styles"
)

type state int

const (
	stateIdle state = iota
	stateRunning
	stateComplete
	stateError
)

type Model struct {
	state  state
	width  int
	height int

	// Components
	viewport  viewport.Model
	textInput textinput.Model
	spinner   spinner.Model

	// Test execution
	runner      *runner.Runner
	testRunner  *TestRunner
	featurePath string
	output      []string

	// UI state
	ready     bool
	showHelp  bool
	startTime time.Time
	duration  time.Duration

	// Styles
	styles styles.Styles
}

func NewModel(cfg *config.Config) Model {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Path to feature file..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	// Initialize viewport
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return Model{
		state:      stateIdle,
		viewport:   vp,
		textInput:  ti,
		spinner:    s,
		runner:     runner.New(cfg),
		testRunner: NewTestRunner(cfg),
		output:     []string{},
		styles:     styles.NewStyles(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}
