package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/tui"
)

var demoMode bool

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch interactive test runner interface",
	Long: `Launch an interactive TUI (Terminal User Interface) for running InfraSpec tests.
	
The UI provides a Claude Code-inspired interface with:
- Real-time test output streaming
- Interactive file selection
- Keyboard shortcuts for navigation
- Progress indicators and status updates

Example usage:
  infraspec ui
  infraspec ui --demo    # Show demo without requiring TTY
`,
	Run: func(cmd *cobra.Command, args []string) {
		if demoMode {
			tui.RunDemo()
			return
		}

		// Check if we're in an interactive terminal
		if !isTerminalInteractive() {
			fmt.Println("Error: TUI requires an interactive terminal.")
			fmt.Println("Run 'infraspec ui --demo' to see the interface design.")
			return
		}

		cfg, err := config.DefaultConfig()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			return
		}

		if verbose {
			cfg.Verbose = true
		}

		if err := tui.RunTUI(cfg); err != nil {
			fmt.Printf("Failed to run TUI: %v\n", err)
		}
	},
}

func isTerminalInteractive() bool {
	// Check if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	// Check if stdin is a terminal
	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

func init() {
	RootCmd.AddCommand(uiCmd)
	uiCmd.Flags().BoolVar(&demoMode, "demo", false, "Show TUI demo without requiring interactive terminal")
}
