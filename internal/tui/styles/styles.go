package styles

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Base    lipgloss.Style
	Header  lipgloss.Style
	Status  lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Muted   lipgloss.Style
	Input   lipgloss.Style
	Output  lipgloss.Style
	Help    lipgloss.Style
	Spinner lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Base: lipgloss.NewStyle().
			Padding(1, 2),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("238")).
			Padding(0, 1),
		Status: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")),
		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46")),
		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220")),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Input: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")),
		Output: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),
		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")),
	}
}
