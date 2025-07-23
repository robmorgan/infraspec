package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robmorgan/infraspec/internal/tui/styles"
)

func RunDemo() {
	fmt.Printf(`
InfraSpec TUI Demo
==================

The Claude Code-inspired test runner interface provides:

🎯 Key Features:
  • Real-time test output streaming with timestamps
  • Interactive file selection with auto-completion
  • Keyboard shortcuts similar to Claude Code:
    - Enter: Run tests
    - Esc: Cancel running tests  
    - R: Re-run last test
    - ?: Toggle help
    - Q: Quit

🎨 UI Elements:
  • Header with status and timing information
  • Scrollable output viewport with syntax highlighting
  • Input field with validation
  • Footer with context-sensitive shortcuts
  • Spinner and progress indicators

⌨️  Keyboard Shortcuts:
  enter     Run the test with the specified feature file
  r         Run the test again (when complete)
  esc       Cancel running test
  ?         Toggle help screen
  q         Quit the application
  ctrl+c    Force quit

📁 Example Usage:
  1. Launch: infraspec ui
  2. Enter: features/aws/s3/s3_bucket.feature
  3. Press Enter to start test execution
  4. Watch real-time output stream
  5. Press 'r' to run again or 'q' to quit

🎭 Demo Mode Active: 
  This demo shows the interface design without requiring TTY access.
  In a real terminal, you'd see the full interactive TUI with:
  - Live updating progress bars
  - Real-time test output streaming  
  - Interactive keyboard navigation
  - Dynamic window resizing

To run the actual TUI: infraspec ui (requires interactive terminal)
`)
}

type DemoMsg string

func RunDemoTUI() tea.Model {
	model := Model{
		state:  stateIdle,
		styles: styles.NewStyles(),
		output: []string{
			"[15:04:05] • Starting InfraSpec test runner...",
			"[15:04:05] ✓ Configuration loaded successfully",
			"[15:04:06] • Running feature: features/aws/s3/s3_bucket.feature",
			"[15:04:06] ✓ Feature file validation passed",
			"[15:04:07] • Initializing AWS clients...",
			"[15:04:07] ✓ AWS authentication successful",
			"[15:04:08] • Executing scenario: S3 bucket creation",
			"[15:04:09] ✓ S3 bucket created successfully",
			"[15:04:10] • Running assertions...",
			"[15:04:11] ✓ Bucket encryption validation passed",
			"[15:04:12] ✓ All tests completed successfully",
		},
		duration: 8 * time.Second,
	}
	return &model
}
