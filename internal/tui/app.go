package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/robmorgan/infraspec/internal/config"
)

func RunTUI(cfg *config.Config) error {
	// Create the TUI model
	model := NewModel(cfg)
	modelPtr := &model

	// Configure the tea program
	program := tea.NewProgram(
		modelPtr,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Handle any final cleanup if needed
	_ = finalModel

	return nil
}

func init() {
	// Ensure we can restore terminal state on exit
	if os.Getenv("DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
}
